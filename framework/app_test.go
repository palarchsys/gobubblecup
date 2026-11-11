package framework

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func key(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func textKey(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Text: s} }

func press(t *testing.T, app App, msg tea.KeyPressMsg) App {
	t.Helper()
	next, cmd := app.Update(msg)
	if cmd != nil {
		t.Fatalf("commande %s", msg.String())
	}
	return next.(App)
}

func TestDemoRechargerEtEcho(t *testing.T) {
	memo := Memory()
	memo.Set(Data{"id": "1", "data": Data{"id": "cyan", "value": "Cyan"}})
	memo.Set(Data{"id": "2", "data": Data{"id": "magenta", "value": "Magenta"}})
	teintes := ListUniq("Teintes", ReactiveObject(ListItem, memo, map[string]string{"id": "id", "label": "value"}))
	pieces := ListMult("Pièces", ListItem("casque", "Casque"), ListItem("veste", "Veste"))
	catalogue := Category("Catalogue", teintes, pieces)
	nom := TextInput("Nom", InputText, 24, true)
	sortie := TextArea("Sortie", "En attente", false, 4, true)
	saisie := Category("Saisie", nom, sortie)
	vues := TabList(TabItem("fiches", "Fiches"), TabItem("saisie", "Saisie"))
	vues.SetTabs(
		Tab("fiches", Grid(Row(catalogue))),
		Tab("saisie", Grid(Row(saisie))),
	)
	recharger := Button("Recharger", Reload{Target: teintes})
	echo := Button("Echo", Exec{Script: "echo", Args: []string{"bonjour"}, Into: sortie})
	app := New(DefaultStyle(), Grid(
		Row(Title("gobubblecup")),
		Row(vues),
		Row(recharger, echo),
	))

	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyEnter))
	if teintes.Value() != "magenta" {
		t.Fatalf("choix %q", teintes.Value())
	}
	memo.Set("4", "Noir")
	if !strings.Contains(plain(app.View().Content), "Noir") {
		t.Fatal("la teinte ajoutée doit apparaître")
	}

	for i := 0; i < 6 && app.grid.Focused() != recharger; i++ {
		app = press(t, app, key(tea.KeyRight))
		t.Logf("droite %d focus %T", i, app.grid.Focused())
	}
	if app.grid.Focused() != recharger {
		spots := app.grid.spots()
		t.Fatalf("pas sur Recharger, %d spots, focus %T", len(spots), app.grid.Focused())
	}
	app = press(t, app, key(tea.KeyEnter))
	if teintes.Value() != "" {
		t.Fatalf("recharger laisse %q", teintes.Value())
	}
	view := plain(app.View().Content)
	if strings.Contains(view, "Noir") || strings.Contains(view, "✓") || !strings.Contains(view, "Cyan") {
		t.Fatal(view)
	}

	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != echo {
		t.Fatalf("focus %T", app.grid.Focused())
	}
	next, cmd := app.Update(key(tea.KeyEnter))
	app = next.(App)
	if cmd == nil {
		t.Fatal("echo sans commande")
	}
	done, ok := cmd().(ExecDone)
	if !ok {
		t.Fatalf("message %T", cmd())
	}
	next, _ = app.Update(done)
	app = next.(App)
	if sortie.Value() != "bonjour" {
		t.Fatalf("sortie %q err %v", sortie.Value(), done.Err)
	}
	if vues.SelectedID() != "saisie" || app.grid.Focused() != sortie {
		t.Fatalf("onglet %s focus %T", vues.SelectedID(), app.grid.Focused())
	}
	if !strings.Contains(plain(app.View().Content), "bonjour") {
		t.Fatal(plain(app.View().Content))
	}
}

func TestFlechesEntreZonesActives(t *testing.T) {
	uniq := ListUniq("Teintes", ListItem("Cyan", "Cyan"), ListItem("Magenta", "Magenta"))
	mute := ListUniq("Off", ListItem("X", "X"))
	mute.SetActive(false)
	mult := ListMult("Pièces", ListItem("Casque", "Casque"), ListItem("Veste", "Veste"))
	cat := Category("Catalogue", uniq, mute, mult)
	executer := Button("Exécuter", nil)
	recharger := Button("Recharger", nil)
	app := New(DefaultStyle(), Grid(
		Row(Title("gobubblecup")),
		Row(cat, executer),
		Row(recharger),
	))

	if app.grid.Focused() != uniq {
		t.Fatal("le focus ignore le titre et part sur la première zone active")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != executer {
		t.Fatal("droite sort de la catégorie")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != mult {
		t.Fatal("droite suit l'ordre haut-gauche vers droite-bas")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != recharger {
		t.Fatal("droite descend à la rangée suivante")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != uniq {
		t.Fatal("droite reprend au premier élément")
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != recharger {
		t.Fatal("gauche reprend au dernier élément")
	}
	app = press(t, app, key(tea.KeyUp))
	app = press(t, app, key(tea.KeyDown))
	if app.grid.Focused() != recharger {
		t.Fatal("haut et bas ne changent pas de zone")
	}
	if app.grid.Focused() == mute {
		t.Fatal("la liste inactive est sautée")
	}
}

func TestSelectionUniqueEtMultiple(t *testing.T) {
	uniq := ListUniq("Teintes", ListItem("Cyan", "Cyan"), ListItem("Magenta", "Magenta"), ListItem("Jaune", "Jaune"))
	mult := ListMult("Pièces", ListItem("Casque", "Casque"), ListItem("Veste", "Veste"))
	app := New(DefaultStyle(), Grid(Row(Category("Catalogue", uniq, mult))))

	if uniq.Focus() != "Cyan" {
		t.Fatalf("focus initial %q", uniq.Focus())
	}
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyDown))
	if uniq.Cursor() != 2 || uniq.Focus() != "Jaune" || app.grid.Focused() != uniq || app.grid.Inside() {
		t.Fatalf("curseur %d focus %q", uniq.Cursor(), uniq.Focus())
	}
	app = press(t, app, key(tea.KeyEnter))
	item, ok := uniq.Choice()
	if !ok || item.Label != "Jaune" || app.grid.Inside() || app.grid.Focused() != uniq {
		t.Fatalf("choix %#v inside %v", item, app.grid.Inside())
	}
	app = press(t, app, key(tea.KeyEnter))
	if uniq.Value() != "Jaune" {
		t.Fatal("un second entrée garde le choix")
	}
	view := plain(app.View().Content)
	if strings.Contains(view, "›") || !strings.Contains(view, "✓ Jaune") {
		t.Fatal("le choix reste en ✓, sans marqueur de focus")
	}
	titleCol, markCol := -1, -1
	for _, line := range strings.Split(view, "\n") {
		if i := strings.Index(line, "Teintes"); i >= 0 && titleCol < 0 {
			titleCol = lipgloss.Width(line[:i])
		}
		if i := strings.Index(line, "✓"); i >= 0 && markCol < 0 {
			markCol = lipgloss.Width(line[:i])
		}
	}
	if titleCol < 0 || titleCol != markCol {
		t.Fatalf("symbole %d titre %d", markCol, titleCol)
	}
	if !strings.Contains(view, "╭ Catalogue") {
		t.Fatal("le sous-titre est sur le contour")
	}

	held := uniq.Focus()
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != mult || uniq.Focus() != held {
		t.Fatal("la liste garde son item")
	}
	left := plain(app.View().Content)
	if !strings.Contains(left, "✓ Jaune") {
		t.Fatal("le ✓ reste visible hors de la liste")
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != uniq || uniq.Focus() != "Jaune" || uniq.Cursor() != 2 {
		t.Fatalf("retour %q", uniq.Focus())
	}

	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyEnter))
	app = press(t, app, key(tea.KeyUp))
	app = press(t, app, key(tea.KeyEnter))
	if app.grid.Inside() || app.grid.Focused() != mult {
		t.Fatal("entrée coche sans ouvrir la liste")
	}
	labels := make([]string, 0, 2)
	for _, item := range mult.Choices() {
		labels = append(labels, item.Label)
	}
	if strings.Join(labels, ",") != "Casque,Veste" || app.grid.Focused() != mult {
		t.Fatalf("multiple %v", labels)
	}
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyEnter))
	if got := mult.Choices(); len(got) != 1 || got[0].Label != "Casque" || app.grid.Inside() {
		t.Fatalf("retrait %#v", got)
	}
}

func TestChampEtZoneDeTexte(t *testing.T) {
	port := TextInput("Port", InputNumber, 4, true)
	secret := TextInput("Code", InputSecret, 0, true)
	zone := TextArea("Notes", "a\nb\nc\nd\ne", false, 2, true)
	edit := TextArea("Brouillon", "", true, 3, true)
	app := New(DefaultStyle(), Grid(
		Row(port, secret),
		Row(zone, edit),
	))

	for _, ch := range []string{"1", "2", "a", "3", "4", "5"} {
		app = press(t, app, textKey(ch))
	}
	if port.Value() != "1234" || !app.grid.Inside() || app.grid.Focused() != port {
		t.Fatalf("nombre %q inside %v", port.Value(), app.grid.Inside())
	}
	app = press(t, app, key(tea.KeyDown))
	if app.grid.Focused() != port {
		t.Fatal("bas ne quitte plus le champ")
	}
	same := false
	for _, line := range strings.Split(plain(port.View(DefaultStyle(), port, true)), "\n") {
		value := strings.Index(line, "1234")
		label := strings.Index(line, "Port")
		if value >= 0 && label >= 0 && label < value {
			same = true
		}
	}
	if !same {
		t.Fatal("le label est à gauche du champ, sur la même ligne")
	}
	app = press(t, app, key(tea.KeyEnter))
	if !app.grid.Inside() || port.Value() != "1234" {
		t.Fatal("entrée laisse le champ ouvert")
	}

	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != secret {
		t.Fatal("droite au bord passe au champ suivant")
	}
	app = press(t, app, textKey("q"))
	app = press(t, app, textKey("z"))
	if secret.Value() != "qz" || !app.grid.Inside() {
		t.Fatalf("secret %q inside %v", secret.Value(), app.grid.Inside())
	}
	if view := plain(secret.View(DefaultStyle(), secret, true)); strings.Contains(view, "qz") || !strings.Contains(view, "••") {
		t.Fatal(view)
	}
	app = press(t, app, key(tea.KeyLeft))
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != secret {
		t.Fatal("gauche déplace le curseur dans le champ")
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != port {
		t.Fatal("gauche au bord revient au champ précédent")
	}
	for i := 0; i < 4; i++ {
		app = press(t, app, key(tea.KeyRight))
	}
	if app.grid.Focused() != zone {
		t.Fatalf("la file continue vers la zone, focus %T", app.grid.Focused())
	}

	if !strings.Contains(plain(zone.View(DefaultStyle(), nil, false)), "█") {
		t.Fatal("la barre apparaît quand le texte dépasse")
	}
	app = press(t, app, key(tea.KeyEnter))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyDown))
	if zone.Top() != 1 {
		t.Fatalf("défilement %d", zone.Top())
	}
	app = press(t, app, key(tea.KeyEnter))

	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyEnter))
	app = press(t, app, textKey("a"))
	app = press(t, app, key(tea.KeyEnter))
	app = press(t, app, key(tea.KeyEsc))
	if edit.Value() != "a\n" || app.grid.Inside() {
		t.Fatalf("texte %q inside %v", edit.Value(), app.grid.Inside())
	}
}

func TestActions(t *testing.T) {
	liste := ListUniq("Teintes", ListItem("Cyan", "Cyan"), ListItem("Jaune", "Jaune"))
	sortie := TextArea("Sortie", "vide", false, 2, true)
	n := 0
	app := New(DefaultStyle(), Grid(Row(
		liste,
		Button("Echo", Exec{Script: "/bin/echo", Args: []string{"ok"}, Into: sortie}),
		Button("Raz", Reload{Target: liste}),
		Button("Plus", Call{Run: func() { n++ }}),
	)))

	app = press(t, app, key(tea.KeyEnter))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, key(tea.KeyEnter))
	if _, ok := liste.Choice(); !ok {
		t.Fatal("sélection")
	}

	app = press(t, app, key(tea.KeyRight))
	next, cmd := app.Update(key(tea.KeyEnter))
	app = next.(App)
	if cmd == nil {
		t.Fatal("echo")
	}
	done, ok := cmd().(ExecDone)
	if !ok {
		t.Fatalf("message %T", cmd())
	}
	next, cmd = app.Update(done)
	if cmd != nil {
		t.Fatal("suite")
	}
	app = next.(App)
	if sortie.Value() != "ok" {
		t.Fatalf("sortie %q", sortie.Value())
	}

	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyEnter))
	if _, ok := liste.Choice(); ok {
		t.Fatal("rechargement")
	}

	app = press(t, app, key(tea.KeyRight))
	app = press(t, app, key(tea.KeyEnter))
	if n != 1 {
		t.Fatalf("appel %d", n)
	}
}

func TestFenetreDeListe(t *testing.T) {
	liste := ListUniq("Nombres",
		ListItem("un", "Un"),
		ListItem("deux", "Deux"),
		ListItem("trois", "Trois"),
		ListItem("quatre", "Quatre"),
		ListItem("cinq", "Cinq"),
	)
	plein := plain(liste.View(DefaultStyle(), liste, false))
	for _, label := range []string{"Un", "Deux", "Trois", "Quatre", "Cinq"} {
		if !strings.Contains(plein, label) {
			t.Fatalf("sans hauteur, %s manque\n%s", label, plein)
		}
	}

	cadre := DefaultStyle()
	cadre.Height = 6
	haut := plain(liste.View(cadre, liste, true))
	if strings.Contains(haut, "Quatre") || strings.Contains(haut, "Cinq") {
		t.Fatalf("la fenêtre montre trop\n%s", haut)
	}
	if !strings.Contains(haut, "Un") || !strings.Contains(haut, "Trois") {
		t.Fatalf("trois lignes\n%s", haut)
	}
	for range 4 {
		liste.Handle(false, "down")
	}
	liste.Handle(false, "enter")
	bas := liste.View(cadre, liste, true)
	texte := plain(bas)
	if strings.Contains(texte, "Un") || !strings.Contains(texte, "Cinq") {
		t.Fatalf("le dernier n'entre pas\n%s", texte)
	}
	if !lineHas(bas, "Cinq", "55;55;61") || !strings.Contains(itemLineOf(texte, "Cinq"), "✓") {
		t.Fatal("le fond ou le ✓ ne suit pas le curseur")
	}
	if strings.Contains(itemLineOf(texte, "Trois"), "✓") || lineHas(bas, "Trois", "55;55;61") {
		t.Fatal("le fond ou le ✓ reste sur un rang de la fenêtre")
	}

	args := []ListArg{ListItem("un", "Un")}
	for i := 0; i < 20; i++ {
		args = append(args, ListItem(fmt.Sprintf("n%d", i), fmt.Sprintf("N%d", i)))
	}
	longue := ListUniq("Longue", args...)
	zone := TextArea("Notes", "ligne-de-zone", false, 4, true)
	app := New(DefaultStyle(), Grid(Row(longue), Row(zone)))
	next, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = next.(App)
	vue := app.View().Content
	if n := lineCount(vue); n > 24 {
		t.Fatalf("la vue dépasse l'écran %d", n)
	}
	if !strings.Contains(plain(vue), "ligne-de-zone") || strings.Contains(plain(vue), "N14") {
		t.Fatal("la zone ou la liste ne tient pas dans l'écran")
	}

	courte := ListUniq("Courte", ListItem("un", "Un"), ListItem("deux", "Deux"))
	naturel := New(DefaultStyle(), Grid(Row(courte)))
	next, _ = naturel.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	naturel = next.(App)
	if n := lineCount(naturel.View().Content); n >= 40 || !strings.Contains(plain(naturel.View().Content), "Deux") {
		t.Fatalf("le contenu court s'étend %d", n)
	}

	format := DefaultStyle()
	format.Height = format.Width * 3 / 4
	cadreApp := New(format, Grid(Row(longue)))
	next, _ = cadreApp.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	cadreApp = next.(App)
	vue = cadreApp.View().Content
	if n := lineCount(vue); n != format.Height {
		t.Fatalf("format %d, obtenu %d", format.Height, n)
	}
	if strings.Contains(plain(vue), "N14") || !strings.Contains(plain(vue), "Un") {
		t.Fatal("le format ne borne pas la liste")
	}
}

func TestRemplissage(t *testing.T) {
	g := Grid(
		Row(Title("gobubblecup")),
		Row(ListUniq("Large", ListItem("Cyan", "Cyan"), ListItem("Magenta", "Magenta")), Button("Autre", nil)),
		Row(Button("Ok", nil)),
	)
	parts := g.render(DefaultStyle())
	edge := rightBorder(parts[0])
	for i, part := range parts {
		if got := rightBorder(part); got != edge {
			t.Fatalf("ligne %d bord %d, attendu %d", i, got, edge)
		}
	}

	zone := TextArea("Notes", "a\nb\nc\nd\ne\nf\ng\nh", false, 6, true)
	row := Grid(Row(zone, Button("Ok", nil))).render(DefaultStyle())[0]
	lines := strings.Split(plain(row), "\n")
	last := bottomLine(lines)
	if last < 0 || strings.Count(lines[last], "╰") != 2 {
		t.Fatalf("hauteurs non alignées, ligne %d", last)
	}

	liste := ListUniq("A", ListItem("un", "un"))
	cat := Grid(Row(Category("Boîte", liste), zone)).render(DefaultStyle())[0]
	lines = strings.Split(plain(cat), "\n")
	last = bottomLine(lines)
	if last < 0 || strings.Count(lines[last], "╰") != 2 {
		t.Fatalf("catégorie %d", last)
	}
	inner := false
	for i := last - 1; i >= 0 && i >= last-3; i-- {
		if strings.Contains(lines[i], "╰") {
			inner = true
		}
	}
	if !inner {
		t.Fatal("le dernier enfant ne descend pas")
	}
}

func TestCadre(t *testing.T) {
	s := DefaultStyle()
	if s.Radius != 1 || s.Thickness != 1 || s.PadX != 1 || s.PadY != 0 || s.Margin != 1 {
		t.Fatalf("mesures %#v", s)
	}
	if s.Background != s.Surface || s.Background != Color("#1e1e1e") {
		t.Fatal("le fond et les zones diffèrent")
	}
	if s.Text != Color("#cccccc") || s.Accent != Color("#e5b684") || s.Idle != Color("#e5b684") || s.Focus != Color("#44aae6") || s.Inside != Color("#44aae6") || s.Hint != Color("#8a8a8a") {
		t.Fatal("encres")
	}
	if ink, ok := statusInk("pending"); !ok || ink != Color("#6b6b6b") {
		t.Fatal("pending")
	}
	if ink, ok := statusInk("success"); !ok || ink != Color("#35cc94") {
		t.Fatal("succès")
	}
	if ink, ok := statusInk("warning"); !ok || ink != Color("#ff8700") {
		t.Fatal("avertissement")
	}
	btn := Button("Ok", nil)
	repos := btn.View(s, nil, false)
	if !strings.Contains(repos, "229;182;132") || strings.Contains(repos, "236;203;168") {
		t.Fatal("contour au repos")
	}
	vif := btn.View(s, btn, false)
	if !strings.Contains(vif, "236;203;168") || strings.Contains(vif, "68;170;230") {
		t.Fatal("contour au focus")
	}
	custom := s
	custom.Idle = Color("#000000")
	if !strings.Contains(btn.View(custom, btn, false), "76;76;76") {
		t.Fatal("la surbrillance ne suit pas Idle")
	}
	btn.Disable()
	eteint := btn.View(s, btn, false)
	if !strings.Contains(eteint, "107;107;107") || strings.Contains(eteint, "229;182;132") {
		t.Fatal("désactivé")
	}
	view := Button("Ok", nil).View(s, nil, false)
	want := 1 + 2*s.PadY + 2
	if got := strings.Count(view, "\n") + 1; got != want {
		t.Fatalf("lignes %d, attendu %d", got, want)
	}
	if !strings.Contains(plain(view), "╭") {
		t.Fatal("coins arrondis")
	}
	square := s
	square.Radius = 0
	if !strings.Contains(plain(Button("Ok", nil).View(square, nil, false)), "┌") {
		t.Fatal("coins droits")
	}

	placed := s.placed(30, 12)
	view = Button("Ok", nil).View(placed, nil, false)
	if got := strings.Count(view, "\n") + 1; got != 12 {
		t.Fatalf("hauteur imposée %d", got)
	}
	if rightBorder(view) != 30 {
		t.Fatalf("largeur imposée %d", rightBorder(view))
	}

	text := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk"
	tall := plain(TextArea("Notes", text, false, 2, true).View(s.placed(22, 20), nil, false))
	if strings.Contains(tall, "█") || !strings.Contains(tall, "k") {
		t.Fatal("la zone doit montrer tout le texte")
	}
	short := plain(TextArea("Notes", text, false, 2, true).View(s.placed(22, 8), nil, false))
	if !strings.Contains(short, "█") || strings.Contains(short, "k") {
		t.Fatal("la barre reste quand le texte dépasse la fenêtre")
	}

	app := New(s, Grid(Row(Button("Ok", nil))))
	if app.View().BackgroundColor != s.paper() {
		t.Fatal("fond uni")
	}
	lines := strings.Split(plain(New(s, Grid(
		Row(Title("gobubblecup")),
		Row(Button("Ok", nil)),
	)).View().Content), "\n")
	bottom, next := -1, -1
	for i, line := range lines {
		if bottom < 0 && strings.Contains(line, "╰") {
			bottom = i
			continue
		}
		if bottom >= 0 && strings.Contains(line, "╭") {
			next = i
			break
		}
	}
	if next-bottom != 1 {
		t.Fatalf("écart entre zones %d", next-bottom-1)
	}
}

func bottomLine(lines []string) int {
	last := -1
	for i, line := range lines {
		if strings.Contains(line, "╰") {
			last = i
		}
	}
	return last
}

func rightBorder(view string) int {
	best := 0
	for _, line := range strings.Split(plain(view), "\n") {
		col := 0
		last := 0
		for _, r := range line {
			col++
			switch r {
			case '╮', '┐', '╯', '┘':
				last = col
			}
		}
		if last > best {
			best = last
		}
	}
	return best
}

func TestOnglets(t *testing.T) {
	fiches := ListUniq("Fiches", ListItem("Une", "Une"))
	notes := Button("ZoneNotes", nil)
	vues := TabList(TabItem("fiches", "Fiches"), TabItem("notes", "Notes"))
	vues.SetTabs(
		Tab("fiches", Grid(Row(fiches))),
		Tab("notes", Grid(Row(notes))),
	)
	bas := Button("Bas", nil)
	app := New(DefaultStyle(), Grid(Row(vues), Row(bas)))

	if app.grid.Focused() != fiches || app.grid.Inside() || vues.SelectedID() != "fiches" {
		t.Fatal("le focus part sur la première zone de l'onglet")
	}
	view := plain(app.View().Content)
	if !strings.Contains(view, "Une") || strings.Contains(view, "ZoneNotes") {
		t.Fatal("le premier onglet est affiché")
	}
	if strings.Contains(view, "Vues") || !strings.Contains(view, "┬") {
		t.Fatal("onglets collés, sans titre de liste")
	}
	if !strings.Contains(view, "onglet") || !strings.Contains(view, "échap") || !strings.Contains(view, "quitte") {
		t.Fatal("la légende liste les contrôles")
	}

	app = press(t, app, key(tea.KeyEnter))
	if app.grid.Inside() || app.grid.Focused() != fiches || fiches.Value() != "Une" {
		t.Fatal("entrée choisit l'item sans ouvrir la liste")
	}
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyTab})
	if vues.Cursor() != 1 || vues.SelectedID() != "notes" || app.grid.Focused() != notes || app.grid.Inside() {
		t.Fatalf("curseur %d onglet %s", vues.Cursor(), vues.SelectedID())
	}
	view = plain(app.View().Content)
	if strings.Contains(view, "Une") || !strings.Contains(view, "ZoneNotes") {
		t.Fatal(view)
	}
	lines := strings.Split(view, "\n")
	joined := false
	for i, line := range lines {
		if !strings.Contains(line, "├") || i+1 >= len(lines) {
			continue
		}
		if strings.Contains(lines[i+1], "│") {
			joined = true
		}
	}
	if !joined {
		t.Fatal("le premier onglet fermé rejoint le panneau")
	}
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyTab})
	if vues.SelectedID() != "fiches" || app.grid.Focused() != fiches || app.grid.Inside() {
		t.Fatal("tab en bout de liste revient au premier onglet")
	}
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyTab})
	if vues.SelectedID() != "notes" || app.grid.Focused() != notes || app.grid.Inside() {
		t.Fatal("tab repart vers l'onglet suivant")
	}
	app = press(t, app, key(tea.KeyEnter))
	if vues.SelectedID() != "notes" || strings.Contains(plain(app.View().Content), "Une") {
		t.Fatal("entrée ne change pas d'onglet")
	}
	app = press(t, app, key(tea.KeyRight))
	if app.grid.Focused() != bas {
		t.Fatal("droite sort de l'onglet")
	}
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyTab})
	if vues.SelectedID() != "fiches" || app.grid.Focused() != fiches || app.grid.Inside() {
		t.Fatal("tab depuis une autre zone ouvre l'onglet suivant")
	}
}

func TestOngletsReactifs(t *testing.T) {
	memo := Memory()
	memo.Set("fiches", "Fiches")
	memo.Set("notes", "Notes")
	src := ReactiveObject(TabItem, memo, map[string]string{"id": "id", "label": "value"})
	fiches := Button("Panneau", nil)
	notes := Button("ZoneNotes", nil)
	plus := Button("ZonePlus", nil)
	vues := TabList(src)
	vues.SetTabs(
		Tab("fiches", Grid(Row(fiches))),
		Tab("notes", Grid(Row(notes))),
		Tab("plus", Grid(Row(plus))),
	)
	app := New(DefaultStyle(), Grid(Row(vues)))
	view := plain(app.View().Content)
	if !strings.Contains(view, "Fiches") || !strings.Contains(view, "Notes") || strings.Contains(view, "Plus") {
		t.Fatal(view)
	}
	if vues.SelectedID() != "fiches" || app.grid.Focused() != fiches {
		t.Fatal("le premier onglet reste affiché")
	}
	memo.Set("fiches", "Atelier")
	memo.Set("plus", "Plus")
	view = plain(app.View().Content)
	if !strings.Contains(view, "Atelier") || !strings.Contains(view, "Plus") || strings.Contains(view, "Fiches") {
		t.Fatal(view)
	}
	if vues.SelectedID() != "fiches" || app.grid.Focused() != fiches {
		t.Fatalf("onglet %s", vues.SelectedID())
	}
	app = press(t, app, key(tea.KeyTab))
	app = press(t, app, key(tea.KeyTab))
	if vues.SelectedID() != "plus" || app.grid.Focused() != plus {
		t.Fatalf("nouvel onglet %s", vues.SelectedID())
	}
	src.Reload()
	view = plain(app.View().Content)
	if strings.Contains(view, "Plus") || strings.Contains(view, "Atelier") || !strings.Contains(view, "Fiches") {
		t.Fatal(view)
	}
	if vues.SelectedID() != "fiches" {
		t.Fatalf("retour %s", vues.SelectedID())
	}
}

func TestUneSeuleListeDOnglets(t *testing.T) {
	mustPanic := func(grid *grid) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("deux listes d'onglets")
			}
		}()
		New(DefaultStyle(), grid)
	}
	mustPanic(Grid(
		Row(TabList(TabItem("a", "A")), TabList(TabItem("b", "B"))),
	))
	inner := TabList(TabItem("b", "B"))
	outer := TabList(TabItem("a", "A"))
	outer.SetTabs(Tab("a", Grid(Row(inner))))
	mustPanic(Grid(Row(outer)))
}

func TestOngletDansLaLargeur(t *testing.T) {
	cat := Category("Catalogue",
		ListUniq("Teintes", ListItem("Cyan", "Cyan")),
		ListMult("Pièces", ListItem("Casque", "Casque")),
	)
	vues := TabList(TabItem("fiches", "Fiches"), TabItem("notes", "Notes"))
	vues.SetTabs(
		Tab("fiches", Grid(Row(cat))),
		Tab("notes", Grid(Row(TextArea("Sortie", "En attente", false, 4, true)))),
	)
	parts := Grid(
		Row(Title("gobubblecup")),
		Row(vues),
		Row(TextInput("Nom", InputText, 24, true), Button("Exécuter", nil), Button("Recharger", nil)),
	).render(DefaultStyle())
	edge := rightBorder(parts[2])
	if got := rightBorder(parts[1]); got != edge {
		t.Fatalf("onglet %d, interface %d", got, edge)
	}
	for _, line := range strings.Split(parts[1], "\n") {
		if w := lipgloss.Width(line); w > edge {
			t.Fatalf("ligne d'onglet %d, max %d", w, edge)
		}
		plainLine := plain(line)
		if !strings.Contains(plainLine, "Catalogue") {
			continue
		}
		col := 0
		last := 0
		for _, r := range plainLine {
			col++
			if r == '╮' || r == '┐' {
				last = col
			}
		}
		if last != edge-2 {
			t.Fatalf("la grille interne s'arrête à %d, intérieur %d", last, edge-2)
		}
	}
}

func TestEcranEfface(t *testing.T) {
	app := New(DefaultStyle(), Grid(Row(Button("Ok", nil))))
	if !app.View().AltScreen {
		t.Fatal("le lancement ne prend pas l'écran alternatif")
	}
}

func TestQuitter(t *testing.T) {
	champ := TextInput("Nom", InputText, 0, true)
	app := New(DefaultStyle(), Grid(Row(champ)))
	app = press(t, app, textKey("a"))
	if !app.grid.Inside() {
		t.Fatal("le champ est ouvert")
	}
	app = press(t, app, textKey("q"))
	if champ.Value() != "aq" {
		t.Fatalf("q saisit %q", champ.Value())
	}

	dir := t.TempDir()
	writeScript(t, dir, "un.sh", "echo success\n")
	writeScript(t, dir, "deux.sh", "echo deja\nsleep 30\n")
	writeScript(t, dir, "trois.sh", "echo success\n")
	makeRT := func() *runtime {
		return Runtime(RuntimeConfig{
			Dir: dir,
			Functions: func(Store) {
				Bash("un", "Un", "un.sh", nil)
				Bash("deux", "Deux", "deux.sh", nil)
				Bash("trois", "Trois", "trois.sh", nil)
			},
		})
	}
	un := makeRT()
	deux := makeRT()
	un.Start().Act()
	deux.Start().Act()
	waitReturn(t, un, "deux", "deja")
	waitReturn(t, deux, "deux", "deja")

	var cmd tea.Cmd
	callWithin(t, 2*time.Second, func() {
		_, cmd = app.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	})
	if cmd == nil {
		t.Fatal("ctrl+c")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c quitte")
	}
	for _, rt := range []*runtime{un, deux} {
		if rt.Status() != "terminated" {
			t.Fatalf("cycle %q", rt.Status())
		}
		if rt.Memory.Get("un").Get("status").String() != "success" {
			t.Fatal("le premier a changé")
		}
		if rt.Memory.Get("deux").Get("status").String() != "error" {
			t.Fatalf("second %q", rt.Memory.Get("deux").Get("status").String())
		}
		if !strings.Contains(rt.Memory.Get("deux").Get("return").String(), "deja") {
			t.Fatal("le texte déjà reçu a disparu")
		}
		if rt.Memory.Get("trois").Get("status").String() != "" {
			t.Fatal("le troisième n'est pas vide")
		}
	}
}
