package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FieldKind distingue une saisie visible d'un secret masqué.
type FieldKind int

const (
	Text FieldKind = iota
	Secret
)

// formField est un champ nommé. Le nom est la clé des libellés.
type formField struct {
	name  string
	label string
	input *textInput
}

// Field prépare un champ. Text est visible, Secret est masqué.
func Field(name string, kind FieldKind) *formField {
	mode := InputText
	if kind == Secret {
		mode = InputSecret
	}
	return &formField{name: name, label: name, input: NewTextInput(name, mode, 0, false)}
}

// form empile des champs, aligne leurs saisies et fait circuler le focus.
type form struct {
	fields []*formField
	box    *category
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
	f := &form{fields: fields, box: NewCategory(title, kids...)}
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

func (f *form) ColumnFields() bool { return true }

func (f *form) paint() {
	max := 0
	for _, field := range f.fields {
		if n := lipgloss.Width(field.label); n > max {
			max = n
		}
	}
	for _, field := range f.fields {
		pad := max - lipgloss.Width(field.label)
		field.input.SetTitle(field.label + strings.Repeat(" ", pad) + " :")
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
