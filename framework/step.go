package framework

import "strings"

// Mark est la pastille d'une branche.
type Mark int

const (
	Info Mark = iota
	OK
	Warn
	Bad
)

// Line est une branche. Bold met le texte en or.
type Line struct {
	Mark Mark
	Text string
	Bold bool
}

type stepView struct {
	title string
	lines func() []Line
}

// Step dessine un arbre. title est sur la ligne du coin.
// lines est relu à chaque dessin. Une fonction nulle ou vide montre « En attente ».
func Step(title string, lines func() []Line) *stepView {
	return &stepView{title: title, lines: lines}
}

// SetTitle remplace le mot du coin. Les lignes restent.
func (v *stepView) SetTitle(title string) {
	if v == nil || v.title == title {
		return
	}
	v.title = title
	notify()
}

func (v *stepView) Active() bool { return false }

func (v *stepView) SetActive(bool) {}

func (v *stepView) Disable() {}

func (v *stepView) Enable() {}

func (v *stepView) Disabled() bool { return false }

func (v *stepView) Handle(bool, string) (bool, Cmd) { return false, nil }

func (v *stepView) View(s Style, _ Widget, _ bool) string {
	if v == nil {
		return ""
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	var rows []Line
	if v.lines != nil {
		rows = v.lines()
	}
	return stepBlock(s, v.title, rows, width)
}

func stepBlock(s Style, title string, rows []Line, width int) string {
	if width < 1 {
		width = 1
	}
	lines := make([]string, 0, len(rows)+1)
	if title != "" {
		label := clipWidth(title, width-2)
		lines = append(lines, fit(s.paint(s.Text, false, "┌ ")+s.paint(s.Accent, false, label), width, s))
	}
	if len(rows) == 0 {
		rows = []Line{{Mark: Info, Text: "En attente"}}
	}
	for i, row := range rows {
		hook := "├── "
		if i == len(rows)-1 {
			hook = "└── "
		}
		color := s.Text
		if row.Bold {
			color = s.Accent
		}
		used := 6
		body := clipWidth(row.Text, width-used)
		line := s.paint(s.Text, false, hook) + markSpan(s, row.Mark) + s.paint(color, row.Bold, " "+body)
		lines = append(lines, fit(line, width, s))
	}
	return strings.Join(lines, "\n")
}

func markSpan(s Style, kind Mark) string {
	switch kind {
	case OK:
		ink, _ := statusInk("success")
		return s.paint(ink, false, "✓")
	case Warn:
		ink, _ := statusInk("warning")
		return s.paint(ink, false, "⚠")
	case Bad:
		ink, _ := statusInk("error")
		return s.paint(ink, false, "✗")
	default:
		return s.paint(s.Focus, false, "•")
	}
}
