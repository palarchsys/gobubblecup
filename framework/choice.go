package framework

// Choice est un arbre dont la ligne courante est la valeur.
// Haut et bas la déplacent, sans boucler. La gauche et la droite quittent l'objet.
type choice struct {
	gate
	title  string
	items  []Item
	cursor int
	store  Store
	key    string
	sent   bool
	sentID string
}

// Choice prépare les lignes. La première est courante.
func Choice(title string, items ...Item) *choice {
	return &choice{gate: gate{on: true}, title: title, items: append([]Item(nil), items...)}
}

// Bind publie l'id focalisé. La première ligne part tout de suite.
// Une mémoire ou une clé vide arrête le programme.
func (c *choice) Bind(store Store, key string) {
	if c == nil {
		return
	}
	if store == nil || key == "" {
		panic("framework: lien")
	}
	c.store = store
	c.key = key
	c.publish()
}

// Value est l'id focalisé.
func (c *choice) Value() string {
	if c == nil || len(c.items) == 0 || c.cursor < 0 || c.cursor >= len(c.items) {
		return ""
	}
	return c.items[c.cursor].ID
}

// Label est le texte focalisé. Vide, il devient —.
func (c *choice) Label() string {
	if c == nil || len(c.items) == 0 || c.cursor < 0 || c.cursor >= len(c.items) {
		return "—"
	}
	if c.items[c.cursor].Label == "" {
		return "—"
	}
	return c.items[c.cursor].Label
}

func (c *choice) publish() {
	if c == nil || c.store == nil || c.key == "" {
		return
	}
	id := c.Value()
	if c.sent && c.sentID == id {
		return
	}
	c.store.Set(c.key, id)
	c.sent = true
	c.sentID = id
}

func (c *choice) move(delta int) {
	if c == nil || len(c.items) == 0 {
		return
	}
	next := c.cursor + delta
	if next < 0 || next >= len(c.items) {
		return
	}
	c.cursor = next
	c.publish()
	notify()
}

func (c *choice) Handle(_ bool, key string) (bool, Cmd) {
	if c == nil || c.disabled {
		return false, nil
	}
	switch key {
	case "up":
		c.move(-1)
	case "down":
		c.move(1)
	case "enter":
		c.publish()
	}
	return false, nil
}

func (c *choice) View(s Style, focused Widget, _ bool) string {
	width := s.Width
	if width < 1 {
		width = 1
	}
	if c == nil {
		return s.shade(width)
	}
	if c.Disabled() {
		s = s.dimmed()
	}
	armed := focused == c && !c.Disabled()
	rows := make([]Line, len(c.items))
	for i, item := range c.items {
		kind := Info
		if i == c.cursor {
			kind = OK
		}
		rows[i] = Line{Mark: kind, Text: item.Label, Bold: armed && i == c.cursor}
	}
	return stepBlock(s, c.title, rows, width)
}
