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
	paneWidth := sidebarWidth + 4
	m.list.SetHeight(appHeight - 8)
	m.list.SetWidth(sidebarWidth)
	m.search.Width = sidebarWidth - 11
	m.paneBorder = m.paneBorder.Height(appHeight).
		MaxHeight(appHeight + 2).
		Width(paneWidth)
}

func (m *Model) append(payload entity.Category) {
	_, index, found := lo.FindIndexOf(m.categories,
		func(item entity.Category) bool {
			return item.Id == payload.Id
		})
	if !found {
		all := m.categories[0]
		uncategorized := m.categories[len(m.categories)-1]
		m.categories = m.categories[1 : len(m.categories)-1]
		m.categories = append(m.categories, payload)
		sort.Slice(m.categories, func(i, j int) bool {
			return m.categories[i].Name < m.categories[j].Name
		})
		m.categories = append([]entity.Category{all}, m.categories...)
		m.categories = append(m.categories, uncategorized)
	} else {
		m.categories[index] = payload
	}
	m.list.SetItems(
		lo.Map(m.categories,
			func(category entity.Category, _ int) list.Item {
				return category
			},
		),
	)
}
