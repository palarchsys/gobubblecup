package framework

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDesactive(t *testing.T) {
	alpha := NewButton("Alpha", nil)
	beta := NewButton("Beta", nil)
	app := New(DefaultStyle(), NewGrid(Row(alpha, beta)))
	if app.grid.Focus() != alpha {
		t.Fatal("départ")
	}
	alpha.Disable()
	raw := app.View().Content
	if !strings.Contains(plain(raw), "Alpha") {
		t.Fatal("l'objet désactivé reste visible")
	}
	if !strings.Contains(raw, "107;107;107") {
		t.Fatal("l'objet désactivé est gris")
	}
	if app.grid.Focus() == alpha || !alpha.Disabled() {
		t.Fatal("le focus quitte l'objet désactivé")
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focus() == alpha {
		t.Fatal("gauche ne revient pas sur un objet désactivé")
	}
	alpha.Enable()
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focus() != alpha || alpha.Disabled() {
		t.Fatal("enable rend le focus")
	}

	n := 0
	inner := NewButton("Dedans", Call{Run: func() { n++ }})
	cat := NewCategory("Boîte", inner)
	out := NewButton("Dehors", nil)
	app = New(DefaultStyle(), NewGrid(Row(cat), Row(out)))
	cat.Disable()
	app.View()
	if !inner.Disabled() || app.grid.Focus() == inner {
		t.Fatal("la catégorie désactivée retire ses enfants du focus")
	}
	app = press(t, app, key(tea.KeyEnter))
	if n != 0 {
		t.Fatal("l'enfant désactivé ne s'active pas")
	}
	if !strings.Contains(plain(app.View().Content), "Dedans") {
		t.Fatal("l'enfant reste visible")
	}
	cat.Enable()
	if inner.Disabled() {
		t.Fatal("enable rend les enfants")
	}
}
