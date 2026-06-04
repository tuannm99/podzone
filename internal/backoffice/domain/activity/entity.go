package activity

import "time"

type ActivityEntry struct {
	StoreID          string
	OrderID          string
	ProductTitle     string
	Partner          string
	OperatorAssignee string
	Type             string
	Actor            string
	Message          string
	Details          []ActivityDetail
	CreatedAt        time.Time
}

type ActivityDetail struct {
	Key   string
	Value string
}
