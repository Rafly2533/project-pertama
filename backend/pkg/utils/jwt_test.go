package utils

import (
	"testing"
	"time"
)

func TestTokenRoundTrip(t *testing.T) {
	token, err := GenerateToken("01234567890123456789012345678901", time.Hour, 7, "intan", "staff")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseToken("01234567890123456789012345678901", token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.Username != "intan" || claims.Role != "staff" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken("01234567890123456789012345678901", time.Hour, 1, "admin", "super_admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken("abcdefghijklmnopqrstuvwxyz123456", token); err == nil {
		t.Fatal("expected invalid token")
	}
}

func TestTokenRejectsExpiredToken(t *testing.T) {
	token, err := GenerateToken("01234567890123456789012345678901", -time.Second, 1, "admin", "super_admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken("01234567890123456789012345678901", token); err == nil {
		t.Fatal("expected expired token")
	}
}
