package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// keyLegend est la légende de l'écran, en deux colonnes. Elle ne prend pas le focus.
// Les phrases viennent de l'appelant.
type keyLegend struct {
	title    string
	left     [][2]string
	right    [][2]string
	disabled bool
}

// Keys dessine title sur le trait et deux colonnes touche, texte.
// La touche est en accent, le texte en corps.
func Keys(title string, left, right [][2]string) *keyLegend {
	return &keyLegend{title: title, left: copyPairs(left), right: copyPairs(right)}
}

// SetPairs remplace les deux colonnes. Le titre reste.
// L'ordre des paires est l'ordre affiché. Une colonne vide reste valide.
// Les listes sont copiées : l'appelant peut réutiliser ses tampons.
func (k *keyLegend) SetPairs(left, right [][2]string) {
	if k == nil {
		return
	}
	k.left = copyPairs(left)
	k.right = copyPairs(right)
	notify()
}

// SetTitle remplace le mot posé sur le trait. Les colonnes restent.
func (k *keyLegend) SetTitle(title string) {
	if k == nil || k.title == title {
		return
	}
	k.title = title
	notify()
}

func copyPairs(pairs [][2]string) [][2]string {
	if pairs == nil {
		return nil
	}
	out := make([][2]string, len(pairs))
	copy(out, pairs)
	return out
}

func (k *keyLegend) Active() bool { return false }

func (k *keyLegend) SetActive(bool) {}

func (k *keyLegend) Disable() {
	if k == nil || k.disabled {
		return
	}
	k.disabled = true
	notify()
}

func (k *keyLegend) Enable() {
	if k == nil || !k.disabled {
		return
	}
	k.disabled = false
	notify()
}

func (k *keyLegend) Disabled() bool { return k != nil && k.disabled }

func (k *keyLegend) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (k *keyLegend) View(s Style, _ Widget, _ bool) string {
	if k.Disabled() {
		s = s.dimmed()
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, paintPairs(s, k.left), s.shade(2), paintPairs(s, k.right))
	width := s.Width
	if width < 1 {
		width = 1
	}
	natural := lineWidth(body) + 2 + 2*s.padX()
	if width < natural {
		width = natural
	}
	return s.box(body, frameIdle, width, s.Height, k.title)
}

func paintPairs(s Style, pairs [][2]string) string {
	if len(pairs) == 0 {
		return ""
	}
	width := 0
	for _, pair := range pairs {
		n := lipgloss.Width(pair[0] + " : " + pair[1])
		if n > width {
			width = n
		}
	}
	lines := make([]string, len(pairs))
	for i, pair := range pairs {
		line := s.paint(s.Accent, true, pair[0]) + s.paint(s.Text, false, " : "+pair[1])
		if gap := width - lipgloss.Width(line); gap > 0 {
			line += s.shade(gap)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
