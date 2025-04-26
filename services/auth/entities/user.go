package entities

import "time"

type User struct {
	ID         string
	Username   string
	Email      string
	FullName   string
	MiddleName string
	FirstName  string
	LastName   string
	Address    string
	Age        uint8
	Dob        time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
