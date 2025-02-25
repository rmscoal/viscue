package prompt

import (
	"crypto/rsa"
	"fmt"
	"strings"

	"viscue/tui/entity"
	"viscue/tui/tool/cache"
	"viscue/tui/views/library/message"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Close() tea.Msg {
	switch m.payload.(type) {
	case entity.Category:
		return message.ClosePromptMsg[entity.Category]{}
	case entity.Password:
		return message.ClosePromptMsg[entity.Password]{}
	default:
		return nil
	}
}

type DataSubmittedMsg[T any] struct {
	Data T
}

func (m Model) Submit() tea.Msg {
	switch payload := m.payload.(type) {
	case entity.Category:
		payload = m.buildCategoryEntity()
		if payload.Id == 0 {
			res, err := m.db.NamedExec(`INSERT INTO categories (name) VALUES (:name) RETURNING id`,
				&payload)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			payload.Id = id
		} else {
			_, err := m.db.NamedExec("UPDATE categories SET name = :category WHERE id = :id",
				payload)
			if err != nil {
				return err
			}
		}
		return DataSubmittedMsg[entity.Category]{Data: payload}
	case entity.Password:
		payload = m.buildPasswordEntity()
		publicKey := cache.Get[*rsa.PublicKey](cache.PublicKey)
		enc := payload.Copy()
		if err := enc.Encrypt(publicKey); err != nil {
			return fmt.Errorf("failed to encrypt entity: %w", err)
		}
		if payload.Id == 0 {
			res, err := m.db.NamedExec(
				`INSERT INTO
			    	passwords (name, category_id, email, username, password) 
				VALUES (:name, :category_id, :email, :username, :password)
				RETURNING id`,
				&enc,
			)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			payload.Id = id
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

type DeleteErrorMsg struct {
	Error error
}

type DeleteConfirmedMsg[T interface {
	entity.Category | entity.Password
}] struct {
	Payload T
}

func (m Model) Delete() tea.Msg {
	if !m.isDeletion {
		return nil
	}

	switch payload := m.payload.(type) {
	case entity.Category:
		_, err := m.db.Exec("DELETE FROM categories WHERE id = ?", payload.Id)
		if err != nil {
			return err
		}
		return DeleteConfirmedMsg[entity.Category]{
			Payload: payload,
		}
	case entity.Password:
		_, err := m.db.Exec("DELETE FROM passwords WHERE id = ?", payload.Id)
		if err != nil {
			return err
		}
		return DeleteConfirmedMsg[entity.Password]{
			Payload: payload,
		}
	default:
		return nil
	}
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
		CategoryId: m.payload.(entity.Password).CategoryId,
		Email:      strings.ToLower(m.fields[2].Value()),
		Username:   m.fields[3].Value(),
		Password:   m.fields[4].Value(),
	}
}
