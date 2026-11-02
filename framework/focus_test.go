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
	if cmd != nil {
		t.Fatal("q ne quitte pas")
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

func TestFocusPublie(t *testing.T) {
	lien := Memory()
	uniq := ListUniq("Teintes",
		ListItem("cyan", "Cyan"),
		ListItem("magenta", "Magenta"),
		ListItem("jaune", "Jaune"),
	)
	uniq.Bind(lien, "parent")
	if got := lien.Get("parent").Get("data").Get("value").String(); got != "cyan" {
		t.Fatalf("départ %q", got)
	}
	gen := lien.Generation()
	uniq.Handle(false, "up")
	if lien.Generation() != gen {
		t.Fatal("le même id ne doit pas réécrire")
	}
	uniq.Handle(false, "down")
	if got := lien.Get("parent").Get("data").Get("value").String(); got != "magenta" {
		t.Fatalf("bas %q", got)
	}
	if lien.Generation() == gen {
		t.Fatal("le nouvel id n'a pas été publié")
	}
	gen = lien.Generation()
	uniq.SetFocus("magenta")
	if lien.Generation() != gen {
		t.Fatal("republier le même id change la génération")
	}
	uniq.Handle(false, "enter")
	if uniq.Value() != "magenta" {
		t.Fatalf("valeur %q", uniq.Value())
	}
	gen = lien.Generation()
	uniq.Handle(false, "down")
	if uniq.Value() != "magenta" {
		t.Fatal("Value a suivi le focus")
	}
	if lien.Get("parent").Get("data").Get("value").String() != "jaune" {
		t.Fatal("le focus publié n'est pas le curseur")
	}

	src := Memory()
	src.Set("keep", "oui")
	watcher := ReactiveWatcher(WatcherConfig{
		Watcher: []Store{src},
		Functions: func(args []Store) any {
			if args[0].Get("keep").Get("data").Get("value").String() == "oui" {
				return []Item{ListItem("a", "A"), ListItem("b", "B")}
			}
			return []Item{ListItem("a", "A")}
		},
	})
	live := ListUniq("L", watcher)
	live.SetFocus("b")
	live.Bind(lien, "courant")
	if lien.Get("courant").Get("data").Get("value").String() != "b" {
		t.Fatal("focus initial du réactif")
	}
	src.Set("keep", "non")
	gen = lien.Generation()
	if live.Focus() != "a" {
		t.Fatalf("curseur %q", live.Focus())
	}
	if lien.Generation() != gen {
		t.Fatal("Focus a publié")
	}
	live.View(DefaultStyle(), nil, false)
	if lien.Get("courant").Get("data").Get("value").String() != "a" {
		t.Fatal("le dessin n'a pas publié le nouvel id")
	}

	vide := ListUniq("Vide")
	vide.Bind(lien, "rien")
	if lien.Get("rien").Get("data").Get("value").String() != "" {
		t.Fatal("liste vide")
	}

	pieces := ListMult("Pièces", ListItem("casque", "Casque"), ListItem("veste", "Veste"))
	pieces.Bind(lien, "piece")
	gen = lien.Generation()
	pieces.Handle(false, "enter")
	if strings.Join(pieces.Values(), ",") != "casque" {
		t.Fatalf("coche %v", pieces.Values())
	}
	if lien.Generation() != gen || lien.Get("piece").Get("data").Get("value").String() != "casque" {
		t.Fatal("Entrée a changé le focus")
	}
	pieces.Handle(false, "down")
	if lien.Get("piece").Get("data").Get("value").String() != "veste" {
		t.Fatal("ListMult ne publie pas le focus")
	}

	ro := lien.ReadOnly()
	defer func() {
		if recover() == nil {
			t.Fatal("mémoire en lecture seule")
		}
	}()
	vide.Bind(ro, "bloque")
}

func TestPastilleMot(t *testing.T) {
	erreur := ListUniq("T", ListItem("a", "A").WithDot("error"))
	vue := erreur.View(DefaultStyle(), nil, false)
	if !strings.Contains(plain(vue), "●") || !lineHas(vue, "A", "241;76;76") {
		t.Fatal("le mot error ne dessine pas la pastille")
	}
	jaune := ListUniq("T", ListItem("a", "A").WithDot("#ffee00"))
	if !lineHas(jaune.View(DefaultStyle(), nil, false), "A", "255;238;0") {
		t.Fatal("l'hexadécimal a changé")
	}
	for _, dot := range []string{"nope", ""} {
		item := ListUniq("T", ListItem("a", "A").WithDot(dot))
		if strings.Contains(plain(item.View(DefaultStyle(), nil, false)), "●") {
			t.Fatalf("%q dessine une pastille", dot)
		}
	}
	erreur.Disable()
	vue = erreur.View(DefaultStyle(), nil, false)
	if !strings.Contains(plain(vue), "●") || !lineHas(vue, "A", "107;107;107") {
		t.Fatal("la liste désactivée n'est pas grise")
	}
	if lineHas(vue, "A", "241;76;76") {
		t.Fatal("le mot error reste rouge sur une liste désactivée")
	}
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
