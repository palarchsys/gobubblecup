package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// items est le tronc commun des deux listes : titre, lignes, curseur.
type items struct {
	gate
	title  string
	list   []Item
	origin []Item
	cursor int
}

func newItems(title string, list []Item) items {
	origin := append([]Item(nil), list...)
	return items{gate: gate{on: true}, title: title, list: append([]Item(nil), list...), origin: origin}
}

func (it *items) Reload() {
	it.list = append([]Item(nil), it.origin...)
	it.cursor = 0
}

func (it *items) Title() string { return it.title }

func (it *items) move(delta int) {
	next := it.cursor + delta
	if next < 0 || next >= len(it.list) {
		return
	}
	it.cursor = next
}

func (it *items) handle(inside bool, key string, choose func()) (bool, tea.Cmd) {
	if !inside {
		if key == "enter" {
			return true, nil
		}
		return false, nil
	}
	switch key {
	case "up":
		it.move(-1)
	case "down":
		it.move(1)
	case "enter":
		if it.cursor >= 0 && it.cursor < len(it.list) {
			choose()
		}
		return false, nil
	}
	return true, nil
}

func (it *items) body(s Style, inside bool, picked func(int) bool) string {
	inner := s.content(s.Width)
	lines := make([]string, 0, len(it.list)+1)
	if it.title != "" {
		lines = append(lines, s.line(it.title, s.Accent, true, inner))
	}
	for i, item := range it.list {
		lines = append(lines, s.itemLine(item.Label, inside && i == it.cursor, picked(i), inner))
	}
	if len(lines) == 0 {
		lines = []string{fit("", inner, s)}
	}
	return strings.Join(lines, "\n")
}

// ListUniq garde un seul item choisi.
// Value renvoie l'id confirmé par Entrée. Il est vide tant qu'aucun item n'est choisi.
type itemSource interface{ Items() []Item }

type listUniq struct {
	items
	picked int
	src    itemSource
}

func ListUniq(title string, args ...ListArg) *listUniq { return NewListUniq(title, args...) }

func NewListUniq(title string, args ...ListArg) *listUniq {
	var list []Item
	var src itemSource
	for _, arg := range args {
		switch v := arg.(type) {
		case Item:
			list = append(list, v)
		case *reactiveObject:
			src = v
		case *reactiveWatcher:
			src = v
		default:
			panic("framework: argument de liste inattendu")
		}
	}
	l := &listUniq{items: newItems(title, list), picked: -1, src: src}
	l.sync()
	return l
}

func (l *listUniq) sync() {
	if l.src == nil {
		return
	}
	next := l.src.Items()
	pickedID := ""
	if l.picked >= 0 && l.picked < len(l.list) {
		pickedID = l.list[l.picked].ID
	}
	cursorID := ""
	if l.cursor >= 0 && l.cursor < len(l.list) {
		cursorID = l.list[l.cursor].ID
	}
	l.list = next
	l.picked = -1
	l.cursor = 0
	for i, item := range next {
		if pickedID != "" && item.ID == pickedID {
			l.picked = i
		}
		if cursorID != "" && item.ID == cursorID {
			l.cursor = i
		}
	}
	if len(next) == 0 {
		l.cursor = 0
		return
	}
	if l.cursor >= len(next) {
		l.cursor = len(next) - 1
	}
}

func (l *listUniq) Reload() {
	if l.src != nil {
		l.picked = -1
		l.cursor = 0
		l.list = nil
		l.sync()
		return
	}
	l.items.Reload()
	l.picked = -1
}

func (l *listUniq) Value() string {
	l.sync()
	if l.picked < 0 || l.picked >= len(l.list) {
		return ""
	}
	return l.list[l.picked].ID
}

func (l *listUniq) Cursor() int { return l.cursor }

func (l *listUniq) Picked() int { return l.picked }

func (l *listUniq) Choice() (Item, bool) {
	if l.picked < 0 || l.picked >= len(l.list) {
		return Item{}, false
	}
	return l.list[l.picked], true
}

func (l *listUniq) Handle(inside bool, key string) (bool, tea.Cmd) {
	l.sync()
	return l.handle(inside, key, func() { l.picked = l.cursor })
}

func (l *listUniq) View(s Style, focused Widget, inside bool) string {
	l.sync()
	body := l.body(s, inside && focused == l, func(i int) bool { return i == l.picked })
	return s.box(body, focusedEdge(l, focused, inside), s.Width, s.Height, "")
}

// ListMult garde plusieurs items.
// Entrée ajoute ou retire l'item courant et reste dans la liste. Échap sort.
type listMult struct {
	items
	picked []bool
}

func ListMult(title string, list ...Item) *listMult { return NewListMult(title, list...) }

func NewListMult(title string, list ...Item) *listMult {
	return &listMult{items: newItems(title, list), picked: make([]bool, len(list))}
}

func (l *listMult) Reload() {
	l.items.Reload()
	l.picked = make([]bool, len(l.list))
}

func (l *listMult) Cursor() int { return l.cursor }

func (l *listMult) Choices() []Item {
	out := make([]Item, 0, len(l.picked))
	for i, on := range l.picked {
		if on && i < len(l.list) {
			out = append(out, l.list[i])
		}
	}
	return out
}

// Values renvoie les id des items cochés.
func (l *listMult) Values() []string {
	out := make([]string, 0)
	for i, on := range l.picked {
		if on && i < len(l.list) {
			out = append(out, l.list[i].ID)
		}
	}
	return out
}

func (l *listMult) Handle(inside bool, key string) (bool, tea.Cmd) {
	if !inside {
		if key == "enter" {
			return true, nil
		}
		return false, nil
	}
	switch key {
	case "up":
		l.move(-1)
	case "down":
		l.move(1)
	case "esc":
		return false, nil
	case "enter":
		if l.cursor >= 0 && l.cursor < len(l.list) {
			if l.cursor >= len(l.picked) {
				next := make([]bool, len(l.list))
				copy(next, l.picked)
				l.picked = next
			}
			l.picked[l.cursor] = !l.picked[l.cursor]
		}
	}
	return true, nil
}

func (l *listMult) View(s Style, focused Widget, inside bool) string {
	body := l.body(s, inside && focused == l, func(i int) bool {
		return i < len(l.picked) && l.picked[i]
	})
	return s.box(body, focusedEdge(l, focused, inside), s.Width, s.Height, "")
}

func lineWidth(view string) int {
	line, _, _ := strings.Cut(view, "\n")
	return lipgloss.Width(line)
}

func lineCount(view string) int {
	if view == "" {
		return 0
	}
	return strings.Count(view, "\n") + 1
}
