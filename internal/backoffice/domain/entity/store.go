package entity

import "time"

type Store struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerID     string    `json:"owner_id"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
