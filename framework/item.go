package framework

// Item est une ligne de liste. ID est la valeur réactive, Label le texte affiché.
// Dot est la couleur de la pastille : #rrggbb, ou pending, success, warning, error.
// Vide, ou un autre mot : pas de pastille.
type Item struct {
	ID    string
	Label string
	Dot   string
}

// NewListItem est l'alias de ListItem.
func NewListItem(id, label string) Item { return Item{ID: id, Label: label} }

// ListItem est le constructeur court d'une ligne : id puis libellé.
func ListItem(id, label string) Item { return NewListItem(id, label) }

// NewItem est l'ancien nom de NewListItem.
func NewItem(id, label string) Item { return NewListItem(id, label) }

// WithDot pose la couleur de la pastille. Une couleur vide la retire.
func (it Item) WithDot(color string) Item {
	it.Dot = color
	return it
}

// ListArg est un item fixe ou un objet réactif accepté par NewListUniq.
type ListArg interface{ isListArg() }

func (Item) isListArg() {}
