package service

import (
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type AuthService struct{ secret string }

func NewAuthService(secret string) *AuthService { return &AuthService{secret: secret} }
func (s *AuthService) Login(username, password string) (string, error) {
	if username != "admin" || password != "admin123" {
		return "", apperrors.ErrUnauthorized
	}
	claims := jwt.MapClaims{"sub": username, "role": constants.RoleAdmin, "exp": time.Now().Add(time.Hour * 24).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	value, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return value, nil
}
func (s *AuthService) Parse(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) { return []byte(s.secret), nil })
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	return claims, nil
}
