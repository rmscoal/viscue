package sidebar

import (
	"viscue/tui/entity"
	"viscue/tui/views/library/message"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type DataLoadedMsg struct {
	Data []entity.Category
}

func (m Model) LoadItems() tea.Msg {
	rows, err := m.db.Queryx(`
		WITH sorter AS (
			SELECT 0 AS id, 'All' AS name, 1 AS sort_order
			UNION ALL
			SELECT id, name, 2 AS sort_order FROM categories
			UNION ALL
			SELECT -1 AS id, 'Uncategorized' AS name, 3 AS sort_order
			ORDER BY sort_order, id
		)
		SELECT id, name FROM sorter
	`)
	if err != nil {
		return err
	}

	var categories []entity.Category
	for rows.Next() {
		var category entity.Category
		if err := rows.StructScan(&category); err != nil {
			return err
		}

		categories = append(categories, category)
	}

	_ = rows.Close()
	log.Debug("(Model).LoadItems: retrieved categories", "items", categories)
	return DataLoadedMsg{
		Data: categories,
	}
}

func (m Model) AddCategoryPromptMsg() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Category]{
				Payload: entity.Category{},
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}

func (m Model) EditCategoryPromptMsg() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Category]{
				Payload: m.list.SelectedItem().(entity.Category),
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}

func (m Model) CategorySelectedMsg() tea.Msg {
	category := m.list.SelectedItem().(entity.Category)
	return message.CategorySelectedMsg(category.Id)
}

func (m Model) DeleteCategoryPromptMsg() tea.Cmd {
	category := m.list.SelectedItem().(entity.Category)
	if category.Id == 0 || category.Id == -1 {
		// TODO: Show notification, cannot delete...
		return nil
	}
	return tea.Sequence(
		func() tea.Msg {
			return message.OpenPromptMsg[entity.Category]{
				Payload:    category,
				IsDeletion: true,
			}
		},
		func() tea.Msg {
			return message.PromptFocused
		},
	)
}
