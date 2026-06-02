package outputport

type FulfillmentOrderRepository interface {
	FulfillmentOrderCommandRepository
	FulfillmentOrderQueryRepository
}
