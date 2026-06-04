package settlement

type UpdateOrderSettlementCmd struct {
	StoreID          string
	OrderID          string
	FulfillmentCost  string
	ShippingCost     string
	SettlementStatus string
	Notes            string
}

type UpdateOrderIssueHandlingCmd struct {
	StoreID         string
	OrderID         string
	IssueCost       string
	IssueResolution string
	Notes           string
}
