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
// limit borne le nombre de caractères. Zéro ne borne pas.
// border dessine le cadre. Sans cadre, les champs empilés se touchent.
// Le focus suffit pour écrire : Entrée n'ouvre pas le champ.
type textInput struct {
	gate
	title     string
	value     []rune
	origin    string
	caret     int
	kind      InputKind
	limit     int
	editable  bool
	border    bool
	overwrite bool
}

func TextInput(title string, kind InputKind, limit int, border bool) *textInput {
	return NewTextInput(title, kind, limit, border)
}

func NewTextInput(title string, kind InputKind, limit int, border bool) *textInput {
	return &textInput{gate: gate{on: true}, title: title, kind: kind, limit: limit, editable: true, border: border}
}

// Value est la valeur réactive du champ.
func (t *textInput) Value() string { return string(t.value) }

func (t *textInput) SetValue(v string) {
	t.value = []rune(v)
	t.caret = len(t.value)
}

func (t *textInput) SetText(text string) { t.SetValue(text) }

// Paste insère le texte du presse-papier. Les chiffres seuls passent dans un champ nombre.
func (t *textInput) Paste(text string) {
	if t == nil || !t.editable {
		return
	}
	for _, r := range text {
		if r == '\n' || r == '\r' || r < 0x20 {
			continue
		}
		t.insert(r)
	}
}

func (t *textInput) insert(r rune) {
	if !t.accept(r) {
		return
	}
	if t.overwrite && t.caret < len(t.value) {
		t.value[t.caret] = r
		t.caret++
		return
	}
	if t.limit > 0 && len(t.value) >= t.limit {
		return
	}
	next := make([]rune, 0, len(t.value)+1)
	next = append(next, t.value[:t.caret]...)
	next = append(next, r)
	next = append(next, t.value[t.caret:]...)
	t.value = next
	t.caret++
}

// deleteForward retire le caractère sous le curseur. Le curseur ne bouge pas.
func (t *textInput) deleteForward() {
	if !t.editable || t.caret < 0 || t.caret >= len(t.value) {
		return
	}
	t.value = append(t.value[:t.caret], t.value[t.caret+1:]...)
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
	case "enter", "esc", "up", "down":
		return true, nil
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
	case "delete", "del":
		t.deleteForward()
	case "insert", "ins":
		if t.editable {
			t.overwrite = !t.overwrite
		}
	default:
		if r, ok := printable(key); ok {
			t.insert(r)
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
	if t.Disabled() {
		s = s.dimmed()
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	inner := width
	if t.border {
		inner = s.content(width)
	}
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
	line := ""
	if labelW > 0 {
		line = s.line(label, s.Accent, true, labelW) + s.paint(s.Text, false, " ")
	}
	line += s.caretLine(shown, t.caret, inside && focused == t, fieldW)
	if !t.border {
		if gap := width - lipgloss.Width(line); gap > 0 {
			line += s.shade(gap)
		}
		return line
	}
	return s.box(line, focusedEdge(t, focused, inside), width, s.Height, "")
}

// TextArea est un bloc de texte. rows est le nombre de lignes visibles.
// Au-delà, une barre apparaît. Éditable, Entrée va à la ligne et Échap referme.
// Sinon, Entrée ouvre, haut et bas défilent, Entrée referme.
type textArea struct {
	gate
	title     string
	lines     []string
	origin    []string
	editable  bool
	rows      int
	line      int
	col       int
	top       int
	border    bool
	overwrite bool
	src       Store
	focus     Store
	focusKey  string
	seenID    string
	seenSet   bool
}

func TextArea(title, text string, editable bool, rows int, border bool) *textArea {
	return NewTextArea(title, text, editable, rows, border)
}

// TextView est une zone en lecture. À chaque dessin elle relit le return
// du slot dont l'id est publié dans focus, sous la clé.
// Id vide ou nœud absent : texte vide. La zone n'écrit ni dans src ni dans focus.
func TextView(src, focus Store, key string) *textArea {
	return NewTextView(src, focus, key)
}

func NewTextView(src, focus Store, key string) *textArea {
	if src == nil || focus == nil {
		panic("framework: zone liée incomplète")
	}
	zone := NewTextArea("", "", false, 4, true)
	zone.src = src
	zone.focus = focus
	zone.focusKey = key
	return zone
}

func NewTextArea(title, text string, editable bool, rows int, border bool) *textArea {
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
		border:   border,
	}
}

// Paste insère le texte du presse-papier. Une zone en lecture ne change pas.
func (t *textArea) Paste(text string) {
	if t == nil || !t.editable || text == "" {
		return
	}
	for _, r := range text {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			t.splitLine()
			continue
		}
		if r < 0x20 {
			continue
		}
		t.insert(r)
	}
	t.follow()
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
	case "delete", "del":
		t.deleteForward()
	case "insert", "ins":
		t.overwrite = !t.overwrite
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
	chrome := 2
	if !t.border {
		chrome = 0
	}
	avail := s.Height - chrome - 2*s.padY() - title
	if avail > rows {
		return avail
	}
	return rows
}

// pull relit le slot lié. Un autre id remet le haut à zéro.
// Le même texte, déjà en bas, reste en bas. Une fenêtre remontée ne bouge pas.
func (t *textArea) pull(s Style) {
	if t == nil || t.focus == nil {
		return
	}
	id := t.focus.Get(t.focusKey).Get("data").Get("value").String()
	text := ""
	if id != "" && t.src != nil {
		text = t.src.Get(id).Get("return").String()
	}
	visible := t.visible(s)
	if visible < 1 {
		visible = 1
	}
	if !t.seenSet || id != t.seenID {
		t.lines = splitText(text)
		t.line, t.col, t.top = 0, 0, 0
		t.seenID = id
		t.seenSet = true
		t.clamp()
		return
	}
	if text == t.Text() {
		return
	}
	atBottom := t.top+visible >= len(t.lines)
	t.lines = splitText(text)
	t.clamp()
	maxTop := len(t.lines) - visible
	if maxTop < 0 {
		maxTop = 0
	}
	if atBottom {
		t.top = maxTop
		return
	}
	if t.top > maxTop {
		t.top = maxTop
	}
	if t.top < 0 {
		t.top = 0
	}
}

func (t *textArea) View(s Style, focused Widget, inside bool) string {
	if t.Disabled() {
		s = s.dimmed()
	}
	t.pull(s)
	inner := s.Width
	if t.border {
		inner = s.content(s.Width)
	}
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
		var row string
		if open && t.editable && at == t.line {
			row = s.caretLine(src, t.col, true, textW)
		} else {
			row = s.line(src, s.Text, false, textW)
		}
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
	body := strings.Join(lines, "\n")
	if !t.border {
		return body
	}
	return s.box(body, focusedEdge(t, focused, inside), s.Width, s.Height, "")
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
	if t.overwrite && t.col < len(row) {
		row[t.col] = r
		t.lines[t.line] = string(row)
		t.col++
		return
	}
	next := make([]rune, 0, len(row)+1)
	next = append(next, row[:t.col]...)
	next = append(next, r)
	next = append(next, row[t.col:]...)
	t.lines[t.line] = string(next)
	t.col++
}

// deleteForward retire le caractère sous le curseur.
// En fin de ligne, la ligne suivante est collée.
func (t *textArea) deleteForward() {
	t.clamp()
	row := []rune(t.lines[t.line])
	if t.col < len(row) {
		row = append(row[:t.col], row[t.col+1:]...)
		t.lines[t.line] = string(row)
		return
	}
	if t.line >= len(t.lines)-1 {
		return
	}
	t.lines[t.line] += t.lines[t.line+1]
	t.lines = append(t.lines[:t.line+1], t.lines[t.line+2:]...)
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

// caretLine peint le texte sur width colonnes.
// Le curseur recouvre le caractère : il n'ajoute pas de colonne au milieu.
// En fin de texte, il occupe une colonne après le dernier caractère.
func (s Style) caretLine(text string, caret int, show bool, width int) string {
	if width < 1 {
		width = 1
	}
	runes := []rune(text)
	if caret < 0 {
		caret = 0
	}
	if caret > len(runes) {
		caret = len(runes)
	}
	widthOf := func(r rune) int {
		w := lipgloss.Width(string(r))
		if w < 1 {
			w = 1
		}
		return w
	}
	start := 0
	if show {
		need := 1
		if caret < len(runes) {
			need = widthOf(runes[caret])
		}
		start = caret
		for start > 0 {
			w := widthOf(runes[start-1])
			if need+w > width {
				break
			}
			need += w
			start--
		}
	}
	var b strings.Builder
	used := 0
	for i := start; i < len(runes); i++ {
		w := widthOf(runes[i])
		tail := 0
		if show && caret == len(runes) {
			tail = 1
		}
		if used+w+tail > width {
			break
		}
		glyph := string(runes[i])
		if show && i == caret {
			b.WriteString(s.paintOn(s.paper(), s.Focus, false, glyph))
		} else {
			b.WriteString(s.paint(s.Text, false, glyph))
		}
		used += w
	}
	if show && caret >= len(runes) && used < width {
		b.WriteString(s.paintOn(s.paper(), s.Focus, false, " "))
		used++
	}
	return fit(b.String(), width, s)
}
