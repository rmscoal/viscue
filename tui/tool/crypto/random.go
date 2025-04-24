package crypto

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	lowerLettersCharacters = "abcdefghijklmnopqrstuvwxyz"
	upperLettersCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbersCharacters      = "0123456789"
	specialCharacters      = "~!@#%^&*-_+={}|"

	secretKeyEntropy = upperLettersCharacters + numbersCharacters
	saltEntropy      = lowerLettersCharacters + upperLettersCharacters + numbersCharacters
	completeEntropy  = lowerLettersCharacters + upperLettersCharacters + numbersCharacters + specialCharacters
)

func GenerateSecretKey() (string, error) {
	return generateRandomString(secretKeyEntropy, 36)
}

func GenerateSalt() (string, error) {
	return generateRandomString(saltEntropy, 32)
}

func GenerateRandomPassword(length int) (string, error) {
	return generateRandomString(completeEntropy, length)
}

func generateRandomString(entropy string, length int) (string, error) {
	var b strings.Builder
	b.Grow(length)

	entropyLen := int64(len(entropy))
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(entropyLen))
		if err != nil {
			return "", err
		}

		b.WriteByte(entropy[index.Int64()])
	}

	return b.String(), nil
}
