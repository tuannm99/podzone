package routing

import "time"

type RoutedOrderActivityFeedQuery struct {
	StoreID       string
	ActivityType  string
	ActorContains string
	OrderID       string
	Partner       string
	Assignee      string
	Since         *time.Time
	Limit         int
	After         string
	IncludeSystem bool
}
