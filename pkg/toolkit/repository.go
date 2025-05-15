package toolkit

type Create[T any] interface {
	Create(entity T) (*T, error)
}

type GetByID[T any] interface {
	GetByID(id string) (*T, error)
}

type Update[T any] interface {
	Update(entity T) error
}

type UpdateById[T any] interface {
	UpdateById(id string, entity T) error
}

type DeleteByID interface {
	DeleteByID(id string) error
}

type DeleteManyByIDs interface {
	DeleteManyByIDs(ids []string) error
}

type CR[T any] interface {
	Create[T]
	GetByID[T]
}

type CRL[T any] interface {
	CR[T]
}

type CRUD[T any] interface {
	Create[T]
	GetByID[T]
	Update[T]
	UpdateById[T]
	DeleteByID
}

type All[T any] interface {
	CRUD[T]
	DeleteManyByIDs
}
