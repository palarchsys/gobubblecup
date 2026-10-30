package framework

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Grid définit les lignes et les objets de chaque ligne.
// Le focus est l'objet actif, pas la case qui le contient.
type grid struct {
	rows   [][]Widget
	focus  Widget
	inside bool
}

func Row(widgets ...Widget) []Widget { return widgets }

func Grid(rows ...[]Widget) *grid { return NewGrid(rows...) }

func NewGrid(rows ...[]Widget) *grid {
	for _, row := range rows {
		for _, w := range row {
			if w == nil {
				panic("framework: widget nul dans la grille")
			}
		}
	}
	g := &grid{rows: rows}
	g.focusFirst()
	return g
}

func (g *grid) Focus() Widget { return g.focus }

func (g *grid) Inside() bool { return g.inside }

type spot struct {
	widget Widget
	band   int
	col    int
}

func (g *grid) spots() []spot {
	var out []spot
	band := 0
	for _, row := range g.rows {
		height := 1
		for _, w := range row {
			if n := depth(w); n > height {
				height = n
			}
		}
		for sub := 0; sub < height; sub++ {
			col := 0
			for _, w := range row {
				for _, child := range lineAt(w, sub) {
					if child != nil && child.Active() {
						out = append(out, spot{widget: child, band: band, col: col})
					}
					col++
				}
			}
			band++
		}
	}
	return out
}

func (g *grid) focusFirst() {
	g.inside = false
	if spots := g.spots(); len(spots) > 0 {
		g.focus = spots[0].widget
		return
	}
	g.focus = nil
}

func (g *grid) Key(key string) tea.Cmd {
	if key == "ctrl+left" || key == "ctrl+right" {
		if g.switchTab(key) {
			return nil
		}
	}
	w := g.focus
	if w == nil || !w.Active() || !g.has(w) {
		g.focusFirst()
		w = g.focus
	}
	if w == nil {
		return nil
	}
	if g.inside {
		inside, cmd := w.Handle(true, key)
		g.inside = inside
		return cmd
	}
	if key == "enter" || key == "ctrl+left" || key == "ctrl+right" {
		inside, cmd := w.Handle(false, key)
		g.inside = inside
		return cmd
	}
	g.move(key)
	return nil
}

// switchTab consomme Ctrl+flèche s'il existe une liste d'onglets.
// Sans nouvel onglet dans cette direction, le focus reste où il est.
func (g *grid) switchTab(key string) bool {
	lists := g.tabLists()
	if len(lists) == 0 {
		return false
	}
	if len(lists) > 1 {
		panic("framework: une seule liste d'onglets par interface")
	}
	delta := 1
	if key == "ctrl+left" {
		delta = -1
	}
	if !lists[0].move(delta) {
		return true
	}
	g.inside = false
	if w := lists[0].firstFocus(); w != nil {
		g.focus = w
		return true
	}
	g.focusFirst()
	return true
}

func (g *grid) tabLists() []*tabList {
	var out []*tabList
	seen := map[*tabList]bool{}
	var walk func(Widget)
	walk = func(w Widget) {
		if w == nil {
			return
		}
		if list, ok := w.(*tabList); ok {
			if seen[list] {
				return
			}
			seen[list] = true
			out = append(out, list)
			for _, tab := range list.tabs {
				if tab == nil || tab.grid == nil {
					continue
				}
				for _, row := range tab.grid.rows {
					for _, cell := range row {
						walk(cell)
					}
				}
			}
			return
		}
		group, ok := w.(kids)
		if !ok {
			return
		}
		for _, kid := range group.Kids() {
			walk(kid)
		}
	}
	for _, row := range g.rows {
		for _, w := range row {
			walk(w)
		}
	}
	return out
}

func (g *grid) has(target Widget) bool {
	for _, spot := range g.spots() {
		if spot.widget == target {
			return true
		}
	}
	return false
}

func (g *grid) move(key string) {
	spots := g.spots()
	if len(spots) == 0 {
		return
	}
	here, ok := findSpot(spots, g.focus)
	if !ok {
		g.focus = spots[0].widget
		return
	}
	switch key {
	case "left":
		g.focus = neighborCol(spots, here, -1)
	case "right":
		g.focus = neighborCol(spots, here, 1)
	case "up":
		g.focus = neighborBand(spots, here, -1)
	case "down":
		g.focus = neighborBand(spots, here, 1)
	}
}

func findSpot(spots []spot, w Widget) (spot, bool) {
	for _, s := range spots {
		if s.widget == w {
			return s, true
		}
	}
	return spot{}, false
}

func neighborCol(spots []spot, here spot, dir int) Widget {
	best := here.widget
	bestDist := 0
	found := false
	for _, s := range spots {
		delta := s.col - here.col
		if s.band != here.band || delta*dir <= 0 {
			continue
		}
		if delta < 0 {
			delta = -delta
		}
		if !found || delta < bestDist {
			best = s.widget
			bestDist = delta
			found = true
		}
	}
	return best
}

func neighborBand(spots []spot, here spot, dir int) Widget {
	minB, maxB := spots[0].band, spots[0].band
	for _, s := range spots[1:] {
		if s.band < minB {
			minB = s.band
		}
		if s.band > maxB {
			maxB = s.band
		}
	}
	for band := here.band + dir; band >= minB && band <= maxB; band += dir {
		if w, ok := nearest(spots, band, here.col); ok {
			return w
		}
	}
	return here.widget
}

func nearest(spots []spot, band, want int) (Widget, bool) {
	var best Widget
	dist := 0
	found := false
	for _, s := range spots {
		if s.band != band {
			continue
		}
		d := s.col - want
		if d < 0 {
			d = -d
		}
		if !found || d < dist {
			best = s.widget
			dist = d
			found = true
		}
	}
	return best, found
}

// render mesure chaque objet, puis donne le vide de droite au dernier
// de la ligne et la hauteur de la ligne au plus grand.
// L'écart horizontal entre deux objets est une seule cellule, de la couleur des zones.
// Les rangées se suivent sans ligne vide.
func (g *grid) render(s Style) []string {
	gap := s.margin()
	type size struct{ w, h int }
	span := func(sizes []size) int {
		sum := 0
		for _, sz := range sizes {
			sum += sz.w
		}
		if len(sizes) > 1 && gap > 0 {
			sum += (len(sizes) - 1) * gap
		}
		return sum
	}
	measured := make([][]size, len(g.rows))
	iface := 0
	for i, row := range g.rows {
		measured[i] = make([]size, len(row))
		for j, w := range row {
			view := w.View(s, g.focus, g.inside)
			measured[i][j] = size{w: lineWidth(view), h: lineCount(view)}
		}
		if sum := span(measured[i]); sum > iface {
			iface = sum
		}
	}
	lines := make([]string, len(g.rows))
	for i, row := range g.rows {
		if len(row) == 0 {
			continue
		}
		maxH := 0
		for _, sz := range measured[i] {
			if sz.h > maxH {
				maxH = sz.h
			}
		}
		slack := iface - span(measured[i])
		pieces := make([]string, 0, len(row)*2)
		for j, w := range row {
			if j > 0 && gap > 0 {
				pieces = append(pieces, s.veil(gap, maxH))
			}
			width := measured[i][j].w
			if j == len(row)-1 {
				width += slack
			}
			if width < 1 {
				width = 1
			}
			if maxH < 1 {
				maxH = 1
			}
			pieces = append(pieces, w.View(s.placed(width, maxH), g.focus, g.inside))
		}
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
	}
	return lines
}
