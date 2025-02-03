package library

import (
	"crypto/rsa"
	"database/sql"
	"errors"

	"viscue/tui/entity"
	"viscue/tui/tool/cache"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type DataLoadedMsg struct {
	Categories []entity.Category
	Passwords  []entity.Password
}

func (m *library) load() tea.Msg {
	rows, err := m.db.Queryx(`
		WITH results AS (
		    SELECT 0 AS id, 'All' AS name, 1 AS sort_order
			UNION ALL
			SELECT id, name, 2 AS sort_order FROM categories
			UNION ALL
			SELECT -1 AS id, 'Uncategorized' AS name, 3 AS sort_order
			ORDER BY sort_order, name
		)
		SELECT id, name FROM results;
	`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("failed querying categories", "err", err)
		return err
	}

	var categories []entity.Category
	for rows.Next() {
		var category entity.Category
		err = rows.StructScan(&category)
		if err != nil {
			log.Error("failed scanning category struct", "err", err)
			return err
		}
		categories = append(categories, category)
	}
	_ = rows.Close()

	rows, err = m.db.Queryx("SELECT id, category_id, name, email, username, password FROM passwords ORDER BY name")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("failed querying passwords", "err", err)
		return err
	}

	privateKey := cache.Get[*rsa.PrivateKey](cache.PrivateKey)

	var passwords []entity.Password
	for rows.Next() {
		var password entity.Password
		err = rows.StructScan(&password)
		if err != nil {
			log.Error("failed scanning password struct", "err", err)
			return err
		}
		err = password.Decrypt(privateKey)
		if err != nil {
			log.Error("failed to decrpt password entity", "id", password.Id,
				"name", password.Name, "err", err)
		}
		passwords = append(passwords, password)
	}
	_ = rows.Close()

	return DataLoadedMsg{Categories: categories, Passwords: passwords}
}
