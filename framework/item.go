package framework

// Item est une ligne de liste. ID est la valeur réactive, Label le texte affiché.
type Item struct {
	ID    string
	Label string
}

func NewItem(id, label string) Item { return Item{ID: id, Label: label} }

// ListItem est le constructeur court d'une ligne : id puis libellé.
func ListItem(id, label string) Item { return NewItem(id, label) }

// ListArg est un item fixe ou un objet réactif accepté par NewListUniq.
type ListArg interface{ isListArg() }

func (Item) isListArg() {}
