package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "mySecretPassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}
	if hash1 == "" || hash1 == password {
		t.Fatalf("hash should not be empty or plain password")
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password second time: %v", err)
	}

	// Argon2id generates unique salts, so separate hashes must differ
	if hash1 == hash2 {
		t.Errorf("expected different hashes due to salting, but got identical hashes")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mySecretPassword123"
	wrongPassword := "wrongPassword456"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Correct password
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("unexpected error checking password: %v", err)
	}
	if !match {
		t.Errorf("expected password to match hash, but got false")
	}

	// Incorrect password
	match, err = CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("unexpected error checking incorrect password: %v", err)
	}
	if match {
		t.Errorf("expected incorrect password to fail, but got true")
	}

	// Invalid hash format
	_, err = CheckPasswordHash(password, "invalid_hash_format")
	if err == nil {
		t.Errorf("expected error when checking against invalid hash format, got nil")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Logf("MakeJWT returned error: %v", err)
		return
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	if parsedID != userID {
		t.Errorf("expected user ID %v, got %v", userID, parsedID)
	}
}

