package entity

import (
	"fmt"
	"strings"
	"time"

	exceptionentity "github.com/tuannm99/podzone/internal/backoffice/domain/exception/entity"
	fulfillmententity "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/entity"
	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order/entity"
	settlemententity "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/entity"
)

func (o *RoutedOrder) Advance(actor string, now time.Time) error {
	customerOrder, err := o.customerOrderAggregate()
	if err != nil {
		return err
	}
	change, err := customerOrder.Advance(now)
	if err != nil {
		return err
	}
	o.applyCustomerOrderSnapshot(customerOrder.Snapshot())
	o.recordOrderChange(actor, change, now)
	return nil
}

func (o *RoutedOrder) OpenException(exceptionType string, actor string, now time.Time) error {
	exceptionAggregate, err := o.exceptionAggregate()
	if err != nil {
		return err
	}
	change, err := exceptionAggregate.Open(exceptionType, now)
	if err != nil {
		return err
	}
	o.applyExceptionSnapshot(exceptionAggregate.Snapshot())
	o.recordExceptionChange(actor, change, now)
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) UpdateExceptionStatus(status string, actor string, now time.Time) error {
	exceptionAggregate, err := o.exceptionAggregate()
	if err != nil {
		return err
	}
	change, err := exceptionAggregate.UpdateStatus(status, now)
	if err != nil {
		return err
	}
	o.applyExceptionSnapshot(exceptionAggregate.Snapshot())
	o.recordExceptionChange(actor, change, now)
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) UpdateShipment(
	shipmentStatus string,
	carrier string,
	trackingNumber string,
	trackingURL string,
	notes string,
	actor string,
	now time.Time,
) error {
	fulfillmentOrder, err := o.fulfillmentOrderAggregate()
	if err != nil {
		return err
	}
	systemChange, noteChange, markOrderShipped, err := fulfillmentOrder.UpdateShipment(
		fulfillmententity.ShipmentUpdate{
			Status:         shipmentStatus,
			Carrier:        carrier,
			TrackingNumber: trackingNumber,
			TrackingURL:    trackingURL,
			Notes:          notes,
		},
		now,
	)
	if err != nil {
		return err
	}
	if markOrderShipped {
		o.markOrderShipped(actor, now)
	}
	o.applyFulfillmentSnapshot(fulfillmentOrder.Snapshot())
	o.recordFulfillmentChange(actor, systemChange, now)
	if noteChange != nil {
		o.recordActivity(
			RoutedOrderActivityTypeShipmentNote,
			actor,
			noteChange.Message,
			now,
			fulfillmentDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) UpdateQueueControl(
	assignee string,
	shipmentSlaDueAt *time.Time,
	issueSlaDueAt *time.Time,
	actor string,
	now time.Time,
) {
	customerOrder, err := o.customerOrderAggregate()
	if err != nil {
		return
	}
	change := customerOrder.UpdateQueueControl(assignee, shipmentSlaDueAt, issueSlaDueAt, now)
	o.applyCustomerOrderSnapshot(customerOrder.Snapshot())
	o.recordOrderChange(actor, change, now)
}

func (o *RoutedOrder) ApplyBulkQueueControl(
	assignee *string,
	shipmentSlaDueAt *time.Time,
	settlementStatus *string,
	actor string,
	now time.Time,
) error {
	customerOrder, err := o.customerOrderAggregate()
	if err != nil {
		return err
	}
	change, err := customerOrder.ApplyBulkQueueControl(assignee, shipmentSlaDueAt, settlementStatus, now)
	if err != nil {
		return err
	}
	o.applyCustomerOrderSnapshot(customerOrder.Snapshot())
	o.recordOrderChange(actor, change, now)
	return nil
}

func (o *RoutedOrder) UpdateSettlement(
	fulfillmentCost string,
	shippingCost string,
	settlementStatus string,
	notes string,
	actor string,
	now time.Time,
) error {
	settlementRecord, err := o.settlementRecordAggregate()
	if err != nil {
		return err
	}
	systemChange, noteChange, err := settlementRecord.UpdateSettlement(
		fulfillmentCost,
		shippingCost,
		settlementStatus,
		notes,
		now,
	)
	if err != nil {
		return err
	}
	o.applySettlementSnapshot(settlementRecord.Snapshot())
	o.recordSettlementChange(actor, systemChange, now)
	if noteChange != nil {
		o.recordActivity(
			RoutedOrderActivityTypeSettlementNote,
			actor,
			noteChange.Message,
			now,
			settlementDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) UpdateIssueHandling(
	issueCost string,
	issueResolution string,
	notes string,
	actor string,
	now time.Time,
) error {
	settlementRecord, err := o.settlementRecordAggregate()
	if err != nil {
		return err
	}
	systemChange, noteChange, err := settlementRecord.UpdateIssueHandling(issueCost, issueResolution, notes, now)
	if err != nil {
		return err
	}
	o.applySettlementSnapshot(settlementRecord.Snapshot())
	o.recordSettlementChange(actor, systemChange, now)
	if noteChange != nil {
		o.recordActivity(
			RoutedOrderActivityTypeIssueNote,
			actor,
			noteChange.Message,
			now,
			settlementDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) ApplyManualReroute(
	partner string,
	fulfillmentCost string,
	shippingCost string,
	estimatedUnitMargin string,
	routingSummary string,
	actor string,
	now time.Time,
) error {
	partner = strings.TrimSpace(partner)
	if o.Status != RoutedOrderStatusRoutingBlocked {
		return fmt.Errorf("routed order is not in routing_blocked status")
	}
	if partner == "" {
		return fmt.Errorf("selected routing partner is required")
	}

	previousPartner := o.Partner
	previousBlockCode := o.RoutingBlockCode
	previousBlockReason := o.RoutingBlockReason
	o.Status = RoutedOrderStatusQueued
	o.Partner = partner
	o.RoutingBlockCode = ""
	o.RoutingBlockReason = ""
	o.FulfillmentCost = MultiplyMoney(fulfillmentCost, o.Quantity)
	o.ShippingCost = shippingCost
	o.RealizedMargin = CalculateMargin(o.Total, o.FulfillmentCost, o.ShippingCost, o.IssueCost)
	entry := fmt.Sprintf("Routing unblocked: manually rerouted to %s", partner)
	o.recordSystemTimelineActivity(
		actor,
		entry,
		now,
		activityDetails(
			"status", RoutedOrderStatusQueued,
			"previous_partner", previousPartner,
			"partner", partner,
			"previous_routing_block_code", previousBlockCode,
			"previous_routing_block_reason", previousBlockReason,
			"routing_summary", routingSummary,
			"estimated_unit_margin", estimatedUnitMargin,
			"manual_reroute", "true",
		),
	)
	o.UpdatedAt = now
	return nil
}

func (o *RoutedOrder) markOrderShipped(actor string, now time.Time) {
	customerOrder, err := o.customerOrderAggregate()
	if err != nil {
		return
	}
	change, changed := customerOrder.MarkShipped(now)
	if !changed {
		return
	}
	o.applyCustomerOrderSnapshot(customerOrder.Snapshot())
	o.recordOrderChange(actor, change, now)
}

func (o *RoutedOrder) recordActivity(
	activityType string,
	actor string,
	message string,
	createdAt time.Time,
	details []RoutedOrderActivityDetail,
) {
	message = strings.TrimSpace(message)
	o.ActivityLog = append(o.ActivityLog, RoutedOrderActivity{
		Type:      strings.TrimSpace(activityType),
		Actor:     strings.TrimSpace(actor),
		Message:   message,
		Details:   details,
		CreatedAt: createdAt,
	})
}

func (o *RoutedOrder) customerOrderAggregate() (*orderentity.CustomerOrder, error) {
	return orderentity.RehydrateCustomerOrder(orderentity.CustomerOrderSnapshot{
		ID:                 o.ID,
		StoreID:            o.StoreID,
		Status:             o.Status,
		Partner:            o.Partner,
		OperatorAssignee:   o.OperatorAssignee,
		ShipmentSlaDueAt:   o.ShipmentSlaDueAt,
		IssueSlaDueAt:      o.IssueSlaDueAt,
		ExceptionStatus:    o.ExceptionStatus,
		RoutingBlockCode:   o.RoutingBlockCode,
		RoutingBlockReason: o.RoutingBlockReason,
		SettlementStatus:   o.SettlementStatus,
		UpdatedAt:          o.UpdatedAt,
	})
}

func (o *RoutedOrder) applyCustomerOrderSnapshot(snapshot orderentity.CustomerOrderSnapshot) {
	o.Status = snapshot.Status
	o.Partner = snapshot.Partner
	o.OperatorAssignee = snapshot.OperatorAssignee
	o.ShipmentSlaDueAt = snapshot.ShipmentSlaDueAt
	o.IssueSlaDueAt = snapshot.IssueSlaDueAt
	o.RoutingBlockCode = snapshot.RoutingBlockCode
	o.RoutingBlockReason = snapshot.RoutingBlockReason
	o.SettlementStatus = snapshot.SettlementStatus
	o.UpdatedAt = snapshot.UpdatedAt
}

func (o *RoutedOrder) exceptionAggregate() (*exceptionentity.OrderException, error) {
	return exceptionentity.RehydrateOrderException(exceptionentity.OrderExceptionSnapshot{
		OrderID: o.ID,
		Type:    o.ExceptionType,
		Status:  o.ExceptionStatus,
	})
}

func (o *RoutedOrder) applyExceptionSnapshot(snapshot exceptionentity.OrderExceptionSnapshot) {
	o.ExceptionType = snapshot.Type
	o.ExceptionStatus = snapshot.Status
}

func (o *RoutedOrder) fulfillmentOrderAggregate() (*fulfillmententity.FulfillmentOrder, error) {
	return fulfillmententity.RehydrateFulfillmentOrder(fulfillmententity.FulfillmentOrderSnapshot{
		OrderID:        o.ID,
		Partner:        o.Partner,
		Status:         o.ShipmentStatus,
		Carrier:        o.ShipmentCarrier,
		TrackingNumber: o.ShipmentTrackingNumber,
		TrackingURL:    o.ShipmentTrackingURL,
		Notes:          o.ShipmentNotes,
		ShippedAt:      o.ShippedAt,
		DeliveredAt:    o.DeliveredAt,
	})
}

func (o *RoutedOrder) applyFulfillmentSnapshot(snapshot fulfillmententity.FulfillmentOrderSnapshot) {
	o.ShipmentStatus = snapshot.Status
	o.ShipmentCarrier = snapshot.Carrier
	o.ShipmentTrackingNumber = snapshot.TrackingNumber
	o.ShipmentTrackingURL = snapshot.TrackingURL
	o.ShipmentNotes = snapshot.Notes
	o.ShippedAt = snapshot.ShippedAt
	o.DeliveredAt = snapshot.DeliveredAt
}

func (o *RoutedOrder) settlementRecordAggregate() (*settlemententity.SettlementRecord, error) {
	return settlemententity.RehydrateSettlementRecord(settlemententity.SettlementRecordSnapshot{
		OrderID:         o.ID,
		Total:           o.Total,
		FulfillmentCost: o.FulfillmentCost,
		ShippingCost:    o.ShippingCost,
		IssueCost:       o.IssueCost,
		IssueResolution: o.IssueResolution,
		IssueNotes:      o.IssueNotes,
		RealizedMargin:  o.RealizedMargin,
		Status:          o.SettlementStatus,
		Notes:           o.SettlementNotes,
		ExceptionType:   o.ExceptionType,
		ShipmentStatus:  o.ShipmentStatus,
	})
}

func (o *RoutedOrder) applySettlementSnapshot(snapshot settlemententity.SettlementRecordSnapshot) {
	o.FulfillmentCost = snapshot.FulfillmentCost
	o.ShippingCost = snapshot.ShippingCost
	o.IssueCost = snapshot.IssueCost
	o.IssueResolution = snapshot.IssueResolution
	o.IssueNotes = snapshot.IssueNotes
	o.RealizedMargin = snapshot.RealizedMargin
	o.SettlementStatus = snapshot.Status
	o.SettlementNotes = snapshot.Notes
}

func (o *RoutedOrder) recordOrderChange(actor string, change orderentity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	o.recordSystemTimelineActivity(actor, change.Message, now, orderDetails(change.Details))
}

func (o *RoutedOrder) recordExceptionChange(actor string, change exceptionentity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	o.recordSystemTimelineActivity(actor, change.Message, now, exceptionDetails(change.Details))
}

func (o *RoutedOrder) recordFulfillmentChange(actor string, change fulfillmententity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	o.recordSystemTimelineActivity(actor, change.Message, now, fulfillmentDetails(change.Details))
}

func (o *RoutedOrder) recordSettlementChange(actor string, change settlemententity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	o.recordSystemTimelineActivity(actor, change.Message, now, settlementDetails(change.Details))
}

func (o *RoutedOrder) recordSystemTimelineActivity(
	actor string,
	message string,
	createdAt time.Time,
	details []RoutedOrderActivityDetail,
) {
	message = strings.TrimSpace(message)
	o.Timeline = append(o.Timeline, message)
	o.recordActivity(RoutedOrderActivityTypeSystem, actor, message, createdAt, details)
}

func orderDetails(details []orderentity.ActivityDetail) []RoutedOrderActivityDetail {
	out := make([]RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func exceptionDetails(details []exceptionentity.ActivityDetail) []RoutedOrderActivityDetail {
	out := make([]RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func fulfillmentDetails(details []fulfillmententity.ActivityDetail) []RoutedOrderActivityDetail {
	out := make([]RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func settlementDetails(details []settlemententity.ActivityDetail) []RoutedOrderActivityDetail {
	out := make([]RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func activityDetails(pairs ...string) []RoutedOrderActivityDetail {
	details := make([]RoutedOrderActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		details = append(details, RoutedOrderActivityDetail{Key: key, Value: value})
	}
	return details
}
