package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Align place le texte d'une ligne dans son cadre.
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// alignedLine est une ligne dans un contour. Elle ne prend pas le focus.
// title vide laisse le trait haut nu. Le texte trop long est coupé.
type alignedLine struct {
	text     string
	align    Align
	title    string
	disabled bool
}

// Frame aligne text dans un cadre. title est le mot posé sur le trait, vide s'il n'y en a pas.
func Frame(text string, align Align, title string) *alignedLine {
	return &alignedLine{text: text, align: align, title: title}
}

// SetText remplace le texte. L'alignement et le titre du trait restent.
func (f *alignedLine) SetText(text string) {
	if f == nil || f.text == text {
		return
	}
	f.text = text
	notify()
}

// SetTitle remplace le mot posé sur le trait. Une chaîne vide laisse le trait nu.
// Le texte et l'alignement restent.
func (f *alignedLine) SetTitle(title string) {
	if f == nil || f.title == title {
		return
	}
	f.title = title
	notify()
}

func (f *alignedLine) Active() bool { return false }

func (f *alignedLine) SetActive(bool) {}

func (f *alignedLine) Disable() {
	if f == nil || f.disabled {
		return
	}
	f.disabled = true
	notify()
}

func (f *alignedLine) Enable() {
	if f == nil || !f.disabled {
		return
	}
	f.disabled = false
	notify()
}

func (f *alignedLine) Disabled() bool { return f != nil && f.disabled }

func (f *alignedLine) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (f *alignedLine) View(s Style, _ Widget, _ bool) string {
	if f.Disabled() {
		s = s.dimmed()
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	return s.box(alignText(s, f.text, f.align, s.content(width)), frameIdle, width, s.Height, f.title)
}

func alignText(s Style, text string, align Align, width int) string {
	if width < 1 {
		width = 1
	}
	shown := clipWidth(text, width)
	gap := width - lipgloss.Width(shown)
	if gap < 0 {
		gap = 0
	}
	left := 0
	switch align {
	case AlignRight:
		left = gap
	case AlignCenter:
		left = gap / 2
	}
	return s.paint(s.Text, false, strings.Repeat(" ", left)+shown+strings.Repeat(" ", gap-left))
}

// statusLine est une valeur alignée à droite, dans un cadre sans titre.
// label est recopié tel quel, espaces compris.
type statusLine struct {
	label    string
	text     string
	disabled bool
}

// Status aligne label et la valeur à droite.
func Status(label, text string) *statusLine {
	return &statusLine{label: label, text: text}
}

// SetText remplace la valeur. label reste.
func (n *statusLine) SetText(text string) {
	if n == nil || n.text == text {
		return
	}
	n.text = text
	notify()
}

func (n *statusLine) Active() bool { return false }

func (n *statusLine) SetActive(bool) {}

func (n *statusLine) Disable() {
	if n == nil || n.disabled {
		return
	}
	n.disabled = true
	notify()
}

func (n *statusLine) Enable() {
	if n == nil || !n.disabled {
		return
	}
	n.disabled = false
	notify()
}

func (n *statusLine) Disabled() bool { return n != nil && n.disabled }

func (n *statusLine) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (n *statusLine) View(s Style, _ Widget, _ bool) string {
	if n.Disabled() {
		s = s.dimmed()
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	return s.box(alignText(s, n.label+n.text, AlignRight, s.content(width)), frameIdle, width, s.Height, "")
}

// caption est un texte sans cadre. Il ne prend pas le focus.
type caption struct {
	text     string
	cols     int
	disabled bool
}

// Label dessine text sur cols colonnes. 0 suit le texte.
// Le texte trop long est coupé. Les colonnes manquantes sont des espaces.
func Label(text string, cols int) *caption {
	if cols < 0 {
		cols = 0
	}
	return &caption{text: text, cols: cols}
}

// SetText remplace le texte. La largeur reste.
func (c *caption) SetText(text string) {
	if c == nil || c.text == text {
		return
	}
	c.text = text
	notify()
}

// SetColumns fixe le nombre de colonnes. 0 suit le texte.
func (c *caption) SetColumns(cols int) {
	if c == nil {
		return
	}
	if cols < 0 {
		cols = 0
	}
	if c.cols == cols {
		return
	}
	c.cols = cols
	notify()
}

func (c *caption) Active() bool { return false }

func (c *caption) SetActive(bool) {}

func (c *caption) Disable() {
	if c == nil || c.disabled {
		return
	}
	c.disabled = true
	notify()
}

func (c *caption) Enable() {
	if c == nil || !c.disabled {
		return
	}
	c.disabled = false
	notify()
}

func (c *caption) Disabled() bool { return c != nil && c.disabled }

func (c *caption) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (c *caption) View(s Style, _ Widget, _ bool) string {
	if c == nil {
		return ""
	}
	if c.disabled {
		s = s.dimmed()
	}
	width := lipgloss.Width(c.text)
	if c.cols > 0 {
		width = c.cols
	}
	if width < 1 {
		width = 1
	}
	extra := 0
	if s.fit && s.Width > width {
		extra = s.Width - width
	} else if s.fit && s.Width > 0 && s.Width < width {
		width = s.Width
	}
	line := s.line(c.text, s.Text, false, width)
	if extra <= 0 {
		return line
	}
	h := 1
	if s.Height > 1 {
		h = s.Height
	}
	if h == 1 {
		return line + s.shade(extra)
	}
	return padRight(line, s.veil(extra, h))
}
