package operations

import (
	"strings"
	"time"

	exceptionentity "github.com/tuannm99/podzone/internal/backoffice/domain/exception"
	fulfillmententity "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment"
	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	settlemententity "github.com/tuannm99/podzone/internal/backoffice/domain/settlement"
)

func advanceRoutedOrder(o *routingctx.RoutedOrder, actor string, now time.Time) error {
	customerOrder, err := customerOrderAggregate(o)
	if err != nil {
		return err
	}
	change, err := customerOrder.Advance(now)
	if err != nil {
		return err
	}
	applyCustomerOrderSnapshot(o, customerOrder.Snapshot())
	recordOrderChange(o, actor, change, now)
	return nil
}

func openOrderException(o *routingctx.RoutedOrder, exceptionType string, actor string, now time.Time) error {
	exceptionAggregate, err := exceptionAggregate(o)
	if err != nil {
		return err
	}
	change, err := exceptionAggregate.Open(exceptionType, now)
	if err != nil {
		return err
	}
	applyExceptionSnapshot(o, exceptionAggregate.Snapshot())
	recordExceptionChange(o, actor, change, now)
	o.UpdatedAt = now
	return nil
}

func updateOrderExceptionStatus(o *routingctx.RoutedOrder, status string, actor string, now time.Time) error {
	exceptionAggregate, err := exceptionAggregate(o)
	if err != nil {
		return err
	}
	change, err := exceptionAggregate.UpdateStatus(status, now)
	if err != nil {
		return err
	}
	applyExceptionSnapshot(o, exceptionAggregate.Snapshot())
	recordExceptionChange(o, actor, change, now)
	o.UpdatedAt = now
	return nil
}

func updateOrderShipment(
	o *routingctx.RoutedOrder,
	shipmentStatus string,
	carrier string,
	trackingNumber string,
	trackingURL string,
	notes string,
	actor string,
	now time.Time,
) error {
	fulfillmentOrder, err := fulfillmentOrderAggregate(o)
	if err != nil {
		return err
	}
	systemChange, noteChange, shouldMarkOrderShipped, err := fulfillmentOrder.UpdateShipment(
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
	if shouldMarkOrderShipped {
		markOrderShipped(o, actor, now)
	}
	applyFulfillmentSnapshot(o, fulfillmentOrder.Snapshot())
	recordFulfillmentChange(o, actor, systemChange, now)
	if noteChange != nil {
		recordActivity(
			o,
			routingctx.RoutedOrderActivityTypeShipmentNote,
			actor,
			noteChange.Message,
			now,
			fulfillmentDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func updateOrderQueueControl(
	o *routingctx.RoutedOrder,
	assignee string,
	shipmentSlaDueAt *time.Time,
	issueSlaDueAt *time.Time,
	actor string,
	now time.Time,
) {
	customerOrder, err := customerOrderAggregate(o)
	if err != nil {
		return
	}
	change := customerOrder.UpdateQueueControl(assignee, shipmentSlaDueAt, issueSlaDueAt, now)
	applyCustomerOrderSnapshot(o, customerOrder.Snapshot())
	recordOrderChange(o, actor, change, now)
}

func applyBulkQueueControl(
	o *routingctx.RoutedOrder,
	assignee *string,
	shipmentSlaDueAt *time.Time,
	settlementStatus *string,
	actor string,
	now time.Time,
) error {
	customerOrder, err := customerOrderAggregate(o)
	if err != nil {
		return err
	}
	change, err := customerOrder.ApplyBulkQueueControl(assignee, shipmentSlaDueAt, settlementStatus, now)
	if err != nil {
		return err
	}
	applyCustomerOrderSnapshot(o, customerOrder.Snapshot())
	recordOrderChange(o, actor, change, now)
	return nil
}

func updateOrderSettlement(
	o *routingctx.RoutedOrder,
	fulfillmentCost string,
	shippingCost string,
	settlementStatus string,
	notes string,
	actor string,
	now time.Time,
) error {
	settlementRecord, err := settlementRecordAggregate(o)
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
	applySettlementSnapshot(o, settlementRecord.Snapshot())
	recordSettlementChange(o, actor, systemChange, now)
	if noteChange != nil {
		recordActivity(
			o,
			routingctx.RoutedOrderActivityTypeSettlementNote,
			actor,
			noteChange.Message,
			now,
			settlementDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func updateOrderIssueHandling(
	o *routingctx.RoutedOrder,
	issueCost string,
	issueResolution string,
	notes string,
	actor string,
	now time.Time,
) error {
	settlementRecord, err := settlementRecordAggregate(o)
	if err != nil {
		return err
	}
	systemChange, noteChange, err := settlementRecord.UpdateIssueHandling(issueCost, issueResolution, notes, now)
	if err != nil {
		return err
	}
	applySettlementSnapshot(o, settlementRecord.Snapshot())
	recordSettlementChange(o, actor, systemChange, now)
	if noteChange != nil {
		recordActivity(
			o,
			routingctx.RoutedOrderActivityTypeIssueNote,
			actor,
			noteChange.Message,
			now,
			settlementDetails(noteChange.Details),
		)
	}
	o.UpdatedAt = now
	return nil
}

func applyManualReroute(
	o *routingctx.RoutedOrder,
	partner string,
	fulfillmentCost string,
	shippingCost string,
	estimatedUnitMargin string,
	routingSummary string,
	actor string,
	now time.Time,
) error {
	partner = strings.TrimSpace(partner)
	customerOrder, err := customerOrderAggregate(o)
	if err != nil {
		return err
	}
	change, err := customerOrder.RouteManually(partner, now)
	if err != nil {
		return err
	}
	applyCustomerOrderSnapshot(o, customerOrder.Snapshot())
	o.FulfillmentCost = routingctx.MultiplyMoney(fulfillmentCost, o.Quantity)
	o.ShippingCost = shippingCost
	o.RealizedMargin = routingctx.CalculateMargin(o.Total, o.FulfillmentCost, o.ShippingCost, o.IssueCost)
	details := orderDetails(change.Details)
	details = append(
		details,
		routingctx.RoutedOrderActivityDetail{Key: "routing_summary", Value: routingSummary},
		routingctx.RoutedOrderActivityDetail{Key: "estimated_unit_margin", Value: estimatedUnitMargin},
	)
	recordSystemTimelineActivity(o, actor, change.Message, now, details)
	o.UpdatedAt = now
	return nil
}

func markOrderShipped(o *routingctx.RoutedOrder, actor string, now time.Time) {
	customerOrder, err := customerOrderAggregate(o)
	if err != nil {
		return
	}
	change, changed := customerOrder.MarkShipped(now)
	if !changed {
		return
	}
	applyCustomerOrderSnapshot(o, customerOrder.Snapshot())
	recordOrderChange(o, actor, change, now)
}

func recordActivity(
	o *routingctx.RoutedOrder,
	activityType string,
	actor string,
	message string,
	createdAt time.Time,
	details []routingctx.RoutedOrderActivityDetail,
) {
	message = strings.TrimSpace(message)
	o.ActivityLog = append(o.ActivityLog, routingctx.RoutedOrderActivity{
		Type:      strings.TrimSpace(activityType),
		Actor:     strings.TrimSpace(actor),
		Message:   message,
		Details:   details,
		CreatedAt: createdAt,
	})
}

func customerOrderAggregate(o *routingctx.RoutedOrder) (*orderentity.CustomerOrder, error) {
	return orderentity.RehydrateCustomerOrder(orderentity.CustomerOrderSnapshot{
		ID:                 o.ID,
		StoreID:            o.StoreID,
		CandidateID:        o.CandidateID,
		ProductTitle:       o.ProductTitle,
		Quantity:           o.Quantity,
		Total:              o.Total,
		CustomerName:       o.CustomerName,
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

func applyCustomerOrderSnapshot(o *routingctx.RoutedOrder, snapshot orderentity.CustomerOrderSnapshot) {
	o.CandidateID = snapshot.CandidateID
	o.ProductTitle = snapshot.ProductTitle
	o.Quantity = snapshot.Quantity
	o.Total = snapshot.Total
	o.CustomerName = snapshot.CustomerName
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

func exceptionAggregate(o *routingctx.RoutedOrder) (*exceptionentity.OrderException, error) {
	return exceptionentity.RehydrateOrderException(exceptionentity.OrderExceptionSnapshot{
		OrderID: o.ID,
		Type:    o.ExceptionType,
		Status:  o.ExceptionStatus,
	})
}

func applyExceptionSnapshot(o *routingctx.RoutedOrder, snapshot exceptionentity.OrderExceptionSnapshot) {
	o.ExceptionType = snapshot.Type
	o.ExceptionStatus = snapshot.Status
}

func fulfillmentOrderAggregate(o *routingctx.RoutedOrder) (*fulfillmententity.FulfillmentOrder, error) {
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

func applyFulfillmentSnapshot(o *routingctx.RoutedOrder, snapshot fulfillmententity.FulfillmentOrderSnapshot) {
	o.ShipmentStatus = snapshot.Status
	o.ShipmentCarrier = snapshot.Carrier
	o.ShipmentTrackingNumber = snapshot.TrackingNumber
	o.ShipmentTrackingURL = snapshot.TrackingURL
	o.ShipmentNotes = snapshot.Notes
	o.ShippedAt = snapshot.ShippedAt
	o.DeliveredAt = snapshot.DeliveredAt
}

func settlementRecordAggregate(o *routingctx.RoutedOrder) (*settlemententity.SettlementRecord, error) {
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

func applySettlementSnapshot(o *routingctx.RoutedOrder, snapshot settlemententity.SettlementRecordSnapshot) {
	o.FulfillmentCost = snapshot.FulfillmentCost
	o.ShippingCost = snapshot.ShippingCost
	o.IssueCost = snapshot.IssueCost
	o.IssueResolution = snapshot.IssueResolution
	o.IssueNotes = snapshot.IssueNotes
	o.RealizedMargin = snapshot.RealizedMargin
	o.SettlementStatus = snapshot.Status
	o.SettlementNotes = snapshot.Notes
}

func recordOrderChange(o *routingctx.RoutedOrder, actor string, change orderentity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	recordSystemTimelineActivity(o, actor, change.Message, now, orderDetails(change.Details))
}

func recordExceptionChange(o *routingctx.RoutedOrder, actor string, change exceptionentity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	recordSystemTimelineActivity(o, actor, change.Message, now, exceptionDetails(change.Details))
}

func recordFulfillmentChange(o *routingctx.RoutedOrder, actor string, change fulfillmententity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	recordSystemTimelineActivity(o, actor, change.Message, now, fulfillmentDetails(change.Details))
}

func recordSettlementChange(o *routingctx.RoutedOrder, actor string, change settlemententity.Change, now time.Time) {
	if change.Message == "" {
		return
	}
	recordSystemTimelineActivity(o, actor, change.Message, now, settlementDetails(change.Details))
}

func recordSystemTimelineActivity(
	o *routingctx.RoutedOrder,
	actor string,
	message string,
	createdAt time.Time,
	details []routingctx.RoutedOrderActivityDetail,
) {
	message = strings.TrimSpace(message)
	o.Timeline = append(o.Timeline, message)
	recordActivity(o, routingctx.RoutedOrderActivityTypeSystem, actor, message, createdAt, details)
}

func orderDetails(details []orderentity.ActivityDetail) []routingctx.RoutedOrderActivityDetail {
	out := make([]routingctx.RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, routingctx.RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func exceptionDetails(details []exceptionentity.ActivityDetail) []routingctx.RoutedOrderActivityDetail {
	out := make([]routingctx.RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, routingctx.RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func fulfillmentDetails(details []fulfillmententity.ActivityDetail) []routingctx.RoutedOrderActivityDetail {
	out := make([]routingctx.RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, routingctx.RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}

func settlementDetails(details []settlemententity.ActivityDetail) []routingctx.RoutedOrderActivityDetail {
	out := make([]routingctx.RoutedOrderActivityDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, routingctx.RoutedOrderActivityDetail{Key: detail.Key, Value: detail.Value})
	}
	return out
}
