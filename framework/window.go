package framework

// window montre une grille à la fois. Suivante et Précédente font le tour.
// À l'arrivée, le focus est le premier élément de cette grille.
type window struct {
	pages []*grid
	index int
}

func Window() *window { return &window{} }

// Stage est une page du parcours. Legend est le mot du cadre.
type Stage struct {
	Legend string
	Kids   []Widget
}

// Walk pose une page par étape. Le titre et les puces sont communs.
// Suite avance, Retour recule. La première étape n'a pas Retour, la dernière n'a pas Suite.
func (w *window) Walk(title string, facts Widget, stages ...Stage) {
	if w == nil || len(stages) == 0 {
		panic("framework: fenêtre vide")
	}
	pages := make([]*grid, 0, len(stages))
	for i, stage := range stages {
		var next, prev Action
		if i+1 < len(stages) {
			next = w.Next()
		}
		if i > 0 {
			prev = w.Prev()
		}
		body := Section(stage.Legend, Moves(next, prev, stage.Kids...)...)
		pages = append(pages, Grid(Row(Page(title, facts, body))))
	}
	w.Use(pages...)
}

// Use fixe les grilles. La première est affichée.
func (w *window) Use(pages ...*grid) {
	if w == nil || len(pages) == 0 {
		panic("framework: fenêtre vide")
	}
	for _, page := range pages {
		if page == nil {
			panic("framework: grille nulle")
		}
	}
	w.pages = pages
	w.index = 0
}

func (w *window) current() *grid {
	if w == nil || len(w.pages) == 0 {
		return Grid()
	}
	return w.pages[w.index]
}

func (w *window) turn(delta int) {
	n := len(w.pages)
	if n < 2 {
		return
	}
	next := (w.index + delta) % n
	if next < 0 {
		next += n
	}
	w.index = next
	w.pages[next].focusFirst()
}

// Next passe à la grille suivante. Le parcours reprend au début.
func (w *window) Next() Action { return Call{Run: func() { w.turn(1) }} }

// Prev passe à la grille précédente. Le parcours reprend à la fin.
func (w *window) Prev() Action { return Call{Run: func() { w.turn(-1) }} }

func (w *window) Index() int {
	if w == nil {
		return 0
	}
	return w.index
}

func (w *window) reveal(target Widget) {
	if w == nil || target == nil {
		return
	}
	for i, page := range w.pages {
		if page != nil && page.reveal(target) {
			w.index = i
			return
		}
	}
}
