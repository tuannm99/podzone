package pdauthn

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type Verifier struct {
	secret string
	key    string
}

func NewVerifier(cfg Config) *Verifier {
	return &Verifier{
		secret: cfg.JWTSecret,
		key:    cfg.JWTKey,
	}
}

func (v *Verifier) ClaimsFromContext(ctx context.Context) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing request metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, errors.New("missing authorization header")
	}
	raw := strings.TrimSpace(values[0])
	if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return nil, errors.New("invalid authorization header")
	}
	return v.ClaimsFromTokenString(strings.TrimSpace(raw[len("Bearer "):]))
}

func (v *Verifier) ClaimsFromAccessToken(accessToken string) (*Claims, error) {
	tokenString := strings.TrimSpace(accessToken)
	if tokenString == "" {
		return nil, errors.New("missing access token")
	}
	return v.ClaimsFromTokenString(tokenString)
}

func (v *Verifier) ClaimsFromTokenString(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(v.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if v.key != "" && claims.Key != v.key {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}
