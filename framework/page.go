package framework

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Gap est une ligne vide. Elle ne prend pas le focus.
func Gap() Widget { return gap{} }

type gap struct{}

func (gap) Active() bool { return false }

func (gap) SetActive(bool) {}

func (gap) Disable() {}

func (gap) Enable() {}

func (gap) Disabled() bool { return false }

func (gap) Handle(bool, string) (bool, Cmd) { return false, nil }

func (gap) View(s Style, _ Widget, _ bool) string {
	n := s.Width
	if n < 1 {
		n = 1
	}
	return s.shade(n)
}

// Note est une puce d'information. Elle ne prend pas le focus.
func Note(text string) *note { return &note{text: text} }

type note struct{ text string }

// SetText remplace la phrase.
func (n *note) SetText(text string) {
	if n == nil || n.text == text {
		return
	}
	n.text = text
	notify()
}

func (n *note) Active() bool { return false }

func (n *note) SetActive(bool) {}

func (n *note) Disable() {}

func (n *note) Enable() {}

func (n *note) Disabled() bool { return false }

func (n *note) Handle(bool, string) (bool, Cmd) { return false, nil }

func (n *note) View(s Style, _ Widget, _ bool) string {
	if n == nil {
		return ""
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	body := clipWidth(n.text, width-2)
	return fit(s.paint(s.Focus, false, "•")+s.paint(s.Text, false, " "+body), width, s)
}

// Fact est une puce dont le texte est relu à chaque dessin.
// Une fonction nulle ou une chaîne vide affiche —.
type Fact struct {
	Label string
	Text  func() string
}

func (f Fact) value() string {
	if f.Text == nil {
		return "—"
	}
	text := f.Text()
	if text == "" {
		return "—"
	}
	return text
}

// Facts aligne des puces sous un titre. Elles ne prennent pas le focus.
func Facts(rows ...Fact) *facts {
	return &facts{rows: append([]Fact(nil), rows...)}
}

type facts struct{ rows []Fact }

func (f *facts) Active() bool { return false }

func (f *facts) SetActive(bool) {}

func (f *facts) Disable() {}

func (f *facts) Enable() {}

func (f *facts) Disabled() bool { return false }

func (f *facts) Handle(bool, string) (bool, Cmd) { return false, nil }

func (f *facts) View(s Style, _ Widget, _ bool) string {
	width := s.Width
	if width < 1 {
		width = 1
	}
	if f == nil || len(f.rows) == 0 {
		return s.shade(width)
	}
	max := 0
	for _, row := range f.rows {
		if n := lipgloss.Width(row.Label); n > max {
			max = n
		}
	}
	lines := make([]string, len(f.rows))
	for i, row := range f.rows {
		gap := max - lipgloss.Width(row.Label)
		body := row.Label + strings.Repeat(" ", gap) + " : " + row.value()
		line := s.paint(s.Text, false, "  ") + s.paint(s.Focus, false, "•") + s.paint(s.Text, false, " "+body)
		lines[i] = fit(line, width, s)
	}
	return strings.Join(lines, "\n")
}

// Bullet est une puce qui prend le focus. Entrée lance l'action. La ligne ne s'ouvre pas.
func Bullet(label string, action Action) *bullet {
	return &bullet{gate: gate{on: true}, label: label, action: action}
}

type bullet struct {
	gate
	label  string
	action Action
}

// SetLabel remplace le mot. L'action reste.
func (b *bullet) SetLabel(label string) {
	if b == nil || b.label == label {
		return
	}
	b.label = label
	notify()
}

func (b *bullet) Handle(_ bool, key string) (bool, Cmd) {
	if b == nil || b.disabled || key != "enter" || b.action == nil {
		return false, nil
	}
	return false, b.action.Act()
}

func (b *bullet) View(s Style, focused Widget, _ bool) string {
	if b == nil {
		return ""
	}
	if b.Disabled() {
		s = s.dimmed()
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	color := s.Text
	bold := false
	if focused == b && !b.Disabled() {
		color = s.Accent
		bold = true
	}
	line := s.paint(s.Focus, false, "•") + s.paint(color, bold, " "+b.label)
	return fit(line, width, s)
}

// Moves ajoute Suite et Retour après les autres lignes.
// Une action nulle est omise.
func Moves(next, prev Action, kids ...Widget) []Widget {
	out := append([]Widget{}, kids...)
	if next != nil {
		out = append(out, Bullet("Suite", next))
	}
	if prev != nil {
		out = append(out, Bullet("Retour", prev))
	}
	return out
}

// Sheet empile des objets, sans cadre.
func Sheet(kids ...Widget) *sheet { return newSheet(false, "", kids) }

// Section encadre ses enfants. legend est le mot du trait haut.
func Section(legend string, kids ...Widget) *sheet { return newSheet(true, legend, kids) }

// Page est un écran : titre, puces, puis un corps.
// facts peut être nil.
func Page(title string, facts, body Widget) *sheet {
	if body == nil {
		panic("framework: page vide")
	}
	kids := []Widget{Title(title), Gap()}
	if facts != nil {
		kids = append(kids, facts, Gap())
	}
	kids = append(kids, body)
	return Sheet(kids...)
}

type sheet struct {
	legend   string
	boxed    bool
	kids     []Widget
	disabled bool
	held     []Widget
}

func newSheet(boxed bool, legend string, kids []Widget) *sheet {
	if len(kids) == 0 {
		panic("framework: section vide")
	}
	for _, kid := range kids {
		if kid == nil {
			panic("framework: widget nul")
		}
	}
	return &sheet{boxed: boxed, legend: legend, kids: append([]Widget(nil), kids...)}
}

// SetLegend remplace le mot du trait. Les enfants restent.
func (s *sheet) SetLegend(legend string) {
	if s == nil || !s.boxed || s.legend == legend {
		return
	}
	s.legend = legend
	notify()
}

func (s *sheet) Kids() []Widget {
	if s == nil {
		return nil
	}
	return s.kids
}

func (s *sheet) Active() bool {
	if s == nil || s.disabled {
		return false
	}
	for _, kid := range s.kids {
		if kid.Active() {
			return true
		}
	}
	return false
}

func (s *sheet) SetActive(on bool) {
	if s == nil {
		return
	}
	for _, kid := range s.kids {
		kid.SetActive(on)
	}
}

func (s *sheet) Disable() {
	if s == nil || s.disabled {
		return
	}
	s.disabled = true
	s.held = nil
	for _, kid := range s.kids {
		if kid.Disabled() {
			continue
		}
		kid.Disable()
		s.held = append(s.held, kid)
	}
	notify()
}

func (s *sheet) Enable() {
	if s == nil || !s.disabled {
		return
	}
	s.disabled = false
	for _, kid := range s.held {
		kid.Enable()
	}
	s.held = nil
	notify()
}

func (s *sheet) Disabled() bool { return s != nil && s.disabled }

func (s *sheet) Handle(bool, string) (bool, Cmd) { return false, nil }

func (s *sheet) Reload() {
	if s == nil {
		return
	}
	for _, kid := range s.kids {
		if zone, ok := kid.(Reloader); ok {
			zone.Reload()
		}
	}
}

func (s *sheet) View(style Style, focused Widget, inside bool) string {
	if s == nil {
		return ""
	}
	if s.disabled {
		style = style.dimmed()
	}
	if !s.boxed {
		child := style
		child.Height = 0
		parts := make([]string, len(s.kids))
		for i, kid := range s.kids {
			parts[i] = kid.View(child, focused, inside)
		}
		return strings.Join(parts, "\n")
	}
	width := style.Width
	if width < 12 {
		width = 12
	}
	inner := width - 6
	if inner < 1 {
		inner = 1
	}
	child := style
	child.Width = inner
	child.Height = 0
	parts := make([]string, len(s.kids))
	for i, kid := range s.kids {
		parts[i] = kid.View(child, focused, inside)
	}
	return sectionFrame(style, s.legend, strings.Join(parts, "\n"), width, style.Height)
}

func sectionFrame(s Style, legend, body string, width, height int) string {
	tl, tr, bl, br, h, v := s.corners()
	rows := strings.Split(body, "\n")
	if body == "" {
		rows = []string{""}
	}
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, s.paint(s.Idle, false, topRule(tl, tr, h, legend, width)))
	for _, row := range rows {
		lines = append(lines, sideLine(s, v, row, width))
	}
	for height > len(lines)+1 {
		lines = append(lines, sideLine(s, v, "", width))
	}
	lines = append(lines, s.paint(s.Idle, false, bottomRule(bl, br, h, width)))
	return strings.Join(lines, "\n")
}

func topRule(left, right, dash, legend string, width int) string {
	legend = clipWidth(legend, width-6)
	head := left + "─ " + legend + " "
	dashes := width - lipgloss.Width(head) - lipgloss.Width(right)
	if dashes < 1 {
		dashes = 1
	}
	return head + strings.Repeat(dash, dashes) + right
}

func bottomRule(left, right, dash string, width int) string {
	if width < 2 {
		width = 2
	}
	return left + strings.Repeat(dash, width-2) + right
}

func sideLine(s Style, bar, content string, width int) string {
	inner := width - 6
	if inner < 1 {
		inner = 1
	}
	gap := inner - lipgloss.Width(content)
	if gap < 0 {
		gap = 0
	}
	return s.paint(s.Idle, false, bar) +
		s.shade(2) +
		content +
		s.shade(gap+2) +
		s.paint(s.Idle, false, bar)
}
