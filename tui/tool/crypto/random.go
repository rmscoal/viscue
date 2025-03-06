package crypto

import (
	"bytes"
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
	// Before picking random index from the entropy, we would
	// also want to shuffle the entropy's indexing.
	// If usually we have the entropy of "abc...ABC...123...#$%",
	// we want random arrangement "a&3BZs#0..." as the entropy.

	shuffledEntropy := strings.Split(completeEntropy, "")
	N := len(shuffledEntropy)
	for i := 0; i < N; i++ {
		randIdx, err := rand.Int(rand.Reader, big.NewInt(512))
		if err != nil {
			return "", err
		}
		j := randIdx.Int64() % int64(N)
		shuffledEntropy[i], shuffledEntropy[j] = shuffledEntropy[j], shuffledEntropy[i]
	}

	return generateRandomString(strings.Join(shuffledEntropy, ""), length)
}

func generateRandomString(entropy string, length int) (string, error) {
	var b bytes.Buffer

	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(entropy))))
		if err != nil {
			return "", err
		}

		b.WriteByte(entropy[index.Int64()])
	}

	return b.String(), nil
}
