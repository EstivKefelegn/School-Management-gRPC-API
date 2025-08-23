package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPasswword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
	if len(parts) < 2 {
		return ErrorHandler(errors.New("invalid encoded hash format"), "invalid")
	}

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err :=base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(errors.New("failed to decode the salt"), "Failed to decode the salt")
	}	
	hashedassword, err := base64.RawStdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return ErrorHandler(errors.New("failed to decode the hash"), "Failed to decode the hash")
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedassword) {
		return ErrorHandler(errors.New("incorrect length of password"), "Incorrect length of password")
	}

	if subtle.ConstantTimeCompare(hash, hashedassword) == 1 {
		return nil
	}
		return ErrorHandler(errors.New("incorrect password"), "Incorrect password")
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrorHandler(errors.New("password is blank"), "please enter the password")
	}

	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to generate salt"), "error adding data")
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	password = encodedHash

	return encodedHash, nil
}