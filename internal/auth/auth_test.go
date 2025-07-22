package auth

import "testing"

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}

	if err := CheckPasswordHash("password", hash); err != nil {
		t.Fatal(err)
	}
}
