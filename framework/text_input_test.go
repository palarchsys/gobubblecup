package framework

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCurseurRecouvreLeCaractere(t *testing.T) {
	nom := NewTextInput("", InputText, 0, false)
	nom.SetValue("ab")
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
	if !lineHas(raw, "ab", "0;122;204") {
		t.Fatal("le caractère sous le curseur n'est pas marqué")
	}
	nom.caret = 2
	end := plain(nom.View(DefaultStyle(), nom, true))
	j := strings.Index(end, "ab")
	if j < 0 || j+2 >= len(end) || end[j+2] != ' ' {
		t.Fatalf("fin de ligne %q", end)
	}
}

func TestSupprEtInser(t *testing.T) {
	nom := NewTextInput("Nom", InputText, 0, true)
	nom.SetValue("abc")
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

	nom.SetValue("abc")
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

	app := New(DefaultStyle(), NewGrid(Row(nom)))
	nom.SetValue("xy")
	nom.caret = 0
	app = press(t, app, key(tea.KeyDelete))
	if nom.Value() != "y" {
		t.Fatalf("touche %q", nom.Value())
	}
}
