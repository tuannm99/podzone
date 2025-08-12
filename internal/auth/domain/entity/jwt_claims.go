package entity

import "github.com/golang-jwt/jwt"

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"user_name"`
	Key      string `json:"key"`
	jwt.StandardClaims
}
