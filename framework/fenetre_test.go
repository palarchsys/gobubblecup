package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestFenetreEtSaisie(t *testing.T) {
	nom := TextInput("Nom", InputText, 0, true)
	code := TextInput("Code", InputSecret, 0, false)
	note := TextInput("Note", InputText, 0, true)
	w := Window()
	suivante := Button("Suivante", w.Next())
	precedente := Button("Précédente", w.Prev())
	page := Grid(Row(nom), Row(code), Row(suivante))
	autre := Grid(Row(note), Row(precedente))
	w.Use(page, autre)
	app := NewWindow(DefaultStyle(), w)

	app = press(t, app, textKey("a"))
	if nom.Value() != "a" || !app.grid.Inside() {
		t.Fatalf("saisie directe %q", nom.Value())
	}
	haut := lineCount(nom.View(DefaultStyle(), nil, false))
	bas := lineCount(code.View(DefaultStyle(), nil, false))
	if bas >= haut {
		t.Fatalf("sans cadre %d, avec cadre %d", bas, haut)
	}
	if strings.Contains(plain(code.View(DefaultStyle(), nil, false)), "╭") {
		t.Fatal("le champ sans cadre a un contour")
	}

	next, cmd := app.Update(tea.PasteMsg{Content: "b\n2"})
	app = next.(App)
	if cmd != nil || nom.Value() != "ab2" {
		t.Fatalf("collage %q", nom.Value())
	}

	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != code {
		t.Fatal("droite quitte le champ au bord")
	}
	next, cmd = app.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	app = next.(App)
	if cmd == nil {
		t.Fatal("ctrl+v demande le presse-papier")
	}
	_ = cmd()
	next, _ = app.Update(tea.ClipboardMsg{Content: "zz", Selection: 'c'})
	app = next.(App)
	if code.Value() != "zz" || strings.Contains(plain(code.View(DefaultStyle(), code, true)), "zz") {
		t.Fatalf("secret collé %q", code.Value())
	}

	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != suivante {
		t.Fatal("le bouton suivante")
	}
	app = press(t, app, key(tea.KeyEnter))
	if w.Index() != 1 || app.grid != autre || app.grid.Focused() != note {
		t.Fatalf("fenêtre %d", w.Index())
	}
	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyEnter))
	if w.Index() != 0 || app.grid != page || app.grid.Focused() != nom || !app.grid.Inside() {
		t.Fatalf("retour %d", w.Index())
	}
	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != suivante {
		t.Fatal("le bouton suivante, au second tour")
	}
	app = press(t, app, key(tea.KeyEnter))
	if w.Index() != 1 || app.grid.Focused() != note || !app.grid.Inside() {
		t.Fatal("l'arrivée repose le focus sur le premier élément")
	}
}

func TestFichierEtBarre(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.txt")
	memo := Memory()
	memo.Set("file", "bonjour")
	if cmd := (WriteFile{Path: path, Memory: memo}).Act(); cmd != nil {
		t.Fatal("écriture")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "bonjour" {
		t.Fatalf("fichier %q %v", got, err)
	}
	memo.Set("file", "vide")
	if cmd := (ReadFile{Path: path, Memory: memo}).Act(); cmd != nil {
		t.Fatal("lecture")
	}
	if memo.Get("file").Get("data").Get("value").String() != "bonjour" {
		t.Fatal(memo.Get("file").Get("data").Get("value").String())
	}

	a, b := Variable(), Variable()
	if a.Value() != 0 {
		t.Fatal("défaut")
	}
	liste := Variables(a, b)
	barre := ProgressBar(liste)
	if !strings.Contains(plain(barre.View(DefaultStyle(), nil, false)), "0%") {
		t.Fatal(plain(barre.View(DefaultStyle(), nil, false)))
	}
	a.Set(1)
	if a.Value() != 1 || !strings.Contains(plain(barre.View(DefaultStyle(), nil, false)), "50%") {
		t.Fatal(plain(barre.View(DefaultStyle(), nil, false)))
	}
	b.Set(1)
	view := plain(barre.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "100%") {
		t.Fatal(view)
	}
	hors := Variable()
	hors.Set(2)
	liste.Add(hors)
	view = plain(barre.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "100%") {
		t.Fatal(view)
	}
	c := Variable()
	liste.Add(c)
	view = plain(barre.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "66%") {
		t.Fatal(view)
	}
	liste.Set(a)
	view = plain(barre.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "100%") {
		t.Fatal(view)
	}
}
