package ddd

import "context"

// Repository is a small generic contract for aggregate persistence. Service
// domains should still own their specific repository ports when queries or
// commands need business-specific methods.
type Repository[T AggregateRoot] interface {
	FindByID(ctx context.Context, id ID) (T, error)
	Save(ctx context.Context, aggregate T) error
}
