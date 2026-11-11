package framework

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDesactive(t *testing.T) {
	alpha := Button("Alpha", nil)
	beta := Button("Beta", nil)
	app := New(DefaultStyle(), Grid(Row(alpha, beta)))
	if app.grid.Focused() != alpha {
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
	if app.grid.Focused() == alpha || !alpha.Disabled() {
		t.Fatal("le focus quitte l'objet désactivé")
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() == alpha {
		t.Fatal("gauche ne revient pas sur un objet désactivé")
	}
	alpha.Enable()
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != alpha || alpha.Disabled() {
		t.Fatal("enable rend le focus")
	}

	n := 0
	inner := Button("Dedans", Call{Run: func() { n++ }})
	cat := Category("Boîte", inner)
	out := Button("Dehors", nil)
	app = New(DefaultStyle(), Grid(Row(cat), Row(out)))
	cat.Disable()
	app.View()
	if !inner.Disabled() || app.grid.Focused() == inner {
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
