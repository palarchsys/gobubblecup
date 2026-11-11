package framework

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Button déclenche une action. Entrée ne l'ouvre pas.
type button struct {
	gate
	label  string
	action Action
	fitted bool
}

func Button(label string, action Action) *button {
	return &button{gate: gate{on: true}, label: label, action: action}
}

// Fit serre le cadre sur le mot : bordures, padding, puis les runes.
// Le vide que la grille ajoute n'étire plus ce bouton. La hauteur suit le contenu.
func (b *button) Fit() {
	if b == nil || b.fitted {
		return
	}
	b.fitted = true
	notify()
}

// SetLabel remplace le mot. L'ajustement et l'action restent.
func (b *button) SetLabel(label string) {
	if b == nil || b.label == label {
		return
	}
	b.label = label
	notify()
}

func (b *button) Handle(inside bool, key string) (bool, tea.Cmd) {
	if key != "enter" || b.action == nil {
		return false, nil
	}
	_ = inside
	return false, b.action.Act()
}

func (b *button) View(s Style, focused Widget, inside bool) string {
	if b.Disabled() {
		s = s.dimmed()
	}
	width := s.Width
	height := s.Height
	if b.fitted {
		width = b.natural(s)
		height = 0
	}
	if width < 1 {
		width = 1
	}
	return s.box(s.line(b.label, s.Text, false, s.content(width)), focusedEdge(b, focused, inside), width, height, "")
}

func (b *button) natural(s Style) int {
	label := lipgloss.Width(b.label)
	if label < 1 {
		label = 1
	}
	width := label + 2 + 2*s.padX()
	minW := 2 + 2*s.padX() + 1
	if width < minW {
		return minW
	}
	return width
}

// buttonRow est une rangée de mots encadrés. Un seul est en couleur.
// Elle ne prend pas le focus. Au départ le premier item est choisi.
type buttonRow struct {
	items    []Item
	buttons  []*button
	index    int
	disabled bool
}

// Buttons aligne un cadre par item. Le mot dessiné est le label.
// Le premier est en couleur, les autres en gris.
// Un espace d'une colonne sépare les cadres. Le vide d'une rangée plus large va à gauche.
func Buttons(items ...Item) *buttonRow {
	copied := append([]Item(nil), items...)
	row := &buttonRow{items: copied, index: -1}
	if len(copied) > 0 {
		row.index = 0
	}
	row.buttons = make([]*button, len(copied))
	for i, item := range copied {
		btn := Button(item.Label, nil)
		btn.SetActive(false)
		btn.Fit()
		if i != row.index {
			btn.Disable()
		}
		row.buttons[i] = btn
	}
	return row
}

// Select colore cet id et grise les autres. Un id absent ne change rien.
func (r *buttonRow) Select(id string) {
	if r == nil {
		return
	}
	found := -1
	for i, item := range r.items {
		if item.ID == id {
			found = i
			break
		}
	}
	if found < 0 || found == r.index {
		return
	}
	r.apply(found)
}

// Selected rend l'id coloré. Vide tant qu'aucun n'est choisi.
func (r *buttonRow) Selected() string {
	if r == nil || r.index < 0 || r.index >= len(r.items) {
		return ""
	}
	return r.items[r.index].ID
}

// Next avance d'un cran, boucle, et rend le nouvel id.
func (r *buttonRow) Next() string {
	if r == nil || len(r.items) == 0 {
		return ""
	}
	next := r.index + 1
	if next < 0 || next >= len(r.items) {
		next = 0
	}
	r.apply(next)
	return r.items[r.index].ID
}

// IDs rend une copie des id, dans l'ordre des boutons.
func (r *buttonRow) IDs() []string {
	if r == nil {
		return nil
	}
	out := make([]string, len(r.items))
	for i, item := range r.items {
		out[i] = item.ID
	}
	return out
}

// SetItems remplace les items dans l'ordre. Le choix reste sur le même index.
// Une longueur différente arrête le programme.
func (r *buttonRow) SetItems(items []Item) {
	if r == nil {
		return
	}
	if len(items) != len(r.items) {
		panic("framework: boutons")
	}
	copied := append([]Item(nil), items...)
	r.items = copied
	for i, item := range copied {
		r.buttons[i].SetLabel(item.Label)
	}
}

func (r *buttonRow) apply(index int) {
	r.index = index
	for i, btn := range r.buttons {
		if i == index {
			btn.Enable()
		} else {
			btn.Disable()
		}
	}
}

func (r *buttonRow) Active() bool { return false }

func (r *buttonRow) SetActive(bool) {}

func (r *buttonRow) Disable() {
	if r == nil || r.disabled {
		return
	}
	r.disabled = true
	notify()
}

func (r *buttonRow) Enable() {
	if r == nil || !r.disabled {
		return
	}
	r.disabled = false
	notify()
}

func (r *buttonRow) Disabled() bool { return r != nil && r.disabled }

func (r *buttonRow) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (r *buttonRow) View(s Style, focused Widget, inside bool) string {
	if r == nil || len(r.buttons) == 0 {
		return ""
	}
	if r.Disabled() {
		s = s.dimmed()
	}
	child := s
	if s.fit {
		child = s.preferred()
	}
	parts := make([]string, 0, len(r.buttons)*2)
	for i, btn := range r.buttons {
		if i > 0 {
			parts = append(parts, child.shade(1))
		}
		parts = append(parts, btn.View(child, focused, inside))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	if s.fit && s.Width > lineWidth(body) {
		body = padLeft(s, body, s.Width-lineWidth(body))
	}
	return body
}
