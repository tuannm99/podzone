package fulfillment

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/shared"
)

const (
	StatusAwaitingLabel = "awaiting_label"
	StatusLabelReady    = "label_ready"
	StatusInTransit     = "in_transit"
	StatusDelivered     = "delivered"
	StatusDeliveryIssue = "delivery_issue"
)

type ActivityDetail struct {
	Key   string
	Value string
}

type Change struct {
	Message string
	Details []ActivityDetail
}

type FulfillmentOrderSnapshot struct {
	OrderID        string
	Partner        string
	Status         string
	Carrier        string
	TrackingNumber string
	TrackingURL    string
	Notes          string
	ShippedAt      *time.Time
	DeliveredAt    *time.Time
}

type ShipmentUpdate struct {
	Status         string
	Carrier        string
	TrackingNumber string
	TrackingURL    string
	Notes          string
}

type FulfillmentOrder struct {
	orderID        string
	partner        string
	status         string
	carrier        string
	trackingNumber string
	trackingURL    string
	notes          string
	shippedAt      *time.Time
	deliveredAt    *time.Time
	pendingEvents  []DomainEvent
}

var _ shared.AggregateRoot = (*FulfillmentOrder)(nil)

func RehydrateFulfillmentOrder(snapshot FulfillmentOrderSnapshot) (*FulfillmentOrder, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("fulfillment order id is required")
	}
	status := normalizeStatus(snapshot.Status)
	if strings.TrimSpace(snapshot.Status) != "" && status == "" {
		return nil, fmt.Errorf("invalid shipment status")
	}
	return &FulfillmentOrder{
		orderID:        snapshot.OrderID,
		partner:        snapshot.Partner,
		status:         status,
		carrier:        snapshot.Carrier,
		trackingNumber: snapshot.TrackingNumber,
		trackingURL:    snapshot.TrackingURL,
		notes:          snapshot.Notes,
		shippedAt:      snapshot.ShippedAt,
		deliveredAt:    snapshot.DeliveredAt,
	}, nil
}

func (o *FulfillmentOrder) AggregateID() string {
	return o.orderID
}

func (o *FulfillmentOrder) Snapshot() FulfillmentOrderSnapshot {
	return FulfillmentOrderSnapshot{
		OrderID:        o.orderID,
		Partner:        o.partner,
		Status:         o.status,
		Carrier:        o.carrier,
		TrackingNumber: o.trackingNumber,
		TrackingURL:    o.trackingURL,
		Notes:          o.notes,
		ShippedAt:      o.shippedAt,
		DeliveredAt:    o.deliveredAt,
	}
}

func (o *FulfillmentOrder) PullEvents() []DomainEvent {
	events := o.pendingEvents
	o.pendingEvents = nil
	return events
}

func (o *FulfillmentOrder) UpdateShipment(update ShipmentUpdate, now time.Time) (Change, *Change, bool, error) {
	status := normalizeStatus(update.Status)
	if status == "" {
		return Change{}, nil, false, fmt.Errorf("invalid shipment status")
	}
	if requiresAssignedPartner(status) && strings.TrimSpace(o.partner) == "" {
		return Change{}, nil, false, fmt.Errorf("fulfillment partner is required before shipping")
	}
	carrier := strings.TrimSpace(update.Carrier)
	if carrier == "" {
		carrier = strings.TrimSpace(o.carrier)
	}
	if requiresCarrier(status) && carrier == "" {
		return Change{}, nil, false, fmt.Errorf("shipment carrier is required")
	}

	o.status = status
	o.carrier = carrier
	o.trackingNumber = fallbackTrimmed(update.TrackingNumber, o.trackingNumber)
	o.trackingURL = fallbackTrimmed(update.TrackingURL, o.trackingURL)
	o.notes = strings.TrimSpace(update.Notes)

	markOrderShipped := false
	switch status {
	case StatusInTransit:
		markOrderShipped = true
		if o.shippedAt == nil {
			o.shippedAt = &now
		}
		o.deliveredAt = nil
		o.record(ShipmentInTransit{
			OrderID:        o.orderID,
			Partner:        o.partner,
			Carrier:        o.carrier,
			TrackingNumber: o.trackingNumber,
			OccurredAt:     now.UTC(),
		})
	case StatusDelivered:
		markOrderShipped = true
		if o.shippedAt == nil {
			o.shippedAt = &now
		}
		o.deliveredAt = &now
		o.record(ShipmentDelivered{
			OrderID:        o.orderID,
			Partner:        o.partner,
			Carrier:        o.carrier,
			TrackingNumber: o.trackingNumber,
			OccurredAt:     now.UTC(),
		})
	case StatusAwaitingLabel, StatusLabelReady, StatusDeliveryIssue:
		o.deliveredAt = nil
		o.record(ShipmentStatusUpdated{
			OrderID:    o.orderID,
			Partner:    o.partner,
			Status:     o.status,
			OccurredAt: now.UTC(),
		})
	}

	systemChange := Change{
		Message: shipmentTimelineEntry(o),
		Details: details(
			"shipment_status", o.status,
			"carrier", fallbackCarrier(o.carrier),
			"tracking_number", o.trackingNumber,
		),
	}

	var noteChange *Change
	if o.notes != "" {
		noteChange = &Change{
			Message: o.notes,
			Details: details(
				"shipment_status", o.status,
				"carrier", fallbackCarrier(o.carrier),
			),
		}
	}
	return systemChange, noteChange, markOrderShipped, nil
}

func (o *FulfillmentOrder) record(event DomainEvent) {
	o.pendingEvents = append(o.pendingEvents, event)
}

func requiresAssignedPartner(status string) bool {
	switch status {
	case StatusLabelReady, StatusInTransit, StatusDelivered:
		return true
	default:
		return false
	}
}

func requiresCarrier(status string) bool {
	switch status {
	case StatusInTransit, StatusDelivered:
		return true
	default:
		return false
	}
}

func normalizeStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case StatusAwaitingLabel, StatusLabelReady, StatusInTransit, StatusDelivered, StatusDeliveryIssue:
		return raw
	default:
		return ""
	}
}

func shipmentTimelineEntry(order *FulfillmentOrder) string {
	switch order.status {
	case StatusLabelReady:
		return fmt.Sprintf("Shipment prepared with %s", fallbackCarrier(order.carrier))
	case StatusInTransit:
		return fmt.Sprintf(
			"Shipment is in transit via %s%s",
			fallbackCarrier(order.carrier),
			fallbackTrackingSuffix(order.trackingNumber),
		)
	case StatusDelivered:
		return "Shipment marked delivered by store operator"
	case StatusDeliveryIssue:
		return "Shipment issue flagged for manual follow-up"
	default:
		return "Shipment is awaiting label assignment"
	}
}

func fallbackCarrier(carrier string) string {
	carrier = strings.TrimSpace(carrier)
	if carrier == "" {
		return "manual carrier"
	}
	return carrier
}

func fallbackTrackingSuffix(trackingNumber string) string {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", trackingNumber)
}

func fallbackTrimmed(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func details(pairs ...string) []ActivityDetail {
	out := make([]ActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		out = append(out, ActivityDetail{Key: key, Value: value})
	}
	return out
}
