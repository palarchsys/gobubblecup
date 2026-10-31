package framework

import tea "charm.land/bubbletea/v2"

// Button déclenche une action. Entrée ne l'ouvre pas.
type button struct {
	gate
	label  string
	action Action
}

func Button(label string, action Action) *button { return NewButton(label, action) }

func NewButton(label string, action Action) *button {
	return &button{gate: gate{on: true}, label: label, action: action}
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
	return s.box(s.line(b.label, s.Text, false, s.content(s.Width)), focusedEdge(b, focused, inside), s.Width, s.Height, "")
}
