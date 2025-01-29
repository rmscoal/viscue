package prompt

import (
	"crypto/rsa"
	"errors"

	"viscue/tui/entity"
	"viscue/tui/event"
	"viscue/tui/tool/cache"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

func (m *prompt) close() tea.Msg {
	return event.ClosePrompt
}

var insertCategoryQuery = `
	INSERT INTO categories (name)
	VALUES (:name) ON CONFLICT (name)
	DO UPDATE SET name=:name
`

var updateCategoryQuery = `
	UPDATE categories
	SET name=:name
	WHERE id=:id
`

var insertPasswordQuery = `
	INSERT INTO passwords (category_id, name, email, username, password) 
	VALUES (
		:category_id,
	    :name,
		:email,
		:username,
		:password
	) ON CONFLICT (id) 
	DO UPDATE SET
		category_id=:category_id,
		name=:name,
		email=:email,
		username=:username,
		password=:password
`

var updatePasswordQuery = `
	UPDATE passwords
	SET
		category_id=:category_id,
		name=:name,
		email=:email,
		username=:username,
		password=:password
	WHERE id=:id
`

type SubmitData any

func (m *prompt) submit() tea.Msg {
	switch model := m.model.(type) {
	case entity.Category:
		model.Name = m.fields[0].Value()
		if err := model.Validate(); err != nil {
			return err
		}

		if model.Id == 0 {
			res, err := m.db.NamedExec(insertCategoryQuery, &model)
			if err != nil {
				log.Error("failed to create new category", "err", err)
				return errors.New("failed to save category")
			}
			id, err := res.LastInsertId()
			if err != nil {
				log.Error("failed to get last insert id", "err", err)
				return errors.New("something went wrong")
			}
			model.Id = id
		} else {
			_, err := m.db.NamedExec(updateCategoryQuery, &model)
			if err != nil {
				log.Error("failed to update category", "err", err)
				return errors.New("failed to save category")
			}
		}

		return SubmitData(model)
	case entity.Password:
		model.Name = m.fields[0].Value()
		model.Email = m.fields[1].Value()
		model.Username = m.fields[2].Value()
		model.Password = m.fields[3].Value()
		if err := model.Validate(); err != nil {
			return err
		}

		// Encrypt our password
		publicKey := cache.Get[*rsa.PublicKey](cache.PublicKey)
		enc := model.Copy()
		if err := enc.Encrypt(publicKey); err != nil {
			log.Error("(*prompt).submit: failed to encrypt password entity",
				"err", err)
			return err
		}

		if model.Id == 0 {
			res, err := m.db.NamedExec(insertPasswordQuery, &enc)
			if err != nil {
				log.Error("failed to create new password", "err", err)
				return errors.New("failed to save password")
			}
			id, err := res.LastInsertId()
			if err != nil {
				log.Error("failed to get last insert id", "err", err)
				return errors.New("something went wrong")
			}
			model.Id = id
		} else {
			_, err := m.db.NamedExec(updatePasswordQuery, &enc)
			if err != nil {
				log.Error("failed to update password", "err", err)
				return errors.New("failed to save password")
			}
		}

		return SubmitData(model)
	}

	return nil
}
