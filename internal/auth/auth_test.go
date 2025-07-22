package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}

	if err := CheckPasswordHash("password", hash); err != nil {
		t.Fatal(err)
	}
}

func TestMakeAndValidateJWTValid(t *testing.T) {
	tokenSecret := "secret"
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	id, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatal(err)
	}
	if id != userID {
		t.Fatal("user id mismatch")
	}
}

func TestMakeAndValidateJWTInvalid(t *testing.T) {
	tokenSecret := "secret"
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateJWT(tokenString, "wrong secret")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMakeAndValidateJWTExpired(t *testing.T) {
	tokenSecret := "secret"
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, tokenSecret, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{
		"Authorization": {"Bearer tokenTest"},
	}

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatal(err)
	}

	if token != "tokenTest" {
		t.Fatalf("wrong token, expected: tokenTest, got: %s", token)
	}
}

func TestGetBearerTokenEmpty(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected error")
	}
}
