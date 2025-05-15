package entity

import (
	"errors"
	"time"
)

type User struct {
	ID         uint
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

var (
	ErrUsernameExisted = errors.New("username existed")
	ErrEmailExisted    = errors.New("email existed")
)
