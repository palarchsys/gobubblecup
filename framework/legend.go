package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// legend liste les contrôles en bas de l'interface.
type legend struct{ gate }

// Legend construit cette liste. Elle ne prend pas le focus.
func Legend() *legend { return &legend{} }

func (l *legend) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (l *legend) View(s Style, _ Widget, _ bool) string {
	return strings.Join(legendLines(s), "\n")
}

func legendLines(s Style) []string {
	rows := [][]struct{ key, text string }{
		{{"←→", "focus"}, {"↑↓", "item"}, {"entrée", "choisit"}, {"tab", "onglet"}},
		{{"échap", "sort"}, {"ctrl+v", "colle"}, {"ctrl+c", "quitte"}},
	}
	out := make([]string, len(rows))
	for i, row := range rows {
		parts := make([]string, len(row))
		for j, item := range row {
			parts[j] = s.paint(s.Accent, true, item.key) + s.paint(s.Text, false, " "+item.text)
		}
		out[i] = strings.Join(parts, s.paint(s.Text, false, "   "))
	}
	return out
}
