package shelf

import (
	"crypto/rsa"
	"database/sql"

	"viscue/tui/component/notification"
	"viscue/tui/entity"
	"viscue/tui/tool/cache"
	"viscue/tui/views/library/message"

	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
)

type DataLoadedMsg struct {
	Data []entity.Password
}

func (m Model) LoadItems() tea.Msg {
	rows, err := m.db.Queryx(
		"SELECT id, category_id, name, email, username, password FROM passwords",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	privateKey := cache.Get[*rsa.PrivateKey](cache.PrivateKey)
	var passwords []entity.Password
	for rows.Next() {
		var password entity.Password
		err = rows.StructScan(&password)
		if err != nil {
			return err
		}
		err = password.Decrypt(privateKey)
		if err != nil {
			return err
		}
		passwords = append(passwords, password)
	}

	return DataLoadedMsg{
		Data: passwords,
	}
}

func (m Model) EditPasswordPromptMsg() tea.Cmd {
	row := m.table.SelectedRow()
	password, _ := entity.NewPasswordFromTableRow(row)
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Password]{
				Payload: password,
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}

func (m Model) AddPasswordPromptMsg() tea.Cmd {
	selectedCategoryId := m.selectedCategoryId
	if selectedCategoryId <= 0 {
		selectedCategoryId = 0
	}
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Password]{
				Payload: entity.Password{
					CategoryId: sql.NullInt64{
						Int64: selectedCategoryId,
						Valid: selectedCategoryId > 0,
					},
				},
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}

func (m Model) DeletePasswordPromptMsg() tea.Cmd {
	row := m.table.SelectedRow()
	password, _ := entity.NewPasswordFromTableRow(row)
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Password]{
				Payload:    password,
				IsDeletion: true,
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}

func (m Model) CopyToClipboard() tea.Msg {
	row := m.table.SelectedRow()
	password, _ := entity.NewPasswordFromTableRow(row)
	clipboard.Write(clipboard.FmtText, []byte(password.Password))
	return notification.ShowMsg{
		Message: "Password is copied to clipboard",
	}
}
