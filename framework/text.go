package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// InputKind filtre les caractères d'un champ.
type InputKind int

const (
	InputText InputKind = iota
	InputNumber
	InputSecret
)

// TextInput est un champ d'une ligne. Le label est à gauche, la saisie à sa droite.
// limit borne le nombre de caractères. Zéro ne borne pas. Échap referme le champ.
type textInput struct {
	gate
	title    string
	value    []rune
	origin   string
	caret    int
	kind     InputKind
	limit    int
	editable bool
}

func TextInput(title string, kind InputKind, limit int) *textInput {
	return NewTextInput(title, kind, limit)
}

func NewTextInput(title string, kind InputKind, limit int) *textInput {
	return &textInput{gate: gate{on: true}, title: title, kind: kind, limit: limit, editable: true}
}

// Value est la valeur réactive du champ.
func (t *textInput) Value() string { return string(t.value) }

func (t *textInput) SetValue(v string) {
	t.value = []rune(v)
	t.caret = len(t.value)
}

func (t *textInput) SetEditable(on bool) { t.editable = on }

func (t *textInput) Reload() {
	t.SetValue(t.origin)
}

func (t *textInput) Handle(inside bool, key string) (bool, tea.Cmd) {
	if !inside {
		if key == "enter" {
			return true, nil
		}
		return false, nil
	}
	switch key {
	case "enter", "esc":
		return false, nil
	case "left":
		if t.caret > 0 {
			t.caret--
		}
	case "right":
		if t.caret < len(t.value) {
			t.caret++
		}
	case "backspace":
		if t.editable && t.caret > 0 {
			t.value = append(t.value[:t.caret-1], t.value[t.caret:]...)
			t.caret--
		}
	default:
		r, ok := printable(key)
		if ok && t.editable && t.accept(r) && (t.limit <= 0 || len(t.value) < t.limit) {
			next := make([]rune, 0, len(t.value)+1)
			next = append(next, t.value[:t.caret]...)
			next = append(next, r)
			next = append(next, t.value[t.caret:]...)
			t.value = next
			t.caret++
		}
	}
	return true, nil
}

func (t *textInput) accept(r rune) bool {
	if t.kind == InputNumber {
		return r >= '0' && r <= '9'
	}
	return true
}

func (t *textInput) View(s Style, focused Widget, inside bool) string {
	inner := s.content(s.Width)
	label := t.title
	labelW := lipgloss.Width(label)
	gap := 0
	if labelW > 0 {
		gap = 1
	}
	fieldW := inner - labelW - gap
	if fieldW < 1 {
		fieldW = 1
		room := inner - fieldW - gap
		if room < 0 {
			room = 0
		}
		label = clipWidth(label, room)
		labelW = lipgloss.Width(label)
		fieldW = inner - labelW - gap
		if fieldW < 1 {
			fieldW = 1
		}
	}
	shown := t.Value()
	if t.kind == InputSecret {
		shown = strings.Repeat("•", len(t.value))
	}
	if inside && focused == t {
		shown = caretAt(shown, t.caret)
	}
	line := ""
	if labelW > 0 {
		line = s.line(label, s.Accent, true, labelW) + s.paint(s.Text, false, " ")
	}
	line += s.line(shown, s.Text, false, fieldW)
	return s.box(line, focusedEdge(t, focused, inside), s.Width, s.Height, "")
}

// TextArea est un bloc de texte. rows est le nombre de lignes visibles.
// Au-delà, une barre apparaît. Éditable, Entrée va à la ligne et Échap referme.
// Sinon, Entrée ouvre, haut et bas défilent, Entrée referme.
type textArea struct {
	gate
	title    string
	lines    []string
	origin   []string
	editable bool
	rows     int
	line     int
	col      int
	top      int
}

func TextArea(title, text string, editable bool, rows int) *textArea {
	return NewTextArea(title, text, editable, rows)
}

func NewTextArea(title, text string, editable bool, rows int) *textArea {
	if rows < 1 {
		rows = 1
	}
	lines := splitText(text)
	return &textArea{
		gate:     gate{on: true},
		title:    title,
		lines:    lines,
		origin:   append([]string(nil), lines...),
		editable: editable,
		rows:     rows,
	}
}

func (t *textArea) Text() string { return strings.Join(t.lines, "\n") }

func (t *textArea) SetText(text string) {
	t.lines = splitText(text)
	t.line, t.col, t.top = 0, 0, 0
	t.clamp()
}

func (t *textArea) Top() int { return t.top }

func (t *textArea) Reload() {
	t.lines = append([]string(nil), t.origin...)
	t.line, t.col, t.top = 0, 0, 0
}

func (t *textArea) Handle(inside bool, key string) (bool, tea.Cmd) {
	if !inside {
		if key == "enter" {
			return true, nil
		}
		return false, nil
	}
	if !t.editable {
		switch key {
		case "enter", "esc":
			return false, nil
		case "up":
			t.line--
		case "down":
			t.line++
		}
		t.clamp()
		t.follow()
		return true, nil
	}
	switch key {
	case "esc":
		return false, nil
	case "enter":
		t.splitLine()
	case "up":
		t.line--
	case "down":
		t.line++
	case "left":
		t.step(-1)
	case "right":
		t.step(1)
	case "backspace":
		t.backspace()
	default:
		if r, ok := printable(key); ok {
			t.insert(r)
		}
	}
	t.clamp()
	t.follow()
	return true, nil
}

func (t *textArea) visible(s Style) int {
	rows := t.rows
	if rows < 1 {
		rows = 1
	}
	if s.Height <= 0 {
		return rows
	}
	title := 0
	if t.title != "" {
		title = 1
	}
	avail := s.Height - 2 - 2*s.padY() - title
	if avail > rows {
		return avail
	}
	return rows
}

func (t *textArea) View(s Style, focused Widget, inside bool) string {
	inner := s.content(s.Width)
	open := inside && focused == t
	visible := t.visible(s)
	n := len(t.lines)
	scroll := n > visible
	textW := inner
	if scroll {
		textW = inner - 1
	}
	if textW < 1 {
		textW = 1
	}
	if open {
		t.follow()
	}
	top := t.top
	if n <= visible {
		top = 0
	} else if top > n-visible {
		top = n - visible
	}
	if top < 0 {
		top = 0
	}
	lines := make([]string, 0, visible+1)
	if t.title != "" {
		lines = append(lines, s.line(t.title, s.Accent, true, inner))
	}
	thumb := 0
	if scroll {
		span := n - visible
		if span < 1 {
			span = 1
		}
		thumb = top * (visible - 1) / span
	}
	for i := 0; i < visible; i++ {
		src := ""
		at := top + i
		if at >= 0 && at < n {
			src = t.lines[at]
		}
		if open && t.editable && at == t.line {
			src = caretAt(src, t.col)
		}
		row := fit(s.paint(s.Text, false, clipWidth(src, textW)), textW, s)
		if scroll {
			glyph := s.Track
			ink := s.Idle
			if i == thumb {
				glyph = s.Thumb
				ink = s.Text
			}
			row += s.paint(ink, false, glyph)
		}
		lines = append(lines, row)
	}
	return s.box(strings.Join(lines, "\n"), focusedEdge(t, focused, inside), s.Width, s.Height, "")
}

func (t *textArea) clamp() {
	if len(t.lines) == 0 {
		t.lines = []string{""}
	}
	if t.line < 0 {
		t.line = 0
	}
	if t.line >= len(t.lines) {
		t.line = len(t.lines) - 1
	}
	row := []rune(t.lines[t.line])
	if t.col < 0 {
		t.col = 0
	}
	if t.col > len(row) {
		t.col = len(row)
	}
}

func (t *textArea) follow() {
	if t.line < t.top {
		t.top = t.line
	}
	if t.line >= t.top+t.rows {
		t.top = t.line - t.rows + 1
	}
	maxTop := len(t.lines) - t.rows
	if maxTop < 0 {
		maxTop = 0
	}
	if t.top > maxTop {
		t.top = maxTop
	}
	if t.top < 0 {
		t.top = 0
	}
}

func (t *textArea) insert(r rune) {
	t.clamp()
	row := []rune(t.lines[t.line])
	next := make([]rune, 0, len(row)+1)
	next = append(next, row[:t.col]...)
	next = append(next, r)
	next = append(next, row[t.col:]...)
	t.lines[t.line] = string(next)
	t.col++
}

func (t *textArea) splitLine() {
	t.clamp()
	row := []rune(t.lines[t.line])
	left, right := string(row[:t.col]), string(row[t.col:])
	next := make([]string, 0, len(t.lines)+1)
	next = append(next, t.lines[:t.line]...)
	next = append(next, left, right)
	next = append(next, t.lines[t.line+1:]...)
	t.lines = next
	t.line++
	t.col = 0
}

func (t *textArea) backspace() {
	t.clamp()
	if t.col > 0 {
		row := []rune(t.lines[t.line])
		row = append(row[:t.col-1], row[t.col:]...)
		t.lines[t.line] = string(row)
		t.col--
		return
	}
	if t.line == 0 {
		return
	}
	prev := t.lines[t.line-1]
	t.col = len([]rune(prev))
	t.lines[t.line-1] = prev + t.lines[t.line]
	t.lines = append(t.lines[:t.line], t.lines[t.line+1:]...)
	t.line--
}

func (t *textArea) step(delta int) {
	t.clamp()
	if delta < 0 {
		if t.col > 0 {
			t.col--
			return
		}
		if t.line > 0 {
			t.line--
			t.col = len([]rune(t.lines[t.line]))
		}
		return
	}
	if t.col < len([]rune(t.lines[t.line])) {
		t.col++
		return
	}
	if t.line < len(t.lines)-1 {
		t.line++
		t.col = 0
	}
}

func splitText(text string) []string {
	if text == "" {
		return []string{""}
	}
	return strings.Split(text, "\n")
}

func printable(key string) (rune, bool) {
	if key == "space" {
		return ' ', true
	}
	r := []rune(key)
	if len(r) != 1 || r[0] < 0x20 {
		return 0, false
	}
	return r[0], true
}

func caretAt(text string, col int) string {
	r := []rune(text)
	if col < 0 {
		col = 0
	}
	if col > len(r) {
		col = len(r)
	}
	return string(r[:col]) + "▏" + string(r[col:])
}
