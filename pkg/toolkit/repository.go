package toolkit

type Create[T any] interface {
	Create(e T) (*T, error)
}

type GetByID[T any] interface {
	GetByID(id string) (*T, error)
}

type Update[T any] interface {
	Update(e T) error
}

type UpdateById[T any] interface {
	UpdateById(id uint, e T) error
}

type DeleteByID interface {
	DeleteByID(id uint) error
}

type DeleteManyByIDs interface {
	DeleteManyByIDs(ids []uint) error
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
