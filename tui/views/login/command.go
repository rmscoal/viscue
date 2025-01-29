package login

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"viscue/tui/event"
	"viscue/tui/tool/cache"
	"viscue/tui/tool/crypto"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/zalando/go-keyring"
)

type loginRequest struct {
	Username string
	Password string
}

func (r loginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	)
}

func (m *login) submit() tea.Msg {
	req := loginRequest{
		Username: m.usernameInput.Value(),
		Password: m.passwordInput.Value(),
	}

	if err := req.Validate(); err != nil {
		msg := strings.Split(err.Error(), "; ")
		return fmt.Errorf(strings.Join(msg, " and "))
	}

	if m.shouldCreateAccount {
		return m.signup()
	}
	return m.login()
}

// signup is a tea.Cmd that registers anc account
// The steps are as follows:
// 1. Hash password and save username password in DB.
// 2. Generate and store secret key in keyring.
// 3. Generate RSA Private key.
// 4. Compute Account Unlock Key (AUC).
// 5. Encrypt RSA Private Key with AUC.
// 6. Store encrypted RSA in keyring.
func (m *login) signup() tea.Msg {
	username := m.usernameInput.Value()
	password := m.passwordInput.Value()
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	tx, err := m.db.Beginx()
	if err != nil {
		log.Error("failed to start transaction", "err", err)
		_ = tx.Rollback()
		return errors.New("something went wrong with sqlite database")
	}

	_, err = tx.Exec(
		"INSERT INTO configurations VALUES (?, ?), (?, ?)",
		"username", username, "password", hashedPassword,
	)
	if err != nil {
		log.Error("failed to insert to configurations", "err", err)
		_ = tx.Rollback()
		return errors.New("failed saving credentials to database")
	}

	sc, err := crypto.GenerateSecretKey()
	if err != nil {
		log.Error("failed to generate secret key", "err", err)
		_ = tx.Rollback()
		return errors.New("failed generating secret key")
	}

	err = keyring.Set(crypto.SecretKeyStorageName, username, sc)
	if err != nil {
		log.Error("failed saving secret key in keyring", "err", err)
		_ = tx.Rollback()
		return errors.New("failed saving secret key")
	}

	auc, err := crypto.GenerateAccountUnlockKey(password, sc, username)
	if err != nil {
		log.Error("failed generating account unlock key", "err", err)
		_ = tx.Rollback()
		return errors.New("failed generating account unlock key")
	}

	privateKey, err := crypto.GenerateRsaPrivateKey()
	if err != nil {
		log.Error("failed to generate private key", "err", err)
		_ = tx.Rollback()
		return errors.New("failed generating private key")
	}

	encPrivateKey, err := crypto.EncryptRsaKey(privateKey, auc)
	if err != nil {
		log.Error("failed to encrypt private key", "err", err)
		_ = tx.Rollback()
		return errors.New("failed saving private key")
	}

	_, err = tx.Exec(
		"INSERT INTO configurations VALUES (?, ?)",
		"encrypted_private_key", hex.EncodeToString(encPrivateKey))
	if err != nil {
		log.Error("failed to save encrypted private key", "err", err)
		_ = tx.Rollback()
		return errors.New("failed saving private key")
	}

	// Store necessary values in cache
	cache.Set(cache.AccountUnlockKey, auc)
	cache.Set(cache.PrivateKey, privateKey)
	cache.Set(cache.PublicKey, &privateKey.PublicKey)

	err = tx.Commit()
	if err != nil {
		log.Error("failed to commit transaction", "err", err)
		_ = tx.Rollback()
		_ = keyring.Delete(crypto.SecretKeyStorageName, username)
		_ = keyring.Delete(crypto.SaltStorageName, username)
		return errors.New("something went wrong while saving to database")
	}

	return event.UserLoggedIn
}

// login is a tea.Cmd that signs in user given by the username
// and password. The flow is the following:
// 1. Authenticate user by comparing passwords
// 2. Once authenticated, retrieve secret key from keyring
// 3. Retrieve encrypted private key from keyring and decode it
// 4. Generate AUC to decrypt our private key
// 5. Compute public key from private key
// 6. Return SetStoreMessages with payloads.
func (m *login) login() tea.Msg {
	username := m.usernameInput.Value()
	password := m.passwordInput.Value()

	var hashedPassword string
	err := m.db.QueryRowx(
		"SELECT value FROM configurations WHERE key = ?", "password").
		Scan(&hashedPassword)
	if err != nil {
		log.Error("failed querying password from database", "err", err)
		return errors.New("failed querying password from database")
	}

	match, err := crypto.MatchPassword(password, hashedPassword)
	if err != nil {
		return err
	} else if !match {
		return errors.New("authentication failed password mismatched")
	}

	sc, err := keyring.Get(crypto.SecretKeyStorageName, username)
	if err != nil {
		log.Error("failed to find secret key in keyring", "err", err)
		return errors.New("secret key was not found")
	}

	auc, err := crypto.GenerateAccountUnlockKey(password, sc, username)
	if err != nil {
		log.Error("failed to generate account unlock key", "err", err)
		return errors.New("failed generating account unlock key")
	}

	var encodedEncryptedPrivateKey string
	err = m.db.QueryRowx("SELECT value FROM configurations WHERE key = ?",
		"encrypted_private_key").Scan(&encodedEncryptedPrivateKey)
	if err != nil {
		log.Error("failed querying encrypted private key from database", "err",
			err)
		return errors.New("failed querying encrypted private key from database")
	}

	encryptedPrivateKey, err := hex.DecodeString(encodedEncryptedPrivateKey)
	if err != nil {
		log.Error("failed decoding encrypted private key", "err", err)
		return errors.New("failed decoding encrypted private key")
	}

	privateKey, err := crypto.DecryptRsaKey(encryptedPrivateKey, auc)
	if err != nil {
		log.Error("failed decrypting private key", "err", err)
		return errors.New("failed decrypting private key")
	}

	// Store necessary values in cache
	cache.Set(cache.AccountUnlockKey, auc)
	cache.Set(cache.PrivateKey, privateKey)
	cache.Set(cache.PublicKey, &privateKey.PublicKey)

	return event.UserLoggedIn
}
