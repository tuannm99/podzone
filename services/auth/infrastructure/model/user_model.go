package model

import (
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string
	Email       string
	FullName    string
	MiddleName  string
	FirstName   string
	LastName    string
	Address     string
	InitialFrom string
	Age         uint8
	Dob         time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToEntity() *entity.User {
	return toolkit.MapStruct[User, entity.User](*u)
}
