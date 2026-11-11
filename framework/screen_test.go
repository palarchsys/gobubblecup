package framework

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestBindKeepsDefaultQuit(t *testing.T) {
	app := New(DefaultStyle(), NewGrid(NewRow(NewTextInput("a", InputText, 0, false))))
	_, cmd := app.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+c")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c quitte")
	}
	view := plain(app.View().Content)
	if !strings.Contains(view, "quitte") || !strings.Contains(view, "colle") {
		t.Fatal("légende")
	}
}

func TestBindPasteAndHiddenLegend(t *testing.T) {
	field := NewTextInput("a", InputText, 0, false)
	app := New(DefaultStyle(), NewGrid(NewRow(field)))
	app = app.Bind(KeyMap{Quit: "ctrl+q", Paste: "ctrl+c", HideLegend: true})
	if strings.Contains(plain(app.View().Content), "quitte") {
		t.Fatal("légende stock")
	}
	next, cmd := app.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("colle")
	}
	if _, ok := cmd().(tea.QuitMsg); ok {
		t.Fatal("ctrl+c quitte")
	}
	app = next.(App)
	next, cmd = app.Update(tea.ClipboardMsg{Content: "abc", Selection: 'c'})
	if cmd != nil {
		t.Fatal(cmd)
	}
	app = next.(App)
	if field.Value() != "abc" {
		t.Fatalf("collé %q", field.Value())
	}
	_, cmd = app.Update(tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+q")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+q quitte")
	}
}

func TestBindOn(t *testing.T) {
	field := NewTextInput("a", InputText, 0, false)
	hit := ""
	app := New(DefaultStyle(), NewGrid(NewRow(field))).Bind(KeyMap{
		On: map[string]func() tea.Cmd{
			"ctrl+l": func() tea.Cmd { hit = "l"; return nil },
			"ctrl+i": func() tea.Cmd { return tea.Quit },
			"ctrl+u": nil,
		},
	})
	next, cmd := app.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})
	if hit != "l" || cmd != nil {
		t.Fatal("ctrl+l")
	}
	app = next.(App)
	_, cmd = app.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	if cmd != nil || field.Value() != "" {
		t.Fatal("ctrl+u")
	}
	app = press(t, app, textKey("x"))
	if field.Value() != "x" {
		t.Fatalf("touche libre %q", field.Value())
	}
	_, cmd = app.Update(tea.KeyPressMsg{Code: 'i', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+i")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+i quitte")
	}
}

func TestVerticalFieldsMoveFocus(t *testing.T) {
	one := NewTextInput("un", InputText, 0, false)
	two := NewTextInput("deux", InputText, 0, false)
	grid := NewGrid(NewRow(one), NewRow(two))
	app := New(DefaultStyle(), grid)
	app = press(t, app, textKey("a"))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, textKey("b"))
	if one.Value() != "ab" || two.Value() != "" {
		t.Fatalf("défaut %q %q", one.Value(), two.Value())
	}
	one.SetText("")
	grid.SetVerticalFields(true)
	app = New(DefaultStyle(), grid)
	app = press(t, app, textKey("a"))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, textKey("b"))
	if one.Value() != "a" || two.Value() != "b" {
		t.Fatalf("vertical %q %q", one.Value(), two.Value())
	}
	app = press(t, app, key(tea.KeyLeft))
	if app.grid.Focused() != two {
		t.Fatal("gauche sort du champ")
	}
	app = press(t, app, key(tea.KeyUp))
	if app.grid.Focused() != one {
		t.Fatal("haut")
	}
}

func TestSetTitleKeepsValue(t *testing.T) {
	in := NewTextInput("ancien", InputText, 0, false)
	in.SetText("secret")
	in.SetTitle("nouveau")
	if in.Value() != "secret" {
		t.Fatalf("valeur %q", in.Value())
	}
	view := plain(in.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "nouveau") || strings.Contains(view, "ancien") {
		t.Fatal(view)
	}
	cat := NewCategory("Ancien", in)
	cat.SetLegend("Courrier")
	if !strings.Contains(plain(cat.View(DefaultStyle(), nil, false)), "Courrier") {
		t.Fatal("titre")
	}
}

func TestGraySkipsDisable(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	IncludeImage("probe-red", buf.Bytes())
	p := NewImage("probe-red", 2)
	p.SetBorder(false)
	colorView := p.View(DefaultStyle(), nil, false)
	if strings.Contains(plain(colorView), "╭") {
		t.Fatal("cadre")
	}
	if !strings.Contains(colorView, "255;0;0") {
		t.Fatal(colorView)
	}
	p.SetGray(true)
	if p.Disabled() || !p.Active() {
		t.Fatal("gris inactif")
	}
	grayView := p.View(DefaultStyle(), nil, false)
	if !strings.Contains(grayView, "76;76;76") || strings.Contains(grayView, "255;0;0") {
		t.Fatal(grayView)
	}
	p.Disable()
	washed := p.View(DefaultStyle(), nil, false)
	if !strings.Contains(washed, "107;107;107") {
		t.Fatal(washed)
	}
}

func TestRowEndPushesTheTailRight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	IncludeImage("probe-block", buf.Bytes())
	a := NewImage("probe-block", 4)
	b := NewImage("probe-block", 4)
	a.SetBorder(false)
	b.SetBorder(false)
	a.SetActive(false)
	b.SetActive(false)
	s := DefaultStyle()
	s.Width = 40
	rows := NewGrid(NewRow(NewTitle("Liltress")), RowEnd(a, b)).layout(s, 0, false)
	plainRow := plain(rows[1])
	if lineWidth(rows[1]) != 40 {
		t.Fatalf("largeur %d", lineWidth(rows[1]))
	}
	lead := len(plainRow) - len(strings.TrimLeft(plainRow, " "))
	if lead < 10 {
		t.Fatalf("gauche %d\n%s", lead, plainRow)
	}
	tail := len(plainRow) - len(strings.TrimRight(plainRow, " "))
	if tail > 2 {
		t.Fatalf("droite %d", tail)
	}
}

func TestFrameAndStatus(t *testing.T) {
	s := DefaultStyle()
	s.Width = 30
	centered := plain(Frame("Liltress", AlignCenter, "").View(s, nil, false))
	left := plain(Frame("Liltress", AlignLeft, "Bandeau").View(s, nil, false))
	if !strings.Contains(left, "Bandeau") || strings.Contains(centered, "Bandeau") {
		t.Fatal("titre")
	}
	cLine := lineWith(centered, "Liltress")
	lLine := lineWith(left, "Liltress")
	if strings.Index(cLine, "Liltress") <= strings.Index(lLine, "Liltress") {
		t.Fatalf("centrage\n%s\n%s", cLine, lLine)
	}
	long := plain(Frame(strings.Repeat("x", 80), AlignLeft, "").View(s, nil, false))
	if strings.Contains(long, strings.Repeat("x", 40)) {
		t.Fatal("coupe")
	}
	line := Status("Status : ", "pending...")
	st := DefaultStyle()
	st.Width = 40
	view := plain(line.View(st, nil, false))
	if !strings.Contains(view, "Status : pending...") {
		t.Fatal(view)
	}
	if strings.Index(lineWith(view, "Status"), "Status") < 8 {
		t.Fatal("droite")
	}
	line.SetText("ok")
	view = plain(line.View(st, nil, false))
	if !strings.Contains(view, "Status : ok") || strings.Contains(view, "pending") {
		t.Fatal(view)
	}
}

func TestFormAndLabels(t *testing.T) {
	name := Field("LOGIN", Text)
	secret := Field("PASSWORD", Secret)
	form := Form("Courrier", name, secret)
	grid := NewGrid(NewRow(form))
	app := New(DefaultStyle(), grid)
	app = press(t, app, textKey("a"))
	app = press(t, app, key(tea.KeyDown))
	app = press(t, app, textKey("b"))
	if got := form.Values(); got[0] != "a" || got[1] != "b" {
		t.Fatalf("valeurs %q", got)
	}
	app = press(t, app, key(tea.KeyDown))
	if app.grid.Focused() != name.input {
		t.Fatal("boucle")
	}
	form.SetLabels(map[string]string{"LOGIN": "Nom", "PASSWORD": "Mot"})
	if form.Values()[0] != "a" {
		t.Fatal("valeur perdue")
	}
	view := plain(form.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "Nom :") || !strings.Contains(view, "Mot :") || strings.Contains(view, "b") {
		t.Fatal(view)
	}
	form.SetTitle("Mail")
	if !strings.Contains(plain(form.View(DefaultStyle(), nil, false)), "Mail") {
		t.Fatal("titre")
	}
	dir := t.TempDir()
	for _, lang := range []string{"en", "fr", "de", "es"} {
		writeLabel(t, dir, lang, "[LABEL_LOGIN]\nNom\n\n[LABEL_PASSWORD]\nMot\n")
	}
	labels, err := Labels(dir)
	if err != nil {
		t.Fatal(err)
	}
	if labels["fr"]["LOGIN"] != "Nom" || labels["fr"]["LABEL_LOGIN"] != "" {
		t.Fatalf("%v", labels["fr"])
	}
	form.Use(labels, "fr")
	if form.Values()[1] != "b" {
		t.Fatal("use")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("langue")
			}
		}()
		form.Use(labels, "it")
	}()
}

func TestLabelsRejectsALooseLine(t *testing.T) {
	dir := t.TempDir()
	for _, lang := range []string{"en", "fr", "de", "es"} {
		writeLabel(t, dir, lang, "[LABEL_LOGIN]\nNom\n")
	}
	loose := filepath.Join(dir, "de", "label.txt")
	if err := os.WriteFile(loose, []byte("hors\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Labels(dir); err == nil {
		t.Fatal("ligne hors clé")
	}
}

func TestSVGGlyphAndFlags(t *testing.T) {
	fr := []byte(`<svg width="90" height="60"><rect width="30" height="60" fill="#0055A4"/><rect x="30" width="30" height="60" fill="#FFFFFF"/><rect x="60" width="30" height="60" fill="#EF4135"/></svg>`)
	en := []byte(`<svg width="90" height="60"><defs><clipPath id="en"><rect width="90" height="60"/></clipPath></defs><g><rect width="90" height="60" fill="#012169"/><path d="M0 0 L90 60 M90 0 L0 60" stroke="#FFFFFF" stroke-width="4"/><path d="M45 0 V60 M0 30 H90" stroke="#C8102E" stroke-width="4"/></g></svg>`)
	if err := IncludeSVG("fr.svg", fr); err != nil {
		t.Fatal(err)
	}
	if err := IncludeSVG("bad", []byte(`<svg width="10" height="10"><circle r="2"/></svg>`)); err == nil {
		t.Fatal("élément")
	}
	if err := IncludeSVG("curve", []byte(`<svg width="10" height="10"><path d="M0 0 C1 1 2 2 3 3" stroke="#000000" stroke-width="1"/></svg>`)); err == nil {
		t.Fatal("ordre")
	}
	glyph := Glyph("fr.svg", 6)
	if glyph.Active() || strings.Contains(plain(glyph.View(DefaultStyle(), nil, false)), "╭") {
		t.Fatal("glyphe")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fr.svg"), fr, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "en.svg"), en, 0o644); err != nil {
		t.Fatal(err)
	}
	flags, err := Flags(dir, []string{"en", "fr"}, 6)
	if err != nil {
		t.Fatal(err)
	}
	if flags.Selected() != "" || flags.Active() {
		t.Fatal("départ")
	}
	flags.Select("fr")
	if flags.Selected() != "fr" || !flags.pics[0].gray || flags.pics[1].gray {
		t.Fatal("select")
	}
	flags.Select("zz")
	if flags.Selected() != "fr" {
		t.Fatal("inconnu")
	}
	s := DefaultStyle()
	s.Width = 40
	row := NewGrid(NewRow(NewTitle("Liltress")), NewRow(flags)).layout(s, 0, false)[1]
	body := plain(row)
	lead := len(body) - len(strings.TrimLeft(body, " "))
	if lead < 8 {
		t.Fatalf("drapeaux à gauche %d", lead)
	}
}

func TestKeysTakeKeepsTheTail(t *testing.T) {
	s := DefaultStyle()
	s.Width = 60
	s.Margin = 1
	left := Keys("Contrôles", [][2]string{{"CTRL+L", "Langue"}}, nil).Take()
	right := Keys("", [][2]string{{"↑", "Haut"}}, nil)
	rows := NewGrid(NewRow(left, right)).layout(s, 0, false)
	if lineWidth(rows[0]) != 60 {
		t.Fatalf("largeur %d", lineWidth(rows[0]))
	}
	rightView := plain(right.View(s, nil, false))
	rightLines := strings.Split(rightView, "\n")
	rowLines := strings.Split(plain(rows[0]), "\n")
	if len(rowLines) != len(rightLines) {
		t.Fatalf("lignes %d %d", len(rowLines), len(rightLines))
	}
	for i, line := range rowLines {
		if !strings.HasSuffix(line, rightLines[i]) {
			t.Fatalf("flèches %d\n%s\n%s", i, line, rightLines[i])
		}
	}
	leftNatural := lineWidth(left.View(s, nil, false))
	if 60-1-lineWidth(rightView) <= leftNatural {
		t.Fatal("surplus")
	}
	body := rowLines[1]
	phrase := strings.Index(body, "Langue")
	edge := -1
	if phrase >= 0 {
		edge = strings.Index(body[phrase:], "│")
	}
	if phrase < 0 || edge < len("Langue")+2 {
		t.Fatalf("vide intérieur %d %d\n%s", phrase, edge, body)
	}

	plainLeft := Keys("Contrôles", [][2]string{{"CTRL+L", "Langue"}}, nil)
	plainRight := Keys("", [][2]string{{"↑", "Haut"}}, nil)
	grown := NewGrid(NewRow(NewTitle("Liltress")), NewRow(plainLeft, plainRight)).layout(s, 0, false)
	tail := plain(grown[1])
	head := plain(plainLeft.View(s, nil, false))
	if !strings.HasPrefix(tail, strings.Split(head, "\n")[0]) {
		t.Fatal(tail)
	}
	if strings.HasSuffix(tail, strings.Split(plain(plainRight.View(s, nil, false)), "\n")[0]) {
		t.Fatal("le dernier cadre a gardé sa largeur")
	}
	var bare *keyLegend
	if bare.Take() != nil {
		t.Fatal("nil")
	}
}

func TestKeysColumns(t *testing.T) {
	legend := Keys("Legende", [][2]string{{"↑", "Champ précédent"}}, [][2]string{{"ctrl+l", "Langue"}})
	if legend.Active() {
		t.Fatal("focus")
	}
	view := legend.View(DefaultStyle(), nil, false)
	plainView := plain(view)
	if !strings.Contains(plainView, "Legende") || !strings.Contains(plainView, "Champ précédent") || !strings.Contains(plainView, "ctrl+l : Langue") {
		t.Fatal(plainView)
	}
	if !strings.Contains(view, "229;182;132") || !strings.Contains(view, "204;204;204") {
		t.Fatal("encres")
	}
}

func TestFrameAndKeysReplaceContent(t *testing.T) {
	var blank *alignedLine
	blank.SetText("x")
	blank.SetTitle("y")
	var bare *keyLegend
	bare.SetPairs([][2]string{{"a", "b"}}, nil)
	bare.SetTitle("z")

	frame := Frame("Liltress", AlignCenter, "Bandeau")
	frame.SetTitle("")
	s := DefaultStyle()
	s.Width = 30
	cleared := plain(frame.View(s, nil, false))
	if strings.Contains(cleared, "Bandeau") || !strings.Contains(cleared, "Liltress") {
		t.Fatal(cleared)
	}
	frame.SetText("Autre")
	replaced := plain(frame.View(s, nil, false))
	if !strings.Contains(replaced, "Autre") || strings.Contains(replaced, "Liltress") {
		t.Fatal(replaced)
	}
	center := strings.Index(lineWith(replaced, "Autre"), "Autre")
	left := strings.Index(lineWith(plain(Frame("Autre", AlignLeft, "").View(s, nil, false)), "Autre"), "Autre")
	if center <= left {
		t.Fatal("alignement")
	}

	legend := Keys("Legende", [][2]string{{"↑", "avant"}}, [][2]string{{"ctrl+l", "langue"}})
	pairs := [][2]string{{"↓", "après"}}
	legend.SetPairs(pairs, nil)
	pairs[0][1] = "changé"
	view := plain(legend.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "Legende") || !strings.Contains(view, "après") || strings.Contains(view, "changé") || strings.Contains(view, "avant") {
		t.Fatal(view)
	}
	legend.SetTitle("Touches")
	renamed := plain(legend.View(DefaultStyle(), nil, false))
	if !strings.Contains(renamed, "Touches") || !strings.Contains(renamed, "après") || strings.Contains(renamed, "Legende") {
		t.Fatal(renamed)
	}

	field := NewTextInput("a", InputText, 0, false)
	app := New(DefaultStyle(), NewGrid(NewRow(frame), NewRow(field), NewRow(legend)))
	app = press(t, app, textKey("z"))
	frame.SetText("bandeau")
	legend.SetPairs([][2]string{{"ctrl+q", "quitter"}}, [][2]string{})
	if app.grid.Focused() != field || field.Value() != "z" {
		t.Fatalf("focus %v %q", app.grid.Focused(), field.Value())
	}
	hidden := New(DefaultStyle(), NewGrid(NewRow(legend))).Bind(KeyMap{HideLegend: true})
	if strings.Contains(plain(hidden.View().Content), "colle") {
		t.Fatal("légende stock")
	}
}

func TestButtonFitKeepsActionAndFocus(t *testing.T) {
	var blank *button
	blank.Fit()
	blank.SetLabel("x")

	wide := DefaultStyle()
	wide.Width = 40
	full := Button("OK", nil)
	if lineWidth(full.View(wide, nil, false)) < 30 {
		t.Fatal("colonne")
	}
	hit := 0
	btn := Button("OK", Call{Run: func() { hit++ }})
	btn.Fit()
	hug := btn.View(wide, nil, false)
	if lineWidth(hug) >= 30 || lineWidth(hug) != 6 || !strings.Contains(plain(hug), "OK") {
		t.Fatalf("serré %d", lineWidth(hug))
	}
	btn.SetLabel("Go")
	if !strings.Contains(plain(btn.View(wide, nil, false)), "Go") || strings.Contains(plain(btn.View(wide, nil, false)), "OK") {
		t.Fatal("label")
	}
	btn.Handle(false, "enter")
	if hit != 1 {
		t.Fatal(hit)
	}

	s := DefaultStyle()
	s.Width = 40
	rows := NewGrid(NewRow(NewTitle("Liltress")), NewRow(btn)).layout(s, 0, false)
	if lineWidth(rows[1]) >= lineWidth(rows[0]) {
		t.Fatalf("étiré %d %d", lineWidth(rows[1]), lineWidth(rows[0]))
	}
	pushed := NewGrid(NewRow(NewTitle("Liltress")), RowEnd(btn)).layout(s, 0, false)
	lead := len(plain(pushed[1])) - len(strings.TrimLeft(plain(pushed[1]), " "))
	if lineWidth(pushed[1]) != lineWidth(pushed[0]) || lead < 8 {
		t.Fatalf("gauche %d\n%s", lead, plain(pushed[1]))
	}

	field := NewTextInput("a", InputText, 0, false)
	app := New(DefaultStyle(), NewGrid(NewRow(field), NewRow(btn)))
	app = press(t, app, textKey("z"))
	btn.SetLabel("OK")
	btn.Fit()
	if app.grid.Focused() != field || field.Value() != "z" {
		t.Fatalf("focus %v %q", app.grid.Focused(), field.Value())
	}
}

func TestButtonsChoice(t *testing.T) {
	var bare *buttonRow
	bare.Select("FR")
	if bare.Selected() != "" || bare.Next() != "" {
		t.Fatal("nil")
	}
	row := Buttons("FR", "DE", "ES")
	if row.Selected() != "FR" || row.Active() || row.buttons[0].Active() || row.buttons[1].Active() {
		t.Fatal("anneau")
	}
	if w := lineWidth(row.View(DefaultStyle(), nil, false)); w != 6+1+6+1+6 {
		t.Fatalf("jointure %d", w)
	}
	view := row.View(DefaultStyle(), nil, false)
	if !strings.Contains(view, "229;182;132") || !strings.Contains(view, "107;107;107") || strings.Contains(view, "236;203;168") {
		t.Fatal("encres")
	}
	row.Select("nope")
	if row.Selected() != "FR" {
		t.Fatal("absent")
	}
	if row.Next() != "DE" || row.Next() != "ES" || row.Next() != "FR" {
		t.Fatal(row.Selected())
	}
	labels := []string{"EN", "DE", "ES"}
	row.SetLabels(labels)
	labels[0] = "changé"
	shown := plain(row.View(DefaultStyle(), nil, false))
	if row.Selected() != "EN" || strings.Contains(shown, "changé") || !strings.Contains(shown, "EN") {
		t.Fatal(shown)
	}
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		row.SetLabels([]string{"EN"})
	}()
	if !panicked {
		t.Fatal("longueur")
	}

	s := DefaultStyle()
	s.Width = 40
	laid := NewGrid(NewRow(NewTitle("Liltress")), NewRow(row)).layout(s, 0, false)
	body := plain(laid[1])
	lead := len(body) - len(strings.TrimLeft(body, " "))
	if lineWidth(laid[1]) != lineWidth(laid[0]) || lead < 8 {
		t.Fatalf("gauche %d\n%s", lead, body)
	}
	field := NewTextInput("a", InputText, 0, false)
	app := New(s, NewGrid(NewRow(field), NewRow(row)))
	app = press(t, app, textKey("z"))
	row.Next()
	if app.grid.Focused() != field || field.Value() != "z" {
		t.Fatalf("focus %v %q", app.grid.Focused(), field.Value())
	}
}

func TestCatalogAndKeysUse(t *testing.T) {
	dir := t.TempDir()
	body := "# cartouche\n\n[TITLE]\nTouches\n[UP]\nHaut\n"
	for _, lang := range []string{"en", "fr", "de", "es"} {
		writeText(t, dir, lang, "screen.txt", body)
	}
	got, err := Catalog(dir, "screen.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got["fr"]["TITLE"] != "Touches" || got["fr"]["UP"] != "Haut" {
		t.Fatal(got["fr"])
	}
	if _, ok := got["fr"]["ABSENT"]; ok {
		t.Fatal("clé absente")
	}
	if _, ok := got["fr"]["LABEL_TITLE"]; ok {
		t.Fatal("préfixe")
	}
	writeText(t, dir, "en", "screen.txt", "hors\n")
	if _, err := Catalog(dir, "screen.txt"); err == nil {
		t.Fatal("ligne hors clé")
	}
	writeText(t, dir, "en", "screen.txt", "[]\nx\n")
	if _, err := Catalog(dir, "screen.txt"); err == nil {
		t.Fatal("clé vide")
	}
	if _, err := Catalog(dir, "absent.txt"); err == nil {
		t.Fatal("illisible")
	}
	for _, lang := range []string{"en", "fr", "de", "es"} {
		writeLabel(t, dir, lang, "# note\n[LABEL_LOGIN]\nNom\n")
	}
	if _, err := Labels(dir); err == nil {
		t.Fatal("labels")
	}

	var bare *keyLegend
	bare.Use(got, "fr", "TITLE", nil, nil, nil, nil)
	texts := map[string]map[string]string{
		"fr": {"TITLE": "Touches", "UP": "Haut", "QUIT": "Quitter"},
	}
	legend := Keys("ancien", [][2]string{{"x", "y"}}, [][2]string{{"z", "w"}})
	legend.Use(texts, "fr", "TITLE", []string{"↑"}, []string{"UP"}, []string{"ctrl+q"}, []string{"QUIT"})
	view := plain(legend.View(DefaultStyle(), nil, false))
	if !strings.Contains(view, "Touches") || !strings.Contains(view, "Haut") || !strings.Contains(view, "Quitter") || strings.Contains(view, "ancien") {
		t.Fatal(view)
	}
	texts["fr"]["UP"] = "bas"
	if strings.Contains(plain(legend.View(DefaultStyle(), nil, false)), "bas") {
		t.Fatal("copie")
	}
	expectPanic(t, "langue", func() {
		legend.Use(texts, "it", "TITLE", nil, nil, nil, nil)
	})
	expectPanic(t, "titre", func() {
		legend.Use(texts, "fr", "ABSENT", nil, nil, nil, nil)
	})
	expectPanic(t, "identifiant", func() {
		legend.Use(texts, "fr", "TITLE", []string{"↑"}, []string{"NON"}, nil, nil)
	})
	expectPanic(t, "paires", func() {
		legend.Use(texts, "fr", "TITLE", []string{"↑", "↓"}, []string{"UP"}, nil, nil)
	})
}

func TestQuitAndMeasure(t *testing.T) {
	if _, ok := Quit()().(tea.QuitMsg); !ok {
		t.Fatal("quit")
	}
	base := DefaultStyle()
	base.Height = 12
	next := base.WithMeasure(40, 3, 0)
	if next.Width != 40 || next.PadX != 3 || next.Margin != 0 {
		t.Fatal("mesure")
	}
	if next.Height != 12 || next.Radius != 1 || next.Mark != base.Mark || next.Idle != base.Idle || base.Width == 40 {
		t.Fatal("reste")
	}
}

func expectPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		fn()
	}()
	if !panicked {
		t.Fatal(name)
	}
}

func lineWith(view, word string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, word) {
			return line
		}
	}
	return ""
}

func writeLabel(t *testing.T, root, lang, body string) {
	t.Helper()
	writeText(t, root, lang, "label.txt", body)
}

func writeText(t *testing.T, root, lang, name, body string) {
	t.Helper()
	dir := filepath.Join(root, lang)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
