package framework

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Style porte tous les paramètres visuels. Les objets n'en gardent aucun.
// Width est la largeur de cadre d'un objet. Height à zéro suit le contenu.
// La grille remplace les deux le temps d'un dessin pour remplir le vide.
type Style struct {
	Background color.Color
	Surface    color.Color
	Text       color.Color
	Accent     color.Color
	Idle       color.Color
	Focus      color.Color
	Inside     color.Color
	Width      int
	Height     int
	PadX       int
	PadY       int
	Margin     int
	Radius     int
	Thickness  int
	Mark       string
	Track      string
	Thumb      string
	fit        bool
	pref       int
}

type frame int

const (
	frameIdle frame = iota
	frameFocus
	frameInside
)

func edgeOf(focused, inside bool) frame {
	if inside {
		return frameInside
	}
	if focused {
		return frameFocus
	}
	return frameIdle
}

func focusedEdge(self, focused Widget, inside bool) frame {
	if self == nil || self.Disabled() || self != focused {
		return frameIdle
	}
	return edgeOf(true, inside)
}

// dimmed peint textes et titres avec le gris de repos.
func (s Style) dimmed() Style {
	s.Text = s.Idle
	s.Accent = s.Idle
	s.Focus = s.Idle
	s.Inside = s.Idle
	return s
}

// DefaultStyle reprend Dark (Visual Studio), le thème de base de VS Code.
// Fond #1e1e1e, texte #d4d4d4, titres #569cd6.
// Contour au repos #6b6b6b, focus #007acc, édition #9cdcfe.
// L'écran et les zones partagent ce fond : un écart non peint
// laisserait voir le fond du terminal.
// Mesures figées : PadX 1, PadY 0, Margin 1 horizontal.
// Une ligne de padding est trop haute. Le trait du cadre suffit.
// Les rangées et les enfants d'une catégorie se touchent.
func DefaultStyle() Style {
	return Style{
		Background: Color("#1e1e1e"),
		Surface:    Color("#1e1e1e"),
		Text:       Color("#d4d4d4"),
		Accent:     Color("#569cd6"),
		Idle:       Color("#6b6b6b"),
		Focus:      Color("#007acc"),
		Inside:     Color("#9cdcfe"),
		Width:      22,
		PadX:       1,
		PadY:       0,
		Margin:     1,
		Radius:     1,
		Thickness:  1,
		Mark:       "✓",
		Track:      "│",
		Thumb:      "█",
	}
}

// placed fixe le cadre d'un objet pour un dessin. height à zéro reste naturel.
// La largeur d'origine est gardée : une mesure imbriquée ne la confond pas avec le cadre imposé.
func (s Style) placed(width, height int) Style {
	if !s.fit {
		s.pref = s.Width
	}
	if width < 1 {
		width = 1
	}
	s.Width = width
	s.Height = height
	s.fit = true
	return s
}

// preferred retrouve la largeur d'origine, avant le cadre imposé par la grille.
func (s Style) preferred() Style {
	s.Height = 0
	s.fit = false
	if s.pref > 0 {
		s.Width = s.pref
	}
	return s
}

func Color(hex string) color.Color { return lipgloss.Color(hex) }

func (s Style) padX() int {
	if s.PadX < 0 {
		return 0
	}
	return s.PadX
}

func (s Style) padY() int {
	if s.PadY < 0 {
		return 0
	}
	return s.PadY
}

func (s Style) margin() int {
	if s.Margin < 0 {
		return 0
	}
	return s.Margin
}

func (s Style) rounded() bool { return s.Radius > 0 }

func (s Style) content(boxWidth int) int {
	n := boxWidth - 2 - 2*s.padX()
	if n < 1 {
		return 1
	}
	return n
}

func (s Style) line(text string, c color.Color, bold bool, width int) string {
	if width < 1 {
		width = 1
	}
	return fit(s.paint(c, bold, clipWidth(text, width)), width, s)
}

func (s Style) ink(edge frame) (color.Color, bool) {
	switch edge {
	case frameInside:
		return s.Inside, true
	case frameFocus:
		return s.Focus, true
	default:
		return s.Idle, false
	}
}

// paper est la couleur unique de l'écran et des zones.
func (s Style) paper() color.Color {
	if s.Surface != nil {
		return s.Surface
	}
	return s.Background
}

func (s Style) shade(n int) string {
	if n <= 0 {
		return ""
	}
	return s.paint(s.Text, false, strings.Repeat(" ", n))
}

func (s Style) veil(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	row := s.shade(width)
	rows := make([]string, height)
	for i := range rows {
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}

func (s Style) paint(c color.Color, bold bool, text string) string {
	return s.paintOn(c, s.paper(), bold, text)
}

func (s Style) paintOn(fg, bg color.Color, bold bool, text string) string {
	st := lipgloss.NewStyle().Foreground(fg).Background(bg)
	if bold {
		st = st.Bold(true)
	}
	return st.Render(text)
}

func (s Style) corners() (tl, tr, bl, br, h, v string) {
	if s.rounded() {
		return "╭", "╮", "╰", "╯", "─", "│"
	}
	return "┌", "┐", "└", "┘", "─", "│"
}

// itemLine dessine une ligne de liste.
// remembered pose le fond #37373d, que la liste ait le focus ou non.
// armed met le texte en gras : la liste a le focus de grille.
// picked pose ✓ dans la première colonne.
// dot est une couleur #rrggbb. Vide : pas de pastille. Sinon ● en dernière colonne.
func (s Style) itemLine(label string, remembered, armed, picked bool, dot string, width int) string {
	symbol := " "
	if picked {
		symbol = s.Mark
	}
	prefix := symbol + " "
	reserve := 0
	if dot != "" {
		reserve = 1
	}
	room := width - lipgloss.Width(prefix) - reserve
	if room < 0 {
		room = 0
	}
	text := prefix + clipWidth(label, room)
	bg := s.paper()
	if remembered {
		bg = Color("#37373d")
	}
	col := s.Text
	bold := false
	switch {
	case armed && picked:
		col = s.Accent
		bold = true
	case armed:
		col = s.Focus
		bold = true
	case picked:
		col = s.Inside
	}
	line := s.paintOn(col, bg, bold, text)
	pad := width - lipgloss.Width(text) - reserve
	if pad > 0 {
		line += s.paintOn(s.Text, bg, false, strings.Repeat(" ", pad))
	}
	if reserve > 0 {
		line += s.paintOn(Color(dot), bg, false, "●")
	}
	return line
}

// box dessine un contour.
// width et height comptent la bordure et le padding, pas la marge.
// height à zéro prend la hauteur du contenu.
// Le padding horizontal entoure le texte. Les rangées se suivent sans ligne vide.
// legend, s'il est posé, remplace le début du trait haut.
func (s Style) box(body string, edge frame, width, height int, legend string) string {
	ink, bold := s.ink(edge)
	padX, padY := s.padX(), s.padY()
	minW := 2 + 2*padX + 1
	if width < minW {
		width = minW
	}
	inner := s.content(width)
	rows := strings.Split(body, "\n")
	if body == "" {
		rows = []string{""}
	}
	block := make([]string, 0, len(rows)+2*padY)
	for i := 0; i < padY; i++ {
		block = append(block, "")
	}
	block = append(block, rows...)
	contentEnd := len(block)
	for i := 0; i < padY; i++ {
		block = append(block, "")
	}
	if height > 2+len(block) {
		extra := height - 2 - len(block)
		grown := make([]string, 0, len(block)+extra)
		grown = append(grown, block[:contentEnd]...)
		for i := 0; i < extra; i++ {
			grown = append(grown, "")
		}
		grown = append(grown, block[contentEnd:]...)
		block = grown
	}
	var b strings.Builder
	write := func(line string) {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	write(s.topLine(width, legend, ink, bold))
	_, _, _, _, _, v := s.corners()
	side := s.paint(s.Text, false, strings.Repeat(" ", padX))
	bar := s.paint(ink, bold, v)
	empty := s.paint(s.Text, false, strings.Repeat(" ", inner))
	for _, row := range block {
		cell := empty
		if row != "" {
			cell = fit(row, inner, s)
		}
		write(bar + side + cell + side + bar)
	}
	write(s.rule(width, ink, bold, false))
	return strings.TrimSuffix(b.String(), "\n")
}

func (s Style) topLine(width int, legend string, ink color.Color, bold bool) string {
	tl, tr, _, _, h, _ := s.corners()
	inner := width - 2
	if inner < 0 {
		inner = 0
	}
	var mid string
	if legend == "" {
		mid = s.paint(ink, bold, strings.Repeat(h, inner))
	} else {
		label := " " + legend + " "
		if lipgloss.Width(label) > inner {
			label = clipWidth(legend, inner)
		}
		rest := inner - lipgloss.Width(label)
		if rest < 0 {
			rest = 0
		}
		mid = s.paint(s.Accent, true, label) + s.paint(ink, bold, strings.Repeat(h, rest))
	}
	return s.paint(ink, bold, tl) + mid + s.paint(ink, bold, tr)
}

func (s Style) rule(width int, ink color.Color, bold, top bool) string {
	tl, tr, bl, br, h, _ := s.corners()
	left, right := bl, br
	if top {
		left, right = tl, tr
	}
	inner := width - 2
	if inner < 0 {
		inner = 0
	}
	return s.paint(ink, bold, left) + s.paint(ink, bold, strings.Repeat(h, inner)) + s.paint(ink, bold, right)
}

func fit(line string, width int, s Style) string {
	if width < 0 {
		width = 0
	}
	w := lipgloss.Width(line)
	if w > width {
		return line
	}
	if pad := width - w; pad > 0 {
		line += s.paint(s.Text, false, strings.Repeat(" ", pad))
	}
	return line
}

func clipWidth(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if w+rw > n {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

func plain(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			continue
		}
		j := i + 1
		if j < len(s) && s[j] == '[' {
			j++
			for j < len(s) && (s[j] == ';' || (s[j] >= '0' && s[j] <= '9')) {
				j++
			}
			if j < len(s) {
				j++
			}
		}
		i = j - 1
	}
	return b.String()
}
