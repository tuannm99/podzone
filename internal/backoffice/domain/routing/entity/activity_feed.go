package entity

import "time"

type RoutedOrderActivityFeedQuery struct {
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
