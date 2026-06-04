package store

type CreateStoreCmd struct {
	Name        string
	Description string
}

type UpdateStoreStatusCmd struct {
	ID     string
	Active bool
}
