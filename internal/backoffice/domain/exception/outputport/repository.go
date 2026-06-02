package outputport

type OrderExceptionRepository interface {
	OrderExceptionCommandRepository
	OrderExceptionQueryRepository
}
