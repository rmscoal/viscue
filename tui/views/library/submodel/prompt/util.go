package prompt

import (
	"database/sql"
	"errors"
	"strings"

	"viscue/tui/entity"
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/mattn/go-sqlite3"
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
		if m.isPasswordPrompt() && m.pointer == 1 {
			// Disable category text input
			continue
		}
		m.fields[i], cmd = m.fields[i].Update(msg)
		commands = append(commands, cmd)
	}
	return m, tea.Batch(commands...)
}

func (m *Model) getCategories() error {
	query := `
		WITH results AS (
			SELECT 0 AS id, 'None' AS name, 1 AS sort_order
			UNION ALL
			SELECT id, name, 2 AS sort_order FROM categories
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

func (m Model) isPasswordPrompt() bool {
	_, ok := m.payload.(entity.Password)
	return ok
}

func (m Model) textInputWidth() int {
	return min(m.availableWidth-2, minimumTextInputWidth)
}

func (m *Model) setCategoryField(category entity.Category) {
	if category.Id == 0 {
		category.Name = "None"
	}
	categoryName := category.Name
	tiWidth := m.textInputWidth()
	if len(categoryName) >= tiWidth {
		categoryName = categoryName[:tiWidth-4] + "…"
	}
	categoryName = lipgloss.JoinHorizontal(
		lipgloss.Top,
		categoryName,
		strings.Repeat(" ", 2),
		"⌄",
	)
	m.fields[1].SetValue(categoryName)
	m.fields[1].SetCursor(len(categoryName))
	password := m.payload.(entity.Password)
	password.CategoryId = sql.NullInt64{
		Int64: category.Id,
		Valid: category.Id != 0,
	}
	m.payload = password
}

func handleUpsertCategoryError(err error) SubmitError {
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			switch {
			case errors.Is(sqliteErr.Code, sqlite3.ErrConstraint):
				return SubmitError(errors.New("seems like the category name is taken"))
			}
		}
		return SubmitError(err)
	}
	return nil
}

func handleUpsertPasswordError(err error) SubmitError {
	if err == nil {
		return nil
	}

	sqliteErr, ok := err.(sqlite3.Error)
	if ok {
		switch {
		case errors.Is(sqliteErr.Code, sqlite3.ErrConstraint):
			return SubmitError(errors.New("the password's name has been taken in this category"))
		case errors.Is(sqliteErr.Code, sqlite3.ErrIoErr):
			return SubmitError(errors.New("database can't write to disk at the moment"))
		}
	}
	return SubmitError(err)
}
