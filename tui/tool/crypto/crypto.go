package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltStorageName      = "Viscue AUC Salt"
	SecretKeyStorageName = "Viscue Secret Key"
)

func GenerateAccountUnlockKey(password, secretKey, username string) (
	[]byte, error,
) {
	salt, err := findOrMakeSalt(username)
	if err != nil {
		return nil, err
	}

	// Calculate the derivative of salt using HKDF.
	kdf := hkdf.New(sha256.New, []byte(salt), []byte(username),
		[]byte("viscue-client"))
	saltByte := make([]byte, 32)
	if _, err := io.ReadFull(kdf, saltByte); err != nil {
		return nil, err
	}
	salt = string(saltByte)

	// Create AUC from password and salt.
	aucByte := pbkdf2.Key([]byte(password), []byte(salt), 100000, 32,
		sha256.New)

	// Trim our secret key to match the length of AUC, so we can XOR both.
	trimmedSecret := []byte(secretKey)[:len(aucByte)]
	for i := 0; i < len(aucByte); i++ {
		aucByte[i] ^= trimmedSecret[i]
	}

	return aucByte, nil
}

// findOrMakeSalt searched salt in keyring. If not found, generate it.
func findOrMakeSalt(username string) (string, error) {
	salt, err := keyring.Get(SaltStorageName, username)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			// Generate salt if previously not found, then save it in keyring.
			salt, err = GenerateSalt()
			if err != nil {
				return "", err
			}

			err = keyring.Set(SaltStorageName, username, salt)
			if err != nil {
				return "", err
			}

			return salt, nil
		}
		return "", err
	}
	return salt, nil
}

func GenerateRsaPrivateKey() (*rsa.PrivateKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, err
	}

	return private, nil
}

func rsaPrivateToPem(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
}

func pemToPrivateRsa(b []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(b)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func EncryptRsaKey(key *rsa.PrivateKey, auc []byte) ([]byte, error) {
	block, err := aes.NewCipher(auc)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, rsaPrivateToPem(key), nil), nil
}

func DecryptRsaKey(ciphertext, auc []byte) (*rsa.PrivateKey, error) {
	block, err := aes.NewCipher(auc)
	if err != nil {
		panic(err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return pemToPrivateRsa(plaintext)
}
