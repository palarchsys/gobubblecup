package framework

// Item est une ligne de liste. ID est la valeur réactive, Label le texte affiché.
// Dot est la couleur de la pastille, #rrggbb. Vide : pas de pastille.
type Item struct {
	ID    string
	Label string
	Dot   string
}

func NewItem(id, label string) Item { return Item{ID: id, Label: label} }

// ListItem est le constructeur court d'une ligne : id puis libellé.
func ListItem(id, label string) Item { return NewItem(id, label) }

// WithDot pose la couleur de la pastille. Une couleur vide la retire.
func (it Item) WithDot(color string) Item {
	it.Dot = color
	return it
}

// ListArg est un item fixe ou un objet réactif accepté par NewListUniq.
type ListArg interface{ isListArg() }

func (Item) isListArg() {}
