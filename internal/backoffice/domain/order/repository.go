package order

type CustomerOrderRepository interface {
	CustomerOrderCommandRepository
	CustomerOrderQueryRepository
}
