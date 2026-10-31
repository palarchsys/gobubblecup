package framework

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// variable porte une valeur entière. Elle vaut 0 à la création.
// La barre compte cette valeur quand elle est 0 ou 1.
type variable struct {
	n int
}

func Variable() *variable { return &variable{} }

// Set remplace la valeur et rafraîchit l'écran.
func (v *variable) Set(n int) {
	if v == nil {
		panic("framework: variable nulle")
	}
	v.n = n
	notify()
}

// Value est la valeur courante.
func (v *variable) Value() int {
	if v == nil {
		return 0
	}
	return v.n
}

// variableList est une liste de variables. La barre la relit à chaque dessin.
type variableList struct {
	items []*variable
}

// Variables crée la liste réactive passée à ProgressBar.
func Variables(vars ...*variable) *variableList {
	list := &variableList{}
	list.Set(vars...)
	return list
}

// Set remplace les variables de la liste. La barre suit cette liste.
func (l *variableList) Set(vars ...*variable) {
	if l == nil || len(vars) == 0 {
		panic("framework: liste de variables vide")
	}
	for _, v := range vars {
		if v == nil {
			panic("framework: variable nulle")
		}
	}
	l.items = append([]*variable(nil), vars...)
	notify()
}

// Add ajoute une variable à la fin de la liste.
func (l *variableList) Add(v *variable) {
	if l == nil || v == nil {
		panic("framework: variable nulle")
	}
	l.items = append(l.items, v)
	notify()
}

// progressBar avance selon les variables égales à 0 ou 1.
type progressBar struct {
	gate
	list *variableList
}

func ProgressBar(list *variableList) *progressBar { return NewProgressBar(list) }

func NewProgressBar(list *variableList) *progressBar {
	if list == nil || len(list.items) == 0 {
		panic("framework: barre sans variable")
	}
	return &progressBar{gate: gate{on: false}, list: list}
}

func (p *progressBar) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (p *progressBar) ratio() (done, total, pct int) {
	for _, v := range p.list.items {
		if v == nil || (v.n != 0 && v.n != 1) {
			continue
		}
		total++
		if v.n == 1 {
			done++
		}
	}
	if total > 0 {
		pct = done * 100 / total
	}
	return done, total, pct
}

func (p *progressBar) View(s Style, _ Widget, _ bool) string {
	if p.Disabled() {
		s = s.dimmed()
	}
	_, _, pct := p.ratio()
	width := s.Width
	if width < 1 {
		width = 1
	}
	inner := s.content(width)
	if inner < 1 {
		inner = 1
	}
	filled := pct * inner / 100
	bar := []rune(strings.Repeat("█", filled) + strings.Repeat("░", inner-filled))
	label := []rune(fmt.Sprintf("%d%%", pct))
	start := (inner - len(label)) / 2
	if start < 0 {
		start = 0
	}
	for i, r := range label {
		if start+i >= 0 && start+i < len(bar) {
			bar[start+i] = r
		}
	}
	line := s.line(string(bar), s.Text, true, inner)
	return s.box(line, frameIdle, width, s.Height, "")
}
