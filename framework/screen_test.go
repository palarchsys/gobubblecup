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
	dir := filepath.Join(root, lang)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "label.txt"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
