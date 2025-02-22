package sidebar

import (
	"sort"

	"viscue/tui/component/list"
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"

	"github.com/sahilm/fuzzy"
	"github.com/samber/lo"
)

func (m *Model) filter() {
	value := m.search.Value()

	ranks := fuzzy.Find(value,
		lo.Map(m.categories,
			func(category entity.Category, _ int) string {
				return category.Name
			},
		),
	)
	sort.Stable(ranks)
	indexes := lo.Map(ranks, func(match fuzzy.Match, _ int) int {
		return match.Index
	})
	items := lo.FilterMap(m.categories,
		func(category entity.Category, index int) (list.Item, bool) {
			return category, lo.Contains(
				indexes,
				index,
			)
		},
	)
	m.list.SetItems(items)
}

func (m *Model) calculateDimension() {
	appHeight := style.CalculateAppHeight() - 2
	appWidth := cache.Get[int](cache.TerminalWidth) - 6
	sidebarWidth := appWidth * 20 / 100
	// listPaneWidth := sidebarWidth + 4
	m.list.SetHeight(appHeight - 8)
	m.list.SetWidth(sidebarWidth)
	m.search.Width = sidebarWidth - 11
	// m.listPaneBorder = m.listPaneBorder.Width(listPaneWidth).
	// 	Height(m.height).
	// 	MaxHeight(m.height + 2)
}
