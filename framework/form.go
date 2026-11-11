package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// formField est un champ nommé. Le nom est la clé des libellés.
type formField struct {
	name  string
	label string
	input *textInput
}

// Field prépare un champ. kind est InputText, InputSecret ou InputNumber.
func Field(name string, kind InputKind) *formField {
	return &formField{name: name, label: name, input: TextInput(name, kind, 0, false)}
}

// Fields prépare un champ par nom, dans cet ordre.
// Chaque nom listé dans secrets est masqué. Les autres sont visibles.
// Un secret absent de names arrête le programme.
func Fields(names []string, secrets ...string) []*formField {
	known := make(map[string]bool, len(names))
	for _, name := range names {
		known[name] = true
	}
	for _, name := range secrets {
		if !known[name] {
			panic("framework: champ " + name)
		}
	}
	hidden := make(map[string]bool, len(secrets))
	for _, name := range secrets {
		hidden[name] = true
	}
	out := make([]*formField, len(names))
	for i, name := range names {
		kind := InputText
		if hidden[name] {
			kind = InputSecret
		}
		out[i] = Field(name, kind)
	}
	return out
}

// Placeholder pose l'invite du champ et le verrouille.
// Elle n'entre pas dans la valeur.
func (f *formField) Placeholder(text string) *formField {
	if f == nil || f.input == nil {
		return f
	}
	f.input.Placeholder(text)
	return f
}

// Lock montre l'invite quand la saisie est vide.
func (f *formField) Lock() *formField {
	if f == nil || f.input == nil {
		return f
	}
	f.input.Lock()
	return f
}

// Unlock cache l'invite. Une saisie vide reste vide.
func (f *formField) Unlock() *formField {
	if f == nil || f.input == nil {
		return f
	}
	f.input.Unlock()
	return f
}

// form empile des champs, aligne leurs saisies et fait circuler le focus.
type form struct {
	fields    []*formField
	box       *category
	labelCols int
	fieldCols int
}

// Form empile les champs sous title. Haut et bas passent au champ voisin.
// Les saisies commencent sur la même colonne.
func Form(title string, fields ...*formField) *form {
	kids := make([]Widget, 0, len(fields))
	for _, field := range fields {
		if field == nil || field.input == nil {
			panic("framework: champ nul")
		}
		kids = append(kids, field.input)
	}
	f := &form{fields: fields, box: Category(title, kids...)}
	f.paint()
	return f
}

// Span fixe la colonne des libellés et la largeur des saisies.
// 0 sur les libellés suit le plus long. 0 sur les saisies laisse la grille choisir.
// La colonne compte le texte du libellé, avant « : ».
func (f *form) Span(label, field int) *form {
	if f == nil {
		return nil
	}
	if label < 0 {
		label = 0
	}
	if field < 0 {
		field = 0
	}
	f.labelCols = label
	f.fieldCols = field
	f.paint()
	return f
}

// SetTitle change le titre posé sur le trait.
func (f *form) SetTitle(title string) {
	if f == nil {
		return
	}
	f.box.SetLegend(title)
}

// SetLabels remplace les libellés et garde les valeurs.
// Un nom de champ absent arrête le programme.
func (f *form) SetLabels(labels map[string]string) {
	if f == nil {
		return
	}
	for _, field := range f.fields {
		text, ok := labels[field.name]
		if !ok {
			panic("framework: libellé " + field.name)
		}
		field.label = text
	}
	f.paint()
}

// Use pose les libellés de lang. Une langue ou un nom absent arrête le programme.
func (f *form) Use(labels map[string]map[string]string, lang string) {
	if f == nil {
		return
	}
	table, ok := labels[lang]
	if !ok {
		panic("framework: langue " + lang)
	}
	f.SetLabels(table)
}

// Values rend les textes dans l'ordre des Field.
func (f *form) Values() []string {
	if f == nil {
		return nil
	}
	out := make([]string, len(f.fields))
	for i, field := range f.fields {
		out[i] = field.input.Value()
	}
	return out
}

// Value rend la saisie de ce nom. Un nom absent rend une chaîne vide.
func (f *form) Value(name string) string {
	if f == nil {
		return ""
	}
	for _, field := range f.fields {
		if field.name == name {
			return field.input.Value()
		}
	}
	return ""
}

func (f *form) ColumnFields() bool { return true }

func (f *form) paint() {
	if f == nil {
		return
	}
	max := f.labelCols
	if max <= 0 {
		max = 0
		for _, field := range f.fields {
			if n := lipgloss.Width(field.label); n > max {
				max = n
			}
		}
	}
	for _, field := range f.fields {
		text := field.label
		if max > 0 && lipgloss.Width(text) > max {
			text = clipWidth(text, max)
		}
		pad := max - lipgloss.Width(text)
		if pad < 0 {
			pad = 0
		}
		field.input.labelCols = 0
		field.input.fieldCols = f.fieldCols
		field.input.SetTitle(text + strings.Repeat(" ", pad) + " :")
	}
}

func (f *form) Active() bool { return f.box.Active() }

func (f *form) SetActive(on bool) { f.box.SetActive(on) }

func (f *form) Disable() { f.box.Disable() }

func (f *form) Enable() { f.box.Enable() }

func (f *form) Disabled() bool { return f.box.Disabled() }

func (f *form) Kids() []Widget { return f.box.Kids() }

func (f *form) Reload() { f.box.Reload() }

func (f *form) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (f *form) View(s Style, focused Widget, inside bool) string {
	return f.box.View(s, focused, inside)
}
