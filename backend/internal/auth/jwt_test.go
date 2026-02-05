package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseTokenValid(t *testing.T) {
	secret := "secret"
	claims := Claims{
		GameID: "game-1",
		Player: "p1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	parsed, err := ParseToken(signed, secret)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if parsed.GameID != claims.GameID || parsed.Player != claims.Player {
		t.Fatalf("unexpected claims: %+v", parsed)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	if _, err := ParseToken("", "secret"); err == nil {
		t.Fatalf("expected error for empty token")
	}

	claims := Claims{GameID: "game-1", Player: "p1"}
	badToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, err := badToken.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := ParseToken(signed, "secret"); err == nil {
		t.Fatalf("expected error for invalid signing method")
	}

	missing := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{GameID: "", Player: ""})
	signedMissing, err := missing.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := ParseToken(signedMissing, "secret"); err == nil {
		t.Fatalf("expected error for missing claims")
	}
}
