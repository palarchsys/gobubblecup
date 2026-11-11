package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// focusLink publie l'id focalisé. La valeur se lit comme tout Set(clé, texte).
type focusLink struct {
	store Store
	key   string
	id    string
	set   bool
}

// items est le tronc commun des deux listes : titre, lignes, curseur.
type items struct {
	gate
	title  string
	list   []Item
	origin []Item
	cursor int
	top    int
	link   *focusLink
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

func (it *items) focusID() string {
	if it == nil || len(it.list) == 0 || it.cursor < 0 || it.cursor >= len(it.list) {
		return ""
	}
	return it.list[it.cursor].ID
}

func (it *items) setFocusID(id string) bool {
	if id == "" {
		return false
	}
	for i, item := range it.list {
		if item.ID == id {
			it.cursor = i
			return true
		}
	}
	return false
}

func (it *items) setDot(id, color string) bool {
	if id == "" {
		return false
	}
	for i := range it.list {
		if it.list[i].ID == id {
			it.list[i].Dot = color
			return true
		}
	}
	return false
}

// Bind publie l'id focalisé dans la mémoire, tout de suite, puis à chaque déplacement.
// Entrée ne publie pas : Value et Choices restent le choix confirmé.
// Une mémoire en lecture seule panique à l'écriture.
func (it *items) Bind(store Store, key string) {
	if it == nil || store == nil || key == "" {
		panic("framework: lien de focus incomplet")
	}
	it.link = &focusLink{store: store, key: key}
	it.publish()
}

// publish écrit l'id focalisé. Le même id ne change pas la génération.
func (it *items) publish() {
	if it == nil || it.link == nil {
		return
	}
	id := it.focusID()
	if it.link.set && it.link.id == id {
		return
	}
	it.link.store.Set(it.link.key, id)
	it.link.id = id
	it.link.set = true
}

// handle déplace l'item ou confirme le choix. La liste n'est pas une zone à ouvrir.
// Haut et bas publient le focus. Entrée ne publie pas.
func (it *items) handle(key string, choose func()) tea.Cmd {
	switch key {
	case "up":
		it.move(-1)
		it.publish()
	case "down":
		it.move(1)
		it.publish()
	case "enter":
		if it.cursor >= 0 && it.cursor < len(it.list) {
			choose()
		}
	}
	return nil
}

// minBox est le cadre le plus court : bordures, titre s'il existe, une ligne.
func (it *items) minBox(s Style) int {
	n := 2 + 2*s.padY() + 1
	if it != nil && it.title != "" {
		n++
	}
	return n
}

// itemSlots est le nombre de lignes d'items qui tiennent dans la hauteur.
// Height à zéro : toutes les lignes.
func (it *items) itemSlots(s Style) int {
	n := len(it.list)
	if s.Height <= 0 || n == 0 {
		return n
	}
	inner := s.Height - 2 - 2*s.padY()
	if it.title != "" {
		inner--
	}
	if inner < 1 {
		inner = 1
	}
	if inner > n {
		return n
	}
	return inner
}

// reveal garde le curseur dans la fenêtre. Haut et bas n'ont bougé que le curseur.
func (it *items) reveal(visible int) {
	n := len(it.list)
	if visible <= 0 || visible >= n {
		it.top = 0
		return
	}
	if it.cursor < it.top {
		it.top = it.cursor
	}
	if it.cursor >= it.top+visible {
		it.top = it.cursor - visible + 1
	}
	if it.top < 0 {
		it.top = 0
	}
	if maxTop := n - visible; it.top > maxTop {
		it.top = maxTop
	}
}

func (it *items) body(s Style, armed bool, picked func(int) bool) string {
	inner := s.content(s.Width)
	visible := it.itemSlots(s)
	start, end := 0, len(it.list)
	if s.Height > 0 && visible < len(it.list) {
		it.reveal(visible)
		start = it.top
		end = start + visible
		if end > len(it.list) {
			end = len(it.list)
		}
	}
	lines := make([]string, 0, end-start+1)
	if it.title != "" {
		lines = append(lines, s.line(it.title, s.Accent, true, inner))
	}
	for i := start; i < end; i++ {
		item := it.list[i]
		remembered := i == it.cursor
		lines = append(lines, s.itemLine(item.Label, remembered, armed && remembered, picked(i), item.Dot, inner))
	}
	if len(lines) == 0 {
		lines = []string{fit("", inner, s)}
	}
	return strings.Join(lines, "\n")
}

// ListUniq garde un seul item choisi.
// Les arguments sont des items fixes, ou une seule source, pas les deux.
// Value renvoie l'id confirmé par Entrée. Il est vide tant qu'aucun item n'est choisi.
// Focus renvoie l'id de l'item focalisé. Au départ, c'est le premier.
type itemSource interface{ Items() []Item }

type listUniq struct {
	items
	picked int
	src    itemSource
}

func ListUniq(title string, args ...ListArg) *listUniq {
	var list []Item
	var src itemSource
	for _, arg := range args {
		switch v := arg.(type) {
		case Item:
			list = append(list, v)
		case *reactiveObject[Item]:
			if v == nil {
				panic("framework: argument de liste inattendu")
			}
			if src != nil {
				panic("framework: une seule source de liste")
			}
			src = v
		case *reactiveWatcher:
			if v == nil {
				panic("framework: argument de liste inattendu")
			}
			if src != nil {
				panic("framework: une seule source de liste")
			}
			src = v
		default:
			panic("framework: argument de liste inattendu")
		}
	}
	if src != nil && len(list) > 0 {
		panic("framework: items fixes et réactifs")
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
		if zone, ok := l.src.(Reloader); ok {
			zone.Reload()
		}
		l.picked = -1
		l.cursor = 0
		l.list = nil
		l.sync()
		l.publish()
		return
	}
	l.items.Reload()
	l.picked = -1
	l.publish()
}

func (l *listUniq) Value() string {
	l.sync()
	if l.picked < 0 || l.picked >= len(l.list) {
		return ""
	}
	return l.list[l.picked].ID
}

// Focus est l'id de l'item focalisé. Il est vide si la liste n'a aucune ligne.
func (l *listUniq) Focus() string {
	l.sync()
	return l.focusID()
}

// SetFocus place le curseur sur cet id. Un id absent ne le déplace pas.
func (l *listUniq) SetFocus(id string) {
	l.sync()
	moved := l.setFocusID(id)
	l.publish()
	if moved {
		notify()
	}
}

// SetDot change la couleur de la pastille. Vide la masque.
// Sur une liste réactive qui mappe "dot", l'écriture passe par la mémoire.
func (l *listUniq) SetDot(id, color string) {
	if src, ok := l.src.(*reactiveObject[Item]); ok && src.writeDot(id, color) {
		l.sync()
		l.publish()
		return
	}
	if l.setDot(id, color) {
		notify()
	}
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
	_ = inside
	l.sync()
	l.publish()
	return false, l.handle(key, func() { l.picked = l.cursor })
}

func (l *listUniq) View(s Style, focused Widget, inside bool) string {
	_ = inside
	if l.Disabled() {
		s = s.dimmed()
	}
	l.sync()
	l.publish()
	body := l.body(s, focused == l, func(i int) bool { return i == l.picked })
	return s.box(body, focusedEdge(l, focused, false), s.Width, s.Height, "")
}

// ListMult garde plusieurs items.
// Entrée coche ou décoche l'item focalisé. Le focus de grille reste sur la liste.
type listMult struct {
	items
	picked []bool
}

func ListMult(title string, list ...Item) *listMult {
	return &listMult{items: newItems(title, list), picked: make([]bool, len(list))}
}

func (l *listMult) Reload() {
	l.items.Reload()
	l.picked = make([]bool, len(l.list))
	l.publish()
}

func (l *listMult) Cursor() int { return l.cursor }

// Focus est l'id de l'item focalisé. Il est vide si la liste n'a aucune ligne.
func (l *listMult) Focus() string { return l.focusID() }

// SetFocus place le curseur sur cet id. Un id absent ne le déplace pas.
func (l *listMult) SetFocus(id string) {
	moved := l.setFocusID(id)
	l.publish()
	if moved {
		notify()
	}
}

// SetDot change la couleur de la pastille. Vide la masque.
func (l *listMult) SetDot(id, color string) {
	if l.setDot(id, color) {
		notify()
	}
}

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
	_ = inside
	return false, l.handle(key, func() {
		if l.cursor >= len(l.picked) {
			next := make([]bool, len(l.list))
			copy(next, l.picked)
			l.picked = next
		}
		l.picked[l.cursor] = !l.picked[l.cursor]
	})
}

func (l *listMult) View(s Style, focused Widget, inside bool) string {
	_ = inside
	if l.Disabled() {
		s = s.dimmed()
	}
	body := l.body(s, focused == l, func(i int) bool {
		return i < len(l.picked) && l.picked[i]
	})
	return s.box(body, focusedEdge(l, focused, false), s.Width, s.Height, "")
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
