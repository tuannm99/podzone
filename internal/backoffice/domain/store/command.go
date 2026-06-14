package store

type CreateStoreCmd struct {
	Name        string
	Description string
}

type UpdateStoreStatusCmd struct {
	ID     string
	Active bool
}

type BootstrapStoreCmd struct {
	ID      string
	Name    string
	OwnerID string
}
