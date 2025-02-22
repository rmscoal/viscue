package shelf

import (
	"sort"

	"viscue/tui/component/table"
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"

	"github.com/sahilm/fuzzy"
	"github.com/samber/lo"
)

func (m *Model) filter() {
	value := m.search.Value()
	ranks := fuzzy.Find(value,
		lo.Map(m.passwords,
			func(password entity.Password, _ int) string {
				return password.Name
			},
		),
	)
	sort.Stable(ranks)
	indexes := lo.Map(ranks, func(match fuzzy.Match, _ int) int {
		return match.Index
	})
	rows := lo.FilterMap(m.passwords,
		func(password entity.Password, index int) (table.Row, bool) {
			return password.ToTableRow(), lo.Contains(
				indexes,
				index,
			)
		},
	)
	m.table.SetRows(rows)
}

func (m *Model) sync() {
	switch m.selectedCategoryId {
	case 0:
		lo.Map(m.passwords,
			func(password entity.Password, index int) table.Row {
				return password.ToTableRow()
			})
	case -1:
		lo.FilterMap(m.passwords,
			func(password entity.Password, index int) (table.Row, bool) {
				return password.ToTableRow(), !password.CategoryId.Valid
			})
	default:
		lo.FilterMap(m.passwords,
			func(password entity.Password, index int) (table.Row, bool) {
				return password.ToTableRow(),
					password.CategoryId.Int64 == m.selectedCategoryId &&
						password.CategoryId.Valid
			})
	}
}

func (m *Model) calculateDimension() {
	appHeight := style.CalculateAppHeight() - 2
	appWidth := cache.Get[int](cache.TerminalWidth) - 6
	shelfWidth := appWidth * 60 / 100
	paneWidth := shelfWidth + 4
	columnWidth := (shelfWidth - 8) / 3
	m.table.SetHeight(appHeight - 8)
	m.table.SetWidth(shelfWidth)
	m.table.SetColumnsWidth(0, 0, columnWidth, columnWidth, columnWidth, 0)
	m.search.Width = shelfWidth - 11
	m.paneBorder = m.paneBorder.Height(appHeight).
		MaxHeight(appHeight + 2).
		Width(paneWidth)
}
