package prompt

import (
	"crypto/rsa"
	"database/sql"
	"fmt"

	"viscue/tui/entity"
	"viscue/tui/tool/cache"

	tea "github.com/charmbracelet/bubbletea"
)

type CloseMsg struct{}

func (m Model) Close() tea.Msg {
	return CloseMsg{}
}

type DataSubmittedMsg[T any] struct {
	Data T
}

func (m Model) Submit() tea.Msg {
	switch payload := m.payload.(type) {
	case entity.Category:
		payload = m.buildCategoryEntity()
		if payload.Id == 0 {
			row := m.db.QueryRowx(
				`INSERT INTO categories (name) VALUES (:category) RETURNING id`,
				&payload,
			)
			if err := row.Scan(&payload.Id); err != nil {
				return err
			}
		} else {
			_, err := m.db.NamedExec("UPDATE categories SET name = :category WHERE id = :id",
				payload)
			if err != nil {
				return err
			}
		}
		return DataSubmittedMsg[entity.Category]{Data: payload}
	case entity.Password:
		publicKey := cache.Get[*rsa.PublicKey](cache.PublicKey)
		enc := payload.Copy()
		if err := enc.Encrypt(publicKey); err != nil {
			return fmt.Errorf("failed to encrypt entity: %w", err)
		}
		if payload.Id == 0 {
			row := m.db.QueryRowx(`
				INSERT INTO
			    	passwords (name, category_id, email, username, password) 
				VALUES (:name, :category_id, :email, :username, :password)
				RETURNING id`,
				&enc,
			)
			if err := row.Scan(&payload.Id); err != nil {
				return fmt.Errorf("failed to create password: %w", err)
			}
		} else {
			_, err := m.db.NamedExec(
				`UPDATE passwords SET
						category_id = :category_id,
						name = :name,
						email = :email,
						username = :username,
						password = :password
					WHERE id = :id`,
				&enc,
			)
			if err != nil {
				return fmt.Errorf("failed to update password: %w", err)
			}
		}
		return DataSubmittedMsg[entity.Password]{Data: payload}
	}

	return nil
}

func (m Model) buildCategoryEntity() entity.Category {
	return entity.Category{
		Id:   m.payload.(entity.Category).Id,
		Name: m.fields[0].Value(),
	}
}

func (m Model) buildPasswordEntity() entity.Password {
	return entity.Password{
		Id:         m.payload.(entity.Password).Id,
		Name:       m.fields[0].Value(),
		CategoryId: sql.NullInt64{},
		Email:      m.fields[2].Value(),
		Username:   m.fields[3].Value(),
		Password:   m.fields[4].Value(),
	}
}
