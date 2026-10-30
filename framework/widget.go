// Package framework réduit une interface Bubble Tea à des objets courts.
// Style porte tout le dessin. Grid place les objets par ligne et déplace le focus.
// Title, Category, ListUniq, ListMult, Item, Button, Action, TextArea et TextInput
// se branchent sur le même contrat.
//
// Les flèches passent d'une zone active à l'autre. Le contour passe au bleu.
// Entrée ouvre une liste ou un champ, et le contour passe au bleu clair.
// Dans une liste unique, Entrée choisit et referme. Dans une liste multiple,
// Entrée coche et Échap referme. Le focus reste.
// Ctrl+gauche et Ctrl+droite changent d'onglet depuis n'importe où.
// Sans nouvel onglet, la touche ne fait rien. La barre n'est pas un focus.
// Le contour des onglets et du panneau reste gris.
// L'objet de droite prend le vide horizontal. La hauteur rejoint l'objet d'en dessous.
package framework

import tea "charm.land/bubbletea/v2"

// Widget est une case de la grille.
// Handle reçoit la touche. Il renvoie vrai si la zone reste ouverte.
type Widget interface {
	Active() bool
	SetActive(bool)
	Handle(inside bool, key string) (bool, tea.Cmd)
	View(s Style, focused Widget, inside bool) string
}

// Reloader réinitialise une zone. L'action Reload s'en sert.
type Reloader interface{ Reload() }

// gate est le drapeau partagé : inactif, la case est sautée.
type gate struct{ on bool }

func (g *gate) Active() bool { return g.on }

func (g *gate) SetActive(on bool) { g.on = on }

type kids interface{ Kids() []Widget }

func depth(w Widget) int {
	if row, ok := w.(*tabRow); ok {
		n := 1
		for _, cell := range row.cells {
			if cell == nil {
				continue
			}
			if d := depth(cell); d > n {
				n = d
			}
		}
		return n
	}
	group, ok := w.(kids)
	if !ok {
		return 1
	}
	n := 0
	for _, kid := range group.Kids() {
		n += depth(kid)
	}
	if n == 0 {
		return 1
	}
	return n
}

// lineAt retourne les objets de la bande sub.
// Une rangée d'onglet en occupe plusieurs, côte à côte.
func lineAt(w Widget, sub int) []Widget {
	if w == nil {
		return []Widget{nil}
	}
	if row, ok := w.(*tabRow); ok {
		out := make([]Widget, 0, len(row.cells))
		for _, cell := range row.cells {
			out = append(out, lineAt(cell, sub)...)
		}
		if len(out) == 0 {
			return []Widget{nil}
		}
		return out
	}
	group, ok := w.(kids)
	if !ok {
		if sub == 0 {
			return []Widget{w}
		}
		return []Widget{nil}
	}
	used := 0
	for _, kid := range group.Kids() {
		d := depth(kid)
		if sub < used+d {
			return lineAt(kid, sub-used)
		}
		used += d
	}
	return []Widget{nil}
}
