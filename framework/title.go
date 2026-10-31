package framework

import tea "charm.land/bubbletea/v2"

// Title est le titre de l'application. Il ne prend pas le focus.
// La grille l'étire sur la largeur de l'écran.
type title struct {
	Text     string
	disabled bool
}

func NewTitle(text string) *title { return &title{Text: text} }

func Title(text string) *title { return NewTitle(text) }

func (t *title) Active() bool { return false }

func (t *title) SetActive(bool) {}

func (t *title) Disable() {
	if t == nil || t.disabled {
		return
	}
	t.disabled = true
	notify()
}

func (t *title) Enable() {
	if t == nil || !t.disabled {
		return
	}
	t.disabled = false
	notify()
}

func (t *title) Disabled() bool { return t != nil && t.disabled }

func (t *title) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (t *title) View(s Style, _ Widget, _ bool) string {
	if t.Disabled() {
		s = s.dimmed()
	}
	return t.Banner(s, s.Width)
}

func (t *title) Banner(s Style, width int) string {
	inner := s.content(width)
	return s.box(s.line(t.Text, s.Accent, true, inner), frameIdle, width, s.Height, "")
}
