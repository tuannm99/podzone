package outputport

type CustomerOrderRepository interface {
	CustomerOrderCommandRepository
	CustomerOrderQueryRepository
}
