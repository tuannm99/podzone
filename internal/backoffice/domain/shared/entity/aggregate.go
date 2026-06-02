package entity

// AggregateRoot marks a domain object that owns transactional consistency for an aggregate.
type AggregateRoot interface {
	AggregateID() string
}
