package framework

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Grid définit les lignes et les objets de chaque ligne.
// Le focus est l'objet actif, pas la case qui le contient.
type grid struct {
	rows     [][]Widget
	focus    Widget
	inside   bool
	vertical bool
	minWidth int
}

func Row(widgets ...Widget) []Widget { return widgets }

// RowEnd aligne la rangée à droite. Le vide va à gauche.
// Les objets gardent leur largeur. Row continue de donner le vide au dernier.
func RowEnd(widgets ...Widget) []Widget {
	row := make([]Widget, 0, len(widgets)+1)
	row = append(row, Fill())
	return append(row, widgets...)
}

// Fill prend le vide horizontal de sa rangée.
// Ce qui est à sa gauche reste à gauche. Ce qui est à sa droite reste à droite.
// Le premier Fill de la ligne reçoit le vide. Il ne prend pas le focus.
func Fill() Widget { return slack{} }

func Grid(rows ...[]Widget) *grid {
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

// Width vise au moins n colonnes pour chaque ligne, marge exclue.
// Une ligne déjà plus large le reste. n inférieur à 1 ne change rien.
// Le surplus va à Take, sinon au premier Fill, sinon au dernier objet.
func (g *grid) Width(n int) *grid {
	if g == nil || n < 1 || g.minWidth == n {
		return g
	}
	g.minWidth = n
	notify()
	return g
}

func (g *grid) Focused() Widget { return g.focus }

// SetVerticalFields envoie haut et bas d'un champ au champ voisin.
// Gauche et droite restent dans le champ. Sans appel, haut et bas n'en sortent pas.
// Un formulaire le fait lui-même.
func (g *grid) SetVerticalFields(on bool) {
	if g == nil {
		return
	}
	g.vertical = on
}

func (g *grid) Inside() bool { return g.inside }

// spots est la file affichée, de haut-gauche vers droite-bas.
// Seules les zones actives de l'onglet visible y entrent.
func (g *grid) spots() []Widget {
	var out []Widget
	for _, row := range g.rows {
		height := 1
		for _, w := range row {
			if n := depth(w); n > height {
				height = n
			}
		}
		for sub := 0; sub < height; sub++ {
			for _, w := range row {
				for _, child := range lineAt(w, sub) {
					if child != nil && child.Active() {
						out = append(out, child)
					}
				}
			}
		}
	}
	return out
}

func isList(w Widget) bool {
	switch w.(type) {
	case *listUniq, *listMult:
		return true
	default:
		return false
	}
}

// reveal affiche l'onglet qui contient la zone et y pose le focus.
func (g *grid) reveal(target Widget) bool {
	if target == nil {
		return false
	}
	for _, list := range g.tabLists() {
		for i, tab := range list.tabs {
			if tab == nil || tab.grid == nil || !gridHolds(tab.grid, target) {
				continue
			}
			list.cursor = i
			list.selected = i
			g.inside = false
			g.focus = target
			g.follow()
			return true
		}
	}
	if gridHolds(g, target) {
		g.inside = false
		g.focus = target
		g.follow()
		return true
	}
	return false
}

func gridHolds(g *grid, target Widget) bool {
	if g == nil {
		return false
	}
	for _, row := range g.rows {
		for _, cell := range row {
			if widgetHolds(cell, target) {
				return true
			}
		}
	}
	return false
}

func widgetHolds(w, target Widget) bool {
	if w == nil || target == nil {
		return false
	}
	if w == target {
		return true
	}
	group, ok := w.(kids)
	if !ok {
		return false
	}
	for _, kid := range group.Kids() {
		if widgetHolds(kid, target) {
			return true
		}
	}
	return false
}

func (g *grid) focusFirst() {
	g.inside = false
	if spots := g.spots(); len(spots) > 0 {
		g.focus = spots[0]
	} else {
		g.focus = nil
	}
	g.follow()
}

// follow ouvre la saisie dès que le focus est un champ d'une ligne.
func (g *grid) follow() {
	if _, ok := g.focus.(*textInput); ok {
		g.inside = true
	}
}

func (g *grid) Key(key string) tea.Cmd {
	if key == "tab" {
		if g.switchTab() {
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
	if isList(w) && (key == "up" || key == "down" || key == "enter") {
		_, cmd := w.Handle(false, key)
		g.inside = false
		return cmd
	}
	if field, ok := w.(*textInput); ok {
		if g.columnFields(w) {
			switch key {
			case "up":
				g.inside = false
				g.move("left")
				return nil
			case "down":
				g.inside = false
				g.move("right")
				return nil
			}
			g.inside = true
			_, cmd := field.Handle(true, key)
			return cmd
		}
		if key == "up" || key == "down" {
			g.inside = true
			return nil
		}
		if (key == "left" && field.caret == 0) ||
			(key == "right" && field.caret >= len(field.value)) {
			g.inside = false
			g.move(key)
			g.follow()
			return nil
		}
		g.inside = true
		_, cmd := field.Handle(true, key)
		return cmd
	}
	if g.inside {
		inside, cmd := w.Handle(true, key)
		g.inside = inside
		return cmd
	}
	if key == "enter" || key == "tab" {
		inside, cmd := w.Handle(false, key)
		g.inside = inside
		return cmd
	}
	if key == "up" || key == "down" {
		_, cmd := w.Handle(false, key)
		return cmd
	}
	g.move(key)
	return nil
}

// switchTab consomme Tab s'il existe une liste d'onglets.
// Tab va vers la droite. Le parcours boucle.
func (g *grid) switchTab() bool {
	lists := g.tabLists()
	if len(lists) == 0 {
		return false
	}
	if len(lists) > 1 {
		panic("framework: une seule liste d'onglets par interface")
	}
	if lists[0].Disabled() {
		return true
	}
	if !lists[0].move(1) {
		return true
	}
	g.inside = false
	if w := lists[0].firstFocus(); w != nil {
		g.focus = w
		g.follow()
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
	for _, w := range g.spots() {
		if w == target {
			return true
		}
	}
	return false
}

func (g *grid) canPaste() bool {
	switch w := g.focus.(type) {
	case *textInput:
		return w.editable
	case *textArea:
		return w.editable
	default:
		return false
	}
}

func (g *grid) paste(text string) {
	if g == nil || text == "" || g.focus == nil || !g.focus.Active() {
		return
	}
	switch w := g.focus.(type) {
	case *textInput:
		w.Paste(text)
		g.inside = true
	case *textArea:
		w.Paste(text)
		if w.editable {
			g.inside = true
		}
	}
}

// move avance dans la file affichée. Droite va vers le bas-droit, gauche remonte.
// Les deux bouts se rejoignent. Haut et bas ne changent pas de zone.
func (g *grid) move(key string) {
	if key != "left" && key != "right" {
		return
	}
	spots := g.spots()
	if len(spots) == 0 {
		return
	}
	index := 0
	found := false
	for i, w := range spots {
		if w == g.focus {
			index = i
			found = true
			break
		}
	}
	if !found {
		g.focus = spots[0]
		g.follow()
		return
	}
	if key == "right" {
		index++
		if index >= len(spots) {
			index = 0
		}
	} else {
		index--
		if index < 0 {
			index = len(spots) - 1
		}
	}
	g.focus = spots[index]
	g.follow()
}

// render mesure chaque objet, puis donne le vide de droite au dernier
// de la ligne et la hauteur de la ligne au plus grand.
// RowEnd place ce vide à gauche. Les autres objets gardent leur largeur.
// L'écart horizontal entre deux objets est une seule cellule, de la couleur des zones.
// Les rangées se suivent sans ligne vide.
// Sans budget, la hauteur reste celle du contenu.
func (g *grid) render(s Style) []string {
	return g.layout(s, 0, false)
}

// rowFit décrit une ligne pour le partage de la hauteur utile.
// floor est le plancher : cadre d'une liste, hauteur naturelle d'une zone,
// ou la hauteur d'un objet qui ne défile pas posé sur la même ligne.
type rowFit struct {
	flex  bool
	text  bool
	floor int
}

func (f *rowFit) note(w Widget, h int, s Style) {
	switch v := w.(type) {
	case *listUniq:
		f.flex = true
		if n := v.minBox(s); n > f.floor {
			f.floor = n
		}
	case *listMult:
		f.flex = true
		if n := v.minBox(s); n > f.floor {
			f.floor = n
		}
	case *textArea:
		f.flex = true
		f.text = true
		if h > f.floor {
			f.floor = h
		}
	default:
		if h > f.floor {
			f.floor = h
		}
	}
}

// fitHeights donne aux lignes fixes leur hauteur naturelle.
// fill faux : le budget ne fait que réduire un contenu trop haut.
// fill vrai : l'espace restant va aux listes et aux zones.
// Une zone garde au moins sa hauteur naturelle tant que le budget le permet.
// Le surplus va d'abord aux zones. La somme ne dépasse le budget que si
// les objets fixes sont déjà plus hauts que l'écran.
func fitHeights(natural []int, rows []rowFit, budget int, fill bool) []int {
	out := append([]int(nil), natural...)
	if budget <= 0 {
		return out
	}
	fixed := 0
	var flex []int
	for i, row := range rows {
		if row.flex {
			flex = append(flex, i)
			continue
		}
		fixed += natural[i]
	}
	if len(flex) == 0 {
		return out
	}
	room := budget - fixed
	if room <= 0 {
		return out
	}
	var textRows, listRows []int
	sumNatural := 0
	for _, i := range flex {
		sumNatural += natural[i]
		if rows[i].text {
			textRows = append(textRows, i)
		}
		if rows[i].floor > 0 && !rows[i].text {
			listRows = append(listRows, i)
		}
		if rows[i].floor > 0 && rows[i].text && natural[i] > rows[i].floor {
			listRows = append(listRows, i)
		}
	}
	if sumNatural <= room {
		if !fill {
			return out
		}
		extra := room - sumNatural
		give := textRows
		if len(give) == 0 {
			give = flex
		}
		shareExtra(out, give, extra)
		return out
	}
	textNeed := 0
	for _, i := range textRows {
		if containsInt(listRows, i) {
			continue
		}
		textNeed += natural[i]
		out[i] = natural[i]
	}
	if textNeed >= room {
		assignExact(out, flex, room)
		return out
	}
	assignAtLeast(out, listRows, rows, room-textNeed)
	return out
}

func containsInt(list []int, n int) bool {
	for _, v := range list {
		if v == n {
			return true
		}
	}
	return false
}

func shareExtra(out []int, rows []int, extra int) {
	if len(rows) == 0 || extra <= 0 {
		return
	}
	base := extra / len(rows)
	rem := extra % len(rows)
	for i, row := range rows {
		out[row] += base
		if i >= len(rows)-rem {
			out[row]++
		}
	}
}

func assignExact(out []int, rows []int, total int) {
	if len(rows) == 0 {
		return
	}
	if total < 0 {
		total = 0
	}
	base := total / len(rows)
	rem := total % len(rows)
	for i, row := range rows {
		out[row] = base
		if i >= len(rows)-rem {
			out[row]++
		}
	}
}

func assignAtLeast(out []int, rows []int, fit []rowFit, total int) {
	if len(rows) == 0 {
		return
	}
	sum := 0
	for _, row := range rows {
		floor := fit[row].floor
		if floor < 1 {
			floor = 1
		}
		sum += floor
	}
	if sum > total {
		assignExact(out, rows, total)
		return
	}
	for _, row := range rows {
		floor := fit[row].floor
		if floor < 1 {
			floor = 1
		}
		out[row] = floor
	}
	shareExtra(out, rows, total-sum)
}

func (g *grid) layout(s Style, budget int, fill bool) []string {
	if g.focus != nil && !g.focus.Active() {
		g.inside = false
		g.focusFirst()
	}
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
	natural := make([]int, len(g.rows))
	fits := make([]rowFit, len(g.rows))
	iface := 0
	for i, row := range g.rows {
		measured[i] = make([]size, len(row))
		for j, w := range row {
			view := w.View(s, g.focus, g.inside)
			measured[i][j] = size{w: lineWidth(view), h: lineCount(view)}
			if measured[i][j].h > natural[i] {
				natural[i] = measured[i][j].h
			}
			fits[i].note(w, measured[i][j].h, s)
		}
		if sum := span(measured[i]); sum > iface {
			iface = sum
		}
	}
	if s.Width > iface && rowHasSlack(g.rows) {
		iface = s.Width
	}
	if g.minWidth > iface {
		iface = g.minWidth
	}
	heights := fitHeights(natural, fits, budget, fill)
	lines := make([]string, len(g.rows))
	for i, row := range g.rows {
		if len(row) == 0 {
			continue
		}
		maxH := heights[i]
		slack := iface - span(measured[i])
		slackAt := len(row) - 1
		for j, w := range row {
			if taker, ok := w.(slacker); ok && taker.TakesSlack() {
				slackAt = j
				break
			}
		}
		pieces := make([]string, 0, len(row)*2)
		for j, w := range row {
			if j > 0 && gap > 0 {
				pieces = append(pieces, s.veil(gap, maxH))
			}
			width := measured[i][j].w
			if j == slackAt {
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

func (g *grid) columnFields(w Widget) bool {
	if g == nil || w == nil {
		return false
	}
	if g.vertical {
		return true
	}
	for _, row := range g.rows {
		for _, cell := range row {
			if columnHolds(cell, w) {
				return true
			}
		}
	}
	return false
}

func columnHolds(cell, target Widget) bool {
	if cell == nil || target == nil {
		return false
	}
	if col, ok := cell.(columnFields); ok && col.ColumnFields() && widgetHolds(cell, target) {
		return true
	}
	group, ok := cell.(kids)
	if !ok {
		return false
	}
	for _, kid := range group.Kids() {
		if columnHolds(kid, target) {
			return true
		}
	}
	return false
}

func rowHasSlack(rows [][]Widget) bool {
	for _, row := range rows {
		for _, w := range row {
			if taker, ok := w.(slacker); ok && taker.TakesSlack() {
				return true
			}
		}
	}
	return false
}

// slack est le vide que Fill et RowEnd prennent dans la rangée.
type slack struct{}

func (slack) TakesSlack() bool { return true }

func (slack) Active() bool { return false }

func (slack) SetActive(bool) {}

func (slack) Disable() {}

func (slack) Enable() {}

func (slack) Disabled() bool { return false }

func (slack) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (slack) View(s Style, _ Widget, _ bool) string {
	n := 1
	if s.fit && s.Width > 0 {
		n = s.Width
	}
	h := 1
	if s.fit && s.Height > 1 {
		h = s.Height
	}
	if h == 1 {
		return s.shade(n)
	}
	return s.veil(n, h)
}
