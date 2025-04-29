package entity

import "github.com/golang-jwt/jwt"

type JWTClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
	jwt.StandardClaims
}
