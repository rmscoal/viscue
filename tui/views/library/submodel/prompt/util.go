package prompt

import (
	"viscue/tui/entity"
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
)

func (m *Model) cycleFocus(msg tea.KeyMsg) {
	if m.isDeletion {
		return
	}

	length := len(m.fields) // Including button
	if _, idx, found := lo.FindIndexOf(m.fields,
		func(item textinput.Model) bool {
			return item.Focused()
		},
	); found {
		m.fields[idx].Blur()
	}
	if msg.String() == "tab" {
		m.pointer++
	} else {
		m.pointer--
	}
	switch {
	case m.pointer > length:
		m.pointer = 0
		fallthrough
	case m.pointer < length && m.pointer >= 0:
		m.fields[m.pointer].Focus()
		m.blurSubmitButton()
	case m.pointer < 0:
		m.pointer = length
		fallthrough
	case m.pointer == length:
		m.fields[m.pointer-1].Blur()
		m.focusSubmitButton()
	}
}

func (m *Model) togglePasswordVisibility() {
	if m.payload == nil {
		return
	} else if _, ok := m.payload.(entity.Password); !ok {
		return
	}

	m.showPassword = !m.showPassword
	if m.showPassword {
		m.fields[4].EchoMode = textinput.EchoNormal
	} else {
		m.fields[4].EchoMode = textinput.EchoPassword
	}
}

func (m Model) updateTextInputs(msg tea.Msg) (Model, tea.Cmd) {
	var commands []tea.Cmd
	var cmd tea.Cmd
	for i := range m.fields {
		m.fields[i], cmd = m.fields[i].Update(msg)
		commands = append(commands, cmd)
	}
	return m, tea.Batch(commands...)
}

func (m *Model) getCategories() error {
	query := `
		WITH results AS (
			SELECT 0 AS id, 'All' AS name, 1 AS sort_order
			UNION ALL
			SELECT id, name, 2 AS sort_order FROM categories
			UNION ALL
			SELECT -1 AS id, 'Uncategorized' AS name, 3 AS sort_order
			ORDER BY sort_order, name
		)
		SELECT id, name FROM results
	`
	rows, err := m.db.Queryx(query)
	if err != nil {
		log.Error("prompt.Model.getCategories: failed to get categories",
			"err", err)
		return err
	}

	var categories []entity.Category
	for rows.Next() {
		var category entity.Category
		err = rows.StructScan(&category)
		if err != nil {
			log.Error("prompt.Model.getCategories: failed scanning category struct",
				"err", err)
			return err
		}
		categories = append(categories, category)
	}
	_ = rows.Close()
	m.categories = categories
	return nil
}

func (m *Model) focusSubmitButton() {
	m.button = m.button.
		UnsetForeground().
		UnsetBackground().
		Foreground(style.ActiveButtonStyle.GetForeground()).
		Background(style.ActiveButtonStyle.GetBackground())
}

func (m *Model) blurSubmitButton() {
	m.button = m.button.
		UnsetForeground().
		UnsetBackground().
		Foreground(style.ButtonStyle.GetForeground()).
		Background(style.ButtonStyle.GetBackground())
}

func (m Model) isButtonFocused() bool {
	return m.pointer == len(m.fields)
}
