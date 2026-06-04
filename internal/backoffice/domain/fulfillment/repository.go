package fulfillment

type FulfillmentOrderRepository interface {
	FulfillmentOrderCommandRepository
	FulfillmentOrderQueryRepository
}
