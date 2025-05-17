package entity

import (
	"errors"
	"time"
)

type User struct {
	ID          uint
	Username    string
	Password    string
	Email       string
	FullName    string
	MiddleName  string
	FirstName   string
	LastName    string
	Address     string
	Age         uint8
	InitialFrom string
	Dob         time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrUsernameExisted = errors.New("username existed")
	ErrEmailExisted    = errors.New("email existed")

	ErrUserNotFound  = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
)
