package framework

// window montre une grille à la fois. Suivante et Précédente font le tour.
type window struct {
	pages []*grid
	index int
}

func Window() *window { return &window{} }

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
		return NewGrid()
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
