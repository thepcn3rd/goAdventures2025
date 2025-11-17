package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

// HashPassword generates a PBKDF2 hash of the password with a random salt
func HashPassword(password string) (string, string, int, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", 0, err
	}

	// PBKDF2 parameters
	iterations := 100000
	keyLength := 32 // 32 bytes = 256 bits

	// Derive the key
	derivedKey := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// Encode the results for storage
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedKey := base64.StdEncoding.EncodeToString(derivedKey)

	return encodedKey, encodedSalt, iterations, nil
}

// VerifyPassword checks if the provided password matches the stored hash
func VerifyPassword(password, storedHash, storedSalt string, iterations int) (bool, error) {
	// Decode the stored salt
	salt, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return false, err
	}

	// Decode the stored hash
	storedKey, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the provided password
	derivedKey := pbkdf2.Key([]byte(password), salt, iterations, len(storedKey), sha256.New)

	// Compare the derived key with the stored key
	if len(derivedKey) != len(storedKey) {
		return false, errors.New("hash lengths don't match")
	}

	match := true
	for i := 0; i < len(derivedKey); i++ {
		match = match && (derivedKey[i] == storedKey[i])
	}

	return match, nil
}

func main() {
	// Example usage
	password := "mySecurePassword123!"

	// Hash the password (do this when creating/updating a password)
	hashedPassword, salt, iterations, err := HashPassword(password)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hashed Password: %s\n", hashedPassword)
	fmt.Printf("Salt: %s\n", salt)
	fmt.Printf("Iterations: %d\n", iterations)

	// Later, when verifying a password...
	fmt.Println("\nTesting password verification:")

	// Test with correct password
	fmt.Println("Testing correct password:")
	isValid, err := VerifyPassword(password, hashedPassword, salt, iterations)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Password valid: %v\n", isValid)

	// Test with incorrect password
	fmt.Println("\nTesting wrong password:")
	isValid, err = VerifyPassword("wrongPassword", hashedPassword, salt, iterations)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Password valid: %v\n", isValid)
}
