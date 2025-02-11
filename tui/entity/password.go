package entity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"viscue/tui/component/table"

	"github.com/charmbracelet/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/sync/errgroup"
)

type Password struct {
	Id         int64         `db:"id"`
	CategoryId sql.NullInt64 `db:"category_id"`
	Name       string        `db:"name"`
	Email      string        `db:"email"`
	Username   string        `db:"username"`
	Password   string        `db:"password"`
}

func (password Password) Validate() error {
	err := validation.ValidateStruct(&password,
		validation.Field(&password.Name, validation.Required),
		validation.Field(&password.Email, validation.Required),
		validation.Field(&password.Password, validation.Required),
	)
	if err != nil {
		msg := strings.Split(err.Error(), "; ")
		return fmt.Errorf(strings.Join(msg, " and "))
	}

	return nil
}

func (password Password) Copy() Password {
	return Password{
		Id:         password.Id,
		CategoryId: password.CategoryId,
		Name:       password.Name,
		Email:      password.Email,
		Username:   password.Username,
		Password:   password.Password,
	}
}

func (password *Password) Encrypt(pub *rsa.PublicKey) error {
	group, _ := errgroup.WithContext(context.TODO())
	group.Go(func() error { // Encrypt email
		b, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub,
			[]byte(password.Email), []byte(password.Name))
		if err != nil {
			return err
		}
		password.Email = hex.EncodeToString(b)
		return nil
	})
	group.Go(func() error { // Encrypt password
		b, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub,
			[]byte(password.Password), []byte(password.Name))
		if err != nil {
			return err
		}
		password.Password = hex.EncodeToString(b)
		return nil
	})

	return group.Wait()
}

func (password *Password) Decrypt(priv *rsa.PrivateKey) error {
	group, _ := errgroup.WithContext(context.TODO())
	group.Go(func() error { // Decrypt email
		decoded, err := hex.DecodeString(password.Email)
		if err != nil {
			return err
		}

		plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv,
			decoded, []byte(password.Name))
		if err != nil {
			return err
		}

		password.Email = string(plaintext)
		return nil
	})
	group.Go(func() error { // Decrypt password
		decoded, err := hex.DecodeString(password.Password)
		if err != nil {
			return err
		}

		plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv,
			decoded, []byte(password.Name))
		if err != nil {
			return err
		}

		password.Password = string(plaintext)
		return nil
	})

	return group.Wait()
}

func (password Password) ToTableRow() table.Row {
	username := "-"
	if password.Username != "" {
		username = password.Username
	}

	return table.Row{
		strconv.FormatInt(password.Id, 10),               // ID (hidden)
		strconv.FormatInt(password.CategoryId.Int64, 10), // CategoryId (hidden)
		password.Name,
		password.Email,
		username,
		password.Password, // Password (hidden)
	}
}

func NewPasswordFromTableRow(row table.Row) (Password, error) {
	password := Password{
		Name:     row[2],
		Email:    row[3],
		Username: row[4],
		Password: row[5],
	}
	id, err := strconv.ParseInt(row[0], 10, 64)
	if err != nil {
		log.Error("entity.NewPasswordFromTableRow: something went wrong when parsing id",
			"str", row[0], "err", err)
		return password, err
	}

	password.Id = id
	categoryId, err := strconv.ParseInt(row[1], 10, 64)
	if err != nil {
		log.Error("entity.NewPasswordFromTableRow: something went wrong when parsing categoryId",
			"str", row[1], "err", err)
		return password, err
	}

	if categoryId > 0 {
		password.CategoryId = sql.NullInt64{Valid: true, Int64: categoryId}
	}

	return password, nil
}
