package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	ArgonKeyLength  = 32
	ArgonMemory     = 64 * 1024
	ArgonThreads    = 4
	ArgonIterations = 12
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, ArgonIterations, ArgonMemory,
		ArgonThreads, ArgonKeyLength)

	encSalt := base64.RawStdEncoding.EncodeToString(salt)
	encPass := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$%s$%s",
		argon2.Version,
		encSalt, encPass), nil
}

func MatchPassword(givenPassword, storedPassword string) (bool, error) {
	vals := strings.Split(storedPassword, "$")
	if len(vals) != 5 {
		return false, errors.New("invalid password hash")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, errors.New("failed to parse version from argon string")
	}

	if version != argon2.Version {
		return false, errors.New("incompatible password hash version")
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return false, errors.New("failed to decode salt from argon string")
	}

	hash, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, errors.New("failed to decode hash from argon string")
	}

	comparison := argon2.IDKey([]byte(givenPassword), salt, ArgonIterations,
		ArgonMemory, ArgonThreads, ArgonKeyLength)

	if subtle.ConstantTimeCompare(hash, comparison) != 1 {
		return false, nil
	}

	return true, nil
}
