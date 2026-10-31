package framework

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestFocusMemoireEtPastille(t *testing.T) {
	uniq := NewListUniq("Teintes", NewItem("Cyan", "Cyan"), NewItem("Magenta", "Magenta"), NewItem("Jaune", "Jaune"))
	pieces := NewListMult("Pièces",
		NewItem("Casque", "Casque").WithDot("#ffee00"),
		NewItem("Veste", "Veste"),
	)
	autre := NewButton("Autre", nil)
	app := New(DefaultStyle(), NewGrid(Row(uniq, pieces), Row(autre)))

	if uniq.Focus() != "Cyan" || pieces.Focus() != "Casque" {
		t.Fatalf("départ %q %q", uniq.Focus(), pieces.Focus())
	}
	raw := app.View().Content
	if !lineHas(raw, "Cyan", "55;55;61") || !lineHas(raw, "Casque", "55;55;61") {
		t.Fatal("le fond marque l'item focalisé de chaque liste")
	}
	if lineHas(raw, "Magenta", "55;55;61") {
		t.Fatal("seule la ligne focalisée est surlignée")
	}

	uniq.SetFocus("Jaune")
	if uniq.Focus() != "Jaune" || uniq.Cursor() != 2 {
		t.Fatalf("set %q", uniq.Focus())
	}
	uniq.SetFocus("absent")
	if uniq.Focus() != "Jaune" {
		t.Fatal("un id absent ne déplace pas")
	}

	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyLeft))
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focus() != uniq || uniq.Focus() != "Jaune" {
		t.Fatalf("mémoire %q", uniq.Focus())
	}

	plainView := plain(app.View().Content)
	casque, veste := itemLineOf(plainView, "Casque"), itemLineOf(plainView, "Veste")
	if !strings.Contains(casque, "●") || strings.Contains(veste, "●") {
		t.Fatalf("pastille\n%s\n%s", casque, veste)
	}
	if strings.Index(casque, "●") < strings.Index(casque, "Casque") {
		t.Fatal("la pastille est à droite")
	}
	if !lineHas(app.View().Content, "Casque", "255;238;0") {
		t.Fatal("couleur de la pastille")
	}
	pieces.SetDot("Casque", "")
	if strings.Contains(itemLineOf(plain(pieces.View(DefaultStyle(), nil, false)), "Casque"), "●") {
		t.Fatal("une couleur vide masque la pastille")
	}
	pieces.SetDot("Casque", "#ff2d55")
	if !lineHas(pieces.View(DefaultStyle(), nil, false), "Casque", "255;45;85") {
		t.Fatal("SetDot change la couleur")
	}

	memo := Memory()
	memo.Set(Data{"id": "1", "data": Data{"id": "cyan", "value": "Cyan", "color": "#00f0ff"}})
	live := ListUniq("Teintes", ReactiveObject(ListItem, memo, map[string]string{
		"id": "id", "label": "value", "dot": "color",
	}))
	if !lineHas(live.View(DefaultStyle(), nil, false), "Cyan", "0;240;255") {
		t.Fatal("la pastille suit la mémoire")
	}
	memo.Set(Data{"id": "1", "data": Data{"color": ""}})
	if strings.Contains(plain(live.View(DefaultStyle(), nil, false)), "●") {
		t.Fatal("une couleur vide dans la mémoire masque la pastille")
	}
	live.SetDot("cyan", "#ffee00")
	if memo.Get("1").Get("data").Get("color").String() != "#ffee00" {
		t.Fatal("SetDot écrit la mémoire")
	}
	if !lineHas(live.View(DefaultStyle(), live, false), "Cyan", "255;238;0") {
		t.Fatal("la liste relit la couleur")
	}

	_, cmd := app.Update(textKey("q"))
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("q quitte quand une liste a le focus")
	}
}

func lineHas(view, label, ink string) bool {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(plain(line), label) && strings.Contains(line, ink) {
			return true
		}
	}
	return false
}

func itemLineOf(view, label string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, label) {
			return line
		}
	}
	return ""
}

func TestPastilleAlignee(t *testing.T) {
	item := ListUniq("T", ListItem("a", "A").WithDot("#ffee00"))
	line := itemLineOf(plain(item.View(DefaultStyle(), nil, false)), "A")
	inner := DefaultStyle().content(DefaultStyle().Width)
	// Le cadre et le padding entourent la ligne : la pastille ferme le contenu.
	dot := strings.LastIndex(line, "●")
	if dot < 0 {
		t.Fatal(line)
	}
	tail := line[dot+len("●"):]
	tail = strings.TrimRight(tail, " │╮╯┐┘")
	if strings.TrimSpace(tail) != "" {
		t.Fatalf("après la pastille %q dans %q", tail, line)
	}
	if lipgloss.Width(line) < inner {
		t.Fatalf("largeur %d", lipgloss.Width(line))
	}
}
