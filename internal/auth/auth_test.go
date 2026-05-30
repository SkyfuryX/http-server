package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTValidation(t *testing.T) {
	id, err := uuid.Parse("31dbe93a-1390-4bac-bf0e-f42b74375185")
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := MakeJWT(id, "potato", time.Hour*1)
	if err != nil {
		t.Fatal(err)
	}

	id2, err := ValidateJWT(tokenStr, "potato")
	if err != nil {
		t.Fatal(err)
	}

	if id != id2 {
		t.Fatalf("uuid mismatch: have %v, want %v", id2, id)
	}
}

func TestJWTExpired(t *testing.T) {
	id, err := uuid.Parse("31dbe93a-1390-4bac-bf0e-f42b74375185")
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := MakeJWT(id, "potato", time.Second*1)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	_, err = ValidateJWT(tokenStr, "potato") //token should be expired
	if err == nil {
		t.Fatal("Token not expired")
	}
}

func TestJWTInvalidSecret(t *testing.T) {
	id, err := uuid.Parse("31dbe93a-1390-4bac-bf0e-f42b74375185")
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := MakeJWT(id, "potato", time.Second*1)
	if err != nil {
		t.Fatal(err)
	}
	
	_, err = ValidateJWT(tokenStr, "tomato") //token should be mismatched
	if err == nil {
		t.Fatal("Secret not mismatched")
	}
}
