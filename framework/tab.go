package framework

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TabItem choisit un onglet. id le relie à un Tab.
type tabItem struct {
	id    string
	label string
}

func TabItem(id, label string) *tabItem {
	if id == "" {
		panic("framework: onglet sans id")
	}
	return &tabItem{id: id, label: label}
}

func (t *tabItem) ID() string { return t.id }

// Tab est la grille affichée quand son item est choisi.
type tab struct {
	id   string
	grid *grid
}

func Tab(id string, grid *grid) *tab {
	if id == "" {
		panic("framework: onglet sans id")
	}
	if grid == nil {
		panic("framework: grille nulle")
	}
	return &tab{id: id, grid: grid}
}

// tabRow est une ligne de la grille d'onglet. Elle occupe plusieurs colonnes de focus.
type tabRow struct{ cells []Widget }

func (r *tabRow) Active() bool { return true }

func (r *tabRow) SetActive(bool) {}

func (r *tabRow) Disable() {}

func (r *tabRow) Enable() {}

func (r *tabRow) Disabled() bool { return false }

func (r *tabRow) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (r *tabRow) View(Style, Widget, bool) string { return "" }

// TabArg est un onglet fixe ou une source réactive d'onglets.
type TabArg interface{ isTabArg() }

func (t *tabItem) isTabArg() {}

// TabList pose une rangée d'onglets collée au panneau, sans titre.
// Les arguments sont des onglets fixes, ou une seule source, pas les deux.
// La barre n'est pas un arrêt de focus. Onglets et panneau restent l'or du repos.
type tabList struct {
	gate
	items    []*tabItem
	tabs     []*tab
	cursor   int
	selected int
	src      tabSource
	held     []Widget
}

// tabSource reconstruit les onglets, comme un ReactiveObject(TabItem, ...).
type tabSource interface{ Items() []*tabItem }

func TabList(args ...TabArg) *tabList {
	var items []*tabItem
	var src tabSource
	for _, arg := range args {
		switch v := arg.(type) {
		case *tabItem:
			if v == nil {
				panic("framework: item d'onglet nul")
			}
			items = append(items, v)
		case *reactiveObject[*tabItem]:
			if v == nil {
				panic("framework: argument d'onglet inattendu")
			}
			if src != nil {
				panic("framework: une seule source d'onglets")
			}
			src = v
		case nil:
			panic("framework: item d'onglet nul")
		default:
			panic("framework: argument d'onglet inattendu")
		}
	}
	if src != nil && len(items) > 0 {
		panic("framework: onglets fixes et réactifs")
	}
	list := &tabList{gate: gate{on: true}, items: items, selected: -1, src: src}
	list.sync()
	if src == nil && len(items) > 0 {
		list.selected = 0
	}
	return list
}

// sync relit la mémoire. L'onglet affiché reste celui du même id.
func (t *tabList) sync() {
	if t == nil || t.src == nil {
		return
	}
	id := ""
	if t.selected >= 0 && t.selected < len(t.items) && t.items[t.selected] != nil {
		id = t.items[t.selected].id
	}
	next := t.src.Items()
	t.items = next
	t.selected = -1
	t.cursor = 0
	if len(next) == 0 {
		return
	}
	t.selected = 0
	for i, item := range next {
		if item != nil && item.id == id {
			t.selected = i
			t.cursor = i
			return
		}
	}
}

func (t *tabList) SetTabs(tabs ...*tab) {
	for _, tab := range tabs {
		if tab == nil {
			panic("framework: onglet nul")
		}
	}
	t.tabs = tabs
}

func (t *tabList) Cursor() int {
	t.sync()
	return t.cursor
}

func (t *tabList) SelectedID() string {
	t.sync()
	if t.selected < 0 || t.selected >= len(t.items) || t.items[t.selected] == nil {
		return ""
	}
	return t.items[t.selected].id
}

func (t *tabList) Kids() []Widget {
	rows := t.activeRows()
	out := make([]Widget, 0, len(rows))
	for _, row := range rows {
		out = append(out, &tabRow{cells: row})
	}
	return out
}

func (t *tabList) Disable() {
	if t == nil || t.gate.disabled {
		return
	}
	t.gate.disabled = true
	t.held = nil
	t.each(func(w Widget) {
		if w.Disabled() {
			return
		}
		w.Disable()
		t.held = append(t.held, w)
	})
	notify()
}

func (t *tabList) Enable() {
	if t == nil || !t.gate.disabled {
		return
	}
	t.gate.disabled = false
	for _, w := range t.held {
		w.Enable()
	}
	t.held = nil
	notify()
}

func (t *tabList) each(fn func(Widget)) {
	for _, tab := range t.tabs {
		if tab == nil || tab.grid == nil {
			continue
		}
		for _, row := range tab.grid.rows {
			for _, w := range row {
				if w != nil {
					fn(w)
				}
			}
		}
	}
}

func (t *tabList) SetActive(on bool) {
	t.gate.SetActive(on)
	for _, tab := range t.tabs {
		for _, row := range tab.grid.rows {
			for _, w := range row {
				if w != nil {
					w.SetActive(on)
				}
			}
		}
	}
}

func (t *tabList) Reload() {
	for _, tab := range t.tabs {
		for _, row := range tab.grid.rows {
			for _, w := range row {
				if zone, ok := w.(Reloader); ok {
					zone.Reload()
				}
			}
		}
	}
}

func (t *tabList) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

// move change d'onglet et boucle d'un bout à l'autre.
// Faux s'il n'y a pas d'autre onglet.
func (t *tabList) move(delta int) bool {
	t.sync()
	n := len(t.items)
	if n < 2 {
		return false
	}
	next := (t.cursor + delta) % n
	if next < 0 {
		next += n
	}
	t.cursor = next
	t.selected = next
	return true
}

// firstFocus est le premier item ou la première zone de l'onglet affiché.
func (t *tabList) firstFocus() Widget {
	rows := t.activeRows()
	if len(rows) == 0 {
		return nil
	}
	if spots := (&grid{rows: rows}).spots(); len(spots) > 0 {
		return spots[0]
	}
	return nil
}

func (t *tabList) activeRows() [][]Widget {
	id := t.SelectedID()
	if id == "" {
		return nil
	}
	for _, tab := range t.tabs {
		if tab.id == id {
			return tab.grid.rows
		}
	}
	return nil
}

func (t *tabList) View(s Style, focused Widget, inside bool) string {
	if t.Disabled() {
		s = s.dimmed()
	}
	t.sync()
	padX := s.padX()
	rows := t.activeRows()
	// Mesure à la largeur d'origine. La largeur imposée par l'interface
	// ne doit pas redevenir la taille des enfants : le cadre serait compté deux fois.
	measured := measureRows(rows, s.preferred(), focused, inside)
	contentW := 0
	rowH := make([]int, len(measured))
	for i, row := range measured {
		span := 0
		for j, sz := range row {
			span += sz.w
			if j > 0 {
				span += s.margin()
			}
			if sz.h > rowH[i] {
				rowH[i] = sz.h
			}
		}
		if span > contentW {
			contentW = span
		}
	}
	head := tabStripWidth(t.tabBoxes())
	panelW := contentW + 2 + 2*padX
	if head > panelW {
		panelW = head
	}
	width := panelW
	if s.fit {
		width = s.Width
	}
	if width < 1 {
		width = panelW
	}
	inner := width - 2 - 2*padX
	if inner < 1 {
		inner = 1
	}
	bodyLines := 1
	if len(rows) > 0 {
		bodyLines = 0
		for _, h := range rowH {
			bodyLines += h
		}
	}
	// Deux lignes d'onglets, le trait du panneau, le corps, le trait bas.
	chrome := 4
	extra := 0
	if s.fit && s.Height > chrome+bodyLines {
		extra = s.Height - chrome - bodyLines
	}
	var parts []string
	for i, row := range rows {
		h := rowH[i]
		if i == len(rows)-1 {
			h += extra
		}
		parts = append(parts, placeRow(row, s, focused, inside, inner, h))
	}
	body := ""
	if len(parts) > 0 {
		body = strings.Join(parts, "\n")
	}
	return t.draw(s, width, body)
}

type tabBox struct {
	label  string
	inner  int
	L, R   int
	active bool
}

func (t *tabList) tabBoxes() []tabBox {
	boxes := make([]tabBox, len(t.items))
	x := 0
	for i, item := range t.items {
		inner := lipgloss.Width(item.label) + 2
		if inner < 2 {
			inner = 2
		}
		boxes[i] = tabBox{label: item.label, inner: inner, L: x, R: x + inner + 1, active: i == t.selected}
		x = boxes[i].R
	}
	return boxes
}

func tabStripWidth(boxes []tabBox) int {
	if len(boxes) == 0 {
		return 0
	}
	return boxes[len(boxes)-1].R + 1
}

func (t *tabList) draw(s Style, width int, body string) string {
	tl, tr, bl, br, dash, bar := s.corners()
	h := []rune(dash)[0]
	v := []rune(bar)[0]
	tlR, trR := []rune(tl)[0], []rune(tr)[0]
	panel, panelBold := s.ink(frameIdle)
	boxes := t.tabBoxes()
	blank := inkCell{r: ' ', c: s.Text, bold: false}
	idle := inkCell{r: h, c: s.Idle, bold: false}
	tops := make([]inkCell, width)
	mids := make([]inkCell, width)
	for i := range tops {
		tops[i] = blank
		mids[i] = blank
	}
	for i, b := range boxes {
		if b.R >= width {
			break
		}
		left := tlR
		if i > 0 {
			left = '┬'
		}
		tops[b.L] = inkCell{r: left, c: s.Idle, bold: false}
		for c := b.L + 1; c < b.R; c++ {
			tops[c] = idle
		}
		right := rune('┬')
		if i == len(boxes)-1 {
			right = trR
		}
		tops[b.R] = inkCell{r: right, c: s.Idle, bold: false}
		mids[b.L] = inkCell{r: v, c: s.Idle, bold: false}
		mids[b.R] = inkCell{r: v, c: s.Idle, bold: false}
		ink, bold := s.Text, false
		if b.active {
			ink, bold = s.Accent, true
		}
		col := b.L + 1
		if col < b.R {
			mids[col] = inkCell{r: ' ', c: s.Text, bold: false}
			col++
		}
		for _, r := range b.label {
			if col >= b.R {
				break
			}
			mids[col] = inkCell{r: r, c: ink, bold: bold}
			col++
		}
		for col < b.R {
			mids[col] = inkCell{r: ' ', c: s.Text, bold: false}
			col++
		}
	}
	join := make([]inkCell, width)
	for i := range join {
		join[i] = inkCell{r: h, c: panel, bold: panelBold}
	}
	if width > 0 {
		join[0] = inkCell{r: tlR, c: panel, bold: panelBold}
		join[width-1] = inkCell{r: trR, c: panel, bold: panelBold}
	}
	for _, b := range boxes {
		if b.active || b.L >= width {
			continue
		}
		if b.L == 0 {
			join[b.L] = inkCell{r: '├', c: s.Idle, bold: false}
		} else {
			join[b.L] = inkCell{r: '┴', c: s.Idle, bold: false}
		}
		for c := b.L + 1; c < b.R && c < width; c++ {
			join[c] = inkCell{r: h, c: s.Idle, bold: false}
		}
		if b.R >= width {
			continue
		}
		if b.R == width-1 {
			join[b.R] = inkCell{r: '┤', c: s.Idle, bold: false}
		} else {
			join[b.R] = inkCell{r: '┴', c: s.Idle, bold: false}
		}
	}
	for _, b := range boxes {
		if !b.active || b.L >= width {
			continue
		}
		if b.L == 0 {
			join[b.L] = inkCell{r: v, c: panel, bold: panelBold}
		} else {
			join[b.L] = inkCell{r: '┘', c: s.Idle, bold: false}
		}
		for c := b.L + 1; c < b.R && c < width; c++ {
			join[c] = inkCell{r: ' ', c: s.Text, bold: false}
		}
		if b.R >= width {
			continue
		}
		if b.R == width-1 {
			join[b.R] = inkCell{r: v, c: panel, bold: panelBold}
		} else {
			join[b.R] = inkCell{r: '└', c: s.Idle, bold: false}
		}
	}
	lines := []string{paintCells(s, tops), paintCells(s, mids), paintCells(s, join)}
	inner := width - 2 - 2*s.padX()
	if inner < 1 {
		inner = 1
	}
	rows := strings.Split(body, "\n")
	if body == "" {
		rows = []string{""}
	}
	side := s.shade(s.padX())
	wall := s.paint(panel, panelBold, bar)
	for _, row := range rows {
		lines = append(lines, wall+side+fit(row, inner, s)+side+wall)
	}
	dashes := width - 2
	if dashes < 0 {
		dashes = 0
	}
	bot := s.paint(panel, panelBold, bl+strings.Repeat(dash, dashes)+br)
	lines = append(lines, bot)
	return strings.Join(lines, "\n")
}

type inkCell struct {
	r    rune
	c    color.Color
	bold bool
}

func paintCells(s Style, cells []inkCell) string {
	var b strings.Builder
	for i := 0; i < len(cells); {
		j := i + 1
		for j < len(cells) && cells[j].bold == cells[i].bold && cells[j].c == cells[i].c {
			j++
		}
		var text strings.Builder
		for _, cell := range cells[i:j] {
			text.WriteRune(cell.r)
		}
		b.WriteString(s.paint(cells[i].c, cells[i].bold, text.String()))
		i = j
	}
	return b.String()
}

type pair struct{ w, h int }

func measureRows(rows [][]Widget, s Style, focused Widget, inside bool) [][]pair {
	out := make([][]pair, len(rows))
	for i, row := range rows {
		out[i] = make([]pair, len(row))
		for j, w := range row {
			if w == nil {
				continue
			}
			view := w.View(s, focused, inside)
			out[i][j] = pair{w: lineWidth(view), h: lineCount(view)}
		}
	}
	return out
}

func placeRow(row []Widget, s Style, focused Widget, inside bool, width, height int) string {
	if len(row) == 0 {
		return fit("", width, s)
	}
	gap := s.margin()
	base := s.preferred()
	sizes := make([]int, len(row))
	sum := 0
	for i, w := range row {
		if w == nil {
			continue
		}
		sizes[i] = lineWidth(w.View(base, focused, inside))
		sum += sizes[i]
	}
	if len(row) > 1 && gap > 0 {
		sum += (len(row) - 1) * gap
	}
	slack := width - sum
	if slack > 0 && len(sizes) > 0 {
		sizes[len(sizes)-1] += slack
	}
	if slack < 0 {
		over := -slack
		for i := len(sizes) - 1; i >= 0 && over > 0; i-- {
			room := sizes[i] - 1
			if room < 1 {
				continue
			}
			if room > over {
				room = over
			}
			sizes[i] -= room
			over -= room
		}
	}
	if height < 1 {
		height = 1
	}
	pieces := make([]string, 0, len(row)*2)
	for i, w := range row {
		if i > 0 && gap > 0 {
			pieces = append(pieces, s.veil(gap, height))
		}
		cellW := sizes[i]
		if cellW < 1 {
			cellW = 1
		}
		if w == nil {
			pieces = append(pieces, s.veil(cellW, height))
			continue
		}
		pieces = append(pieces, w.View(s.placed(cellW, height), focused, inside))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
}
