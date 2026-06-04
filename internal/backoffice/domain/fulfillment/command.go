package fulfillment

type UpdateOrderShipmentCmd struct {
	StoreID        string
	OrderID        string
	ShipmentStatus string
	Carrier        string
	TrackingNumber string
	TrackingURL    string
	Notes          string
}
