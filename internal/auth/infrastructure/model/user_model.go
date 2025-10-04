package model

import (
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type User struct {
	ID          uint64    `db:"id"           json:"id"`
	Username    string    `db:"username"     json:"username"`
	Email       string    `db:"email"        json:"email"`
	Password    string    `db:"password"     json:"password"`
	FullName    string    `db:"full_name"    json:"full_name"`
	MiddleName  string    `db:"middle_name"  json:"middle_name"`
	FirstName   string    `db:"first_name"   json:"first_name"`
	LastName    string    `db:"last_name"    json:"last_name"`
	Address     string    `db:"address"      json:"address"`
	InitialFrom string    `db:"initial_from" json:"initial_from"`
	Age         uint8     `db:"age"          json:"age"`
	Dob         time.Time `db:"dob"          json:"dob"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

func (u *User) ToEntity() (*entity.User, error) {
	return toolkit.MapStruct[User, entity.User](*u)
}
