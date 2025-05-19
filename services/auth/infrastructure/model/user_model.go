package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
)

type User struct {
	gorm.Model

	Username    string    `gorm:"uniqueIndex" json:"username"`
	Email       string    `gorm:"uniqueIndex" json:"email"`
	Password    string    `                   json:"password"`
	FullName    string    `                   json:"full_name"`
	MiddleName  string    `                   json:"middle_name"`
	FirstName   string    `                   json:"first_name"`
	LastName    string    `                   json:"last_name"`
	Address     string    `                   json:"address"`
	InitialFrom string    `                   json:"initial_from"`
	Age         uint8     `                   json:"age"`
	Dob         time.Time `                   json:"dob"`
	CreatedAt   time.Time `                   json:"created_at"`
	UpdatedAt   time.Time `                   json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToEntity() *entity.User {
	return toolkit.MapStruct[User, entity.User](*u)
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	var count int64

	tx.Model(&User{}).Where("username = ?", u.Username).Count(&count)
	if count > 0 {
		return fmt.Errorf("username %s already exists", u.Username)
	}

	tx.Model(&User{}).Where("email = ?", u.Email).Count(&count)
	if count > 0 {
		return fmt.Errorf("email %s already exists", u.Email)
	}

	if u.Password != "" {
		pass, err := entity.GeneratePasswordHash(u.Password)
		if err != nil {
			return err
		}

		u.Password = pass
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	var count int64

	tx.Model(&User{}).Where("username = ? AND id != ?", u.Username, u.ID).Count(&count)
	if count > 0 {
		return fmt.Errorf("username %s already exists", u.Username)
	}

	tx.Model(&User{}).Where("email = ? AND id != ?", u.Email, u.ID).Count(&count)
	if count > 0 {
		return fmt.Errorf("email %s already exists", u.Email)
	}

	if u.Password != "" {
		pass, err := entity.GeneratePasswordHash(u.Password)
		if err != nil {
			return err
		}

		u.Password = pass
	}

	return nil
}
