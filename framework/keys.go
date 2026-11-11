package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// keyLegend est la légende de l'écran, en deux colonnes. Elle ne prend pas le focus.
// Les phrases viennent de l'appelant.
type keyLegend struct {
	title      string
	left       [][2]string
	right      [][2]string
	disabled   bool
	slack      bool
	phraseCols int
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

// Use pose le titre et les deux colonnes depuis un catalogue.
// Les touches affichées viennent de l'appelant.
// Une langue absente, un identifiant absent ou deux longueurs différentes arrêtent le programme.
// Une clé de titre vide laisse le trait nu.
func (k *keyLegend) Use(texts map[string]map[string]string, lang, titleKey string, leftKeys, leftIDs, rightKeys, rightIDs []string) {
	if k == nil {
		return
	}
	table, ok := texts[lang]
	if !ok {
		panic("framework: langue " + lang)
	}
	title := ""
	if titleKey != "" {
		var found bool
		title, found = table[titleKey]
		if !found {
			panic("framework: identifiant " + titleKey)
		}
	}
	left := pairsFrom(table, leftKeys, leftIDs)
	right := pairsFrom(table, rightKeys, rightIDs)
	k.title = title
	k.left = left
	k.right = right
	notify()
}

func pairsFrom(table map[string]string, keys, ids []string) [][2]string {
	if len(keys) != len(ids) {
		panic("framework: paires")
	}
	out := make([][2]string, len(keys))
	for i := range keys {
		text, ok := table[ids[i]]
		if !ok {
			panic("framework: identifiant " + ids[i])
		}
		out[i] = [2]string{keys[i], text}
	}
	return out
}

// Take donne à ce cadre le surplus de sa ligne.
// Le cadre s'étend. Les colonnes gardent la largeur de leur texte.
// Le vide reste à l'intérieur, à droite des phrases.
// Sans Take, le surplus va au dernier objet de la ligne.
func (k *keyLegend) Take() *keyLegend {
	if k == nil || k.slack {
		return k
	}
	k.slack = true
	notify()
	return k
}

func (k *keyLegend) TakesSlack() bool { return k != nil && k.slack }

// Column fige la largeur de chaque colonne de phrases.
// 0 suit le texte de la langue affichée. Une phrase plus courte est complétée.
// Une phrase plus longue est montrée en entier. Le catalogue n'est pas réécrit.
func (k *keyLegend) Column(n int) *keyLegend {
	if k == nil {
		return nil
	}
	if n < 0 {
		n = 0
	}
	if k.phraseCols == n {
		return k
	}
	k.phraseCols = n
	notify()
	return k
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
	body := lipgloss.JoinHorizontal(lipgloss.Top, paintPairs(s, k.left, k.phraseCols), s.shade(2), paintPairs(s, k.right, k.phraseCols))
	natural := lineWidth(body) + 2 + 2*s.padX()
	if natural < 1 {
		natural = 1
	}
	width := natural
	if s.fit && s.Width > natural {
		width = s.Width
	}
	return s.box(body, frameIdle, width, s.Height, k.title)
}

func paintPairs(s Style, pairs [][2]string, cols int) string {
	if len(pairs) == 0 {
		return ""
	}
	width := 0
	shown := make([]string, len(pairs))
	for i, pair := range pairs {
		shown[i] = padCols(pair[1], cols)
		n := lipgloss.Width(pair[0] + " : " + shown[i])
		if n > width {
			width = n
		}
	}
	lines := make([]string, len(pairs))
	for i, pair := range pairs {
		line := s.paint(s.Accent, true, pair[0]) + s.paint(s.Text, false, " : "+shown[i])
		if gap := width - lipgloss.Width(line); gap > 0 {
			line += s.shade(gap)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func padCols(text string, cols int) string {
	if cols <= 0 {
		return text
	}
	gap := cols - lipgloss.Width(text)
	if gap <= 0 {
		return text
	}
	return text + strings.Repeat(" ", gap)
}
