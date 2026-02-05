package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
	jwt.RegisteredClaims
}

func ParseToken(token string, secret string) (Claims, error) {
	if token == "" {
		return Claims{}, errors.New("missing token")
	}

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	if strings.TrimSpace(claims.GameID) == "" || strings.TrimSpace(claims.Player) == "" {
		return Claims{}, errors.New("missing claims")
	}
	return *claims, nil
}

func SignToken(secret string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
