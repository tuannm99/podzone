package exception

type OpenOrderExceptionCmd struct {
	StoreID       string
	OrderID       string
	ExceptionType string
}

type UpdateOrderExceptionStatusCmd struct {
	StoreID string
	OrderID string
	Status  string
}
