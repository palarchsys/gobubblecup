package framework

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSpanAndFillAlign(t *testing.T) {
	s := DefaultStyle()
	s.Width = 40
	s.Margin = 1
	left := TextInput("Nom", InputText, 0, false).Span(8, 4)
	below := TextInput("Adresse", InputText, 0, false).Span(8, 4)
	left.SetText("AB")
	below.SetText("CD")
	rows := Grid(
		Row(left, Fill(), Label("ok", 4)),
		Row(below, Fill(), Label("non", 4)),
	).layout(s, 0, false)
	top := plain(rows[0])
	bottom := plain(rows[1])
	if lineWidth(rows[0]) != 40 || lineWidth(rows[1]) != 40 {
		t.Fatalf("largeur %d %d", lineWidth(rows[0]), lineWidth(rows[1]))
	}
	if strings.Index(top, "AB") != 9 || strings.Index(bottom, "CD") != 9 {
		t.Fatalf("saisie\n%s\n%s", top, bottom)
	}
	if strings.Index(top, "ok") != strings.Index(bottom, "non") {
		t.Fatalf("droite\n%s\n%s", top, bottom)
	}
	boxed := TextInput("A", InputText, 0, true).Span(4, 6)
	other := TextInput("BBBB", InputText, 0, true).Span(4, 6)
	wide := DefaultStyle()
	wide.Width = 40
	if w := lineWidth(boxed.View(wide, nil, false)); w != lineWidth(other.View(wide, nil, false)) || w >= 40 {
		t.Fatalf("cadre %d %d", w, lineWidth(other.View(wide, nil, false)))
	}
	form := Form("Compte", Field("NAME", InputText), Field("CODE", InputSecret)).Span(4, 3)
	view := plain(form.View(wide, nil, false))
	nameAt, codeAt := -1, -1
	for _, line := range strings.Split(view, "\n") {
		if i := strings.Index(line, "NAME"); i >= 0 {
			nameAt = i
		}
		if i := strings.Index(line, "CODE"); i >= 0 {
			codeAt = i
		}
	}
	if nameAt < 0 || nameAt != codeAt {
		t.Fatalf("formulaire %d %d\n%s", nameAt, codeAt, view)
	}
	var none *textInput
	if none.Span(2, 2) != nil {
		t.Fatal("nil")
	}
}

func TestCurseurRecouvreLeCaractere(t *testing.T) {
	nom := TextInput("", InputText, 0, false)
	nom.SetText("ab")
	nom.caret = 1
	view := plain(nom.View(DefaultStyle(), nom, true))
	i := strings.Index(view, "a")
	if i < 0 || i+1 >= len(view) || view[i+1] != 'b' {
		t.Fatalf("trou dans %q", view)
	}
	if strings.Contains(view, "▏") {
		t.Fatal(view)
	}
	raw := nom.View(DefaultStyle(), nom, true)
	if !lineHas(raw, "ab", "68;170;230") {
		t.Fatal("le caractère sous le curseur n'est pas marqué")
	}
	if !lineHas(nom.View(DefaultStyle(), nom, false), "ab", "68;170;230") {
		t.Fatal("le rectangle n'apparaît pas au focus")
	}
	if lineHas(nom.View(DefaultStyle(), nil, false), "ab", "68;170;230") {
		t.Fatal("le rectangle apparaît sans focus")
	}
	nom.caret = 2
	end := plain(nom.View(DefaultStyle(), nom, true))
	j := strings.Index(end, "ab")
	if j < 0 || j+2 >= len(end) || end[j+2] != ' ' {
		t.Fatalf("fin de ligne %q", end)
	}
}

func TestLabelAuFocus(t *testing.T) {
	nom := TextInput("Nom", InputText, 0, false)
	nom.SetText("a")
	repos := nom.View(DefaultStyle(), nil, false)
	if !lineHas(repos, "Nom", "204;204;204") || lineHas(repos, "Nom", "236;203;168") {
		t.Fatal("label au repos")
	}
	if !lineHas(nom.View(DefaultStyle(), nom, false), "Nom", "236;203;168") {
		t.Fatal("label au focus")
	}
}

func TestArriveeSurChamp(t *testing.T) {
	bouton := Button("Ok", nil)
	nom := TextInput("Nom", InputText, 0, false)
	nom.SetText("ab")
	app := New(DefaultStyle(), Grid(Row(bouton, nom)))
	if app.grid.Focused() != bouton || app.grid.Inside() {
		t.Fatal("départ")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != nom || !app.grid.Inside() {
		t.Fatal("le champ n'est pas ouvert")
	}
	if !lineHas(app.View().Content, "ab", "68;170;230") {
		t.Fatal("rectangle absent")
	}
}

func TestZoneLieeAuSlot(t *testing.T) {
	src := Memory()
	focus := Memory()
	src.setField("un", "return", "alpha")
	src.setField("deux", "return", "beta")
	focus.Set("courant", "un")
	zone := TextView(src, focus, "courant")
	if !strings.Contains(plain(zone.View(DefaultStyle(), nil, false)), "alpha") {
		t.Fatal("le return n'est pas lu")
	}
	focus.Set("courant", "deux")
	zone.top = 3
	vue := plain(zone.View(DefaultStyle(), nil, false))
	if !strings.Contains(vue, "beta") || strings.Contains(vue, "alpha") || zone.Top() != 0 {
		t.Fatalf("changement d'id top %d\n%s", zone.Top(), vue)
	}

	var lignes []string
	for i := 0; i < 3; i++ {
		lignes = append(lignes, "court")
	}
	src.setField("deux", "return", strings.Join(lignes, "\n"))
	zone.View(DefaultStyle(), nil, false)
	if zone.Top() != 0 {
		t.Fatal("un texte court ne doit pas défiler")
	}
	for i := 0; i < 8; i++ {
		lignes = append(lignes, "suite")
	}
	src.setField("deux", "return", strings.Join(lignes, "\n"))
	zone.View(DefaultStyle(), nil, false)
	if zone.Top() == 0 {
		t.Fatal("le bas n'est plus visible")
	}
	haut := zone.Top()
	zone.Handle(true, "up")
	zone.Handle(true, "up")
	remonté := zone.Top()
	if remonté >= haut {
		t.Fatalf("la fenêtre n'est pas remontée %d", remonté)
	}
	lignes = append(lignes, "encore")
	src.setField("deux", "return", strings.Join(lignes, "\n"))
	zone.View(DefaultStyle(), nil, false)
	if zone.Top() != remonté {
		t.Fatalf("le haut a bougé %d → %d", remonté, zone.Top())
	}

	edit := TextArea("Notes", "a\nb\nc\nd\ne", true, 3, true)
	edit.top = 2
	edit.SetText("seul")
	if edit.Top() != 0 {
		t.Fatal("SetText ne remet plus le haut à zéro")
	}
}

func TestSupprEtInser(t *testing.T) {
	nom := TextInput("Nom", InputText, 0, true)
	nom.SetText("abc")
	nom.caret = 1
	nom.Handle(true, "delete")
	if nom.Value() != "ac" || nom.caret != 1 {
		t.Fatalf("suppr %q caret %d", nom.Value(), nom.caret)
	}
	nom.Handle(true, "delete")
	nom.Handle(true, "delete")
	if nom.Value() != "a" || nom.caret != 1 {
		t.Fatalf("fin %q caret %d", nom.Value(), nom.caret)
	}

	nom.SetText("abc")
	nom.caret = 1
	nom.Handle(true, "insert")
	nom.Handle(true, "Z")
	if nom.Value() != "aZc" || nom.caret != 2 {
		t.Fatalf("écrasement %q caret %d", nom.Value(), nom.caret)
	}
	nom.Handle(true, "insert")
	nom.Handle(true, "Y")
	if nom.Value() != "aZYc" || nom.caret != 3 {
		t.Fatalf("insertion %q caret %d", nom.Value(), nom.caret)
	}

	app := New(DefaultStyle(), Grid(Row(nom)))
	nom.SetText("xy")
	nom.caret = 0
	app = press(t, app, key(tea.KeyDelete))
	if nom.Value() != "y" {
		t.Fatalf("touche %q", nom.Value())
	}
}

func TestZoneEtActions(t *testing.T) {
	zone := TextArea("N", "abc", false, 2, true)
	if zone.Value() != "abc" {
		t.Fatal(zone.Value())
	}
	zone.SetEditable(true)
	src := Memory()
	focus := Memory()
	view := TextView(src, focus, "id")
	mustPanic(t, "framework: zone liée", func() { view.SetEditable(true) })
	mustPanic(t, "framework: exec incomplet", func() { Exec{}.Act() })
	mustPanic(t, "framework: rechargement incomplet", func() { Reload{}.Act() })
	mustPanic(t, "framework: appel incomplet", func() { Call{}.Act() })
}
