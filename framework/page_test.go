package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPageChoiceAndSection(t *testing.T) {
	section := Section("Système", Gap())
	s := DefaultStyle()
	s.Width = 40
	view := section.View(s, nil, false)
	lines := strings.Split(plain(view), "\n")
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "╭─ Système ") || !strings.HasSuffix(lines[0], "╮") {
		t.Fatal(lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "╰") || !strings.Contains(view, "229;182;132") {
		t.Fatal(plain(view))
	}
	section.SetLegend("Hôte")
	if !strings.Contains(plain(section.View(s, nil, false)), "╭─ Hôte ") {
		t.Fatal("légende")
	}

	memo := Memory()
	choice := Choice("Langue", ListItem("fr", "Français"), ListItem("en", "English"))
	choice.Bind(memo, "langue")
	if memo.Get("langue").Get("data").Get("value").String() != "fr" || choice.Label() != "Français" {
		t.Fatal(choice.Value())
	}
	app := New(s, Grid(Row(choice)))
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyDown})
	if choice.Value() != "en" || memo.Get("langue").Get("data").Get("value").String() != "en" {
		t.Fatal(choice.Value())
	}
	text := plain(app.View().Content)
	if !strings.Contains(text, "✓ English") || strings.Contains(text, "✓ Français") {
		t.Fatal(text)
	}
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyDown})
	if choice.Value() != "en" {
		t.Fatal("boucle")
	}
	hit := 0
	page := Page("gobubblecup", Facts(Fact{Label: "Langue", Text: choice.Label}),
		Section("Langue", Moves(Call{Run: func() { hit++ }}, nil, choice)...))
	app = New(s, Grid(Row(page)))
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyRight})
	app = press(t, app, key(tea.KeyEnter))
	if hit != 1 || app.grid.Focused() == choice {
		t.Fatal("suite")
	}
}

func TestReportsRulesAndRoot(t *testing.T) {
	body := "hostname=deepcore\nos=ubuntu\nversion=26.10\ninstall=absent\nwarning\n"
	text := plain(stepBlock(DefaultStyle(), "script/monitor-host.sh", hostLines("warning", body), 40))
	for _, part := range []string{"┌ script/monitor-host.sh", "✓ Exécution", "• Hôte : deepcore", "• OS : Ubuntu 26.10", "⚠ Installation absente"} {
		if !strings.Contains(text, part) {
			t.Fatalf("manque %s\n%s", part, text)
		}
	}
	idle := plain(HostReport(nil, "host", "script/monitor-host.sh").View(DefaultStyle(), nil, false))
	if !strings.Contains(idle, "• En attente") {
		t.Fatal(idle)
	}
	run := plain(stepBlock(DefaultStyle(), "script/apply-config.sh", runLines("success", "Contrôle accepté.\nsuccess\n"), 40))
	if !strings.Contains(run, "✓ Contrôle accepté.") || strings.Contains(run, "success") {
		t.Fatal(run)
	}
	memo := Memory()
	memo.setField("host", "status", "warning")
	memo.setField("host", "return", body)
	facts := plain(HostFacts(memo, "host").View(DefaultStyle(), nil, false))
	if !strings.Contains(facts, "deepcore") || !strings.Contains(facts, "Ubuntu 26.10") || !strings.Contains(facts, "absente") {
		t.Fatal(facts)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "configuration.json")
	raw := []byte(`{"variables":[{"name":"CODE","minimum_length":2,"forbidden":[" "],"equals":"AGAIN"},{"name":"AGAIN"}]}`)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	rules, err := LoadRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if name, equal, ok := Fail(rules, []string{"CODE", "AGAIN"}, []string{"ab", "ab"}); ok != true || name != "" || equal {
		t.Fatalf("passe %s %v %v", name, equal, ok)
	}
	if name, equal, ok := Fail(rules, []string{"CODE", "AGAIN"}, []string{"a", "a"}); ok || name != "CODE" || equal {
		t.Fatalf("court %s %v", name, equal)
	}
	if name, equal, ok := Fail(rules, []string{"CODE", "AGAIN"}, []string{"ab", "cd"}); ok || name != "CODE" || !equal {
		t.Fatalf("écart %s %v", name, equal)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "go"), 0o755); err != nil {
		t.Fatal(err)
	}
	scriptDir := filepath.Join(root, "workstation", "script")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "run.sh"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "go"))
	got, err := FindRoot("", "", "workstation", "run.sh")
	if err != nil || got != filepath.Join(root, "workstation") {
		t.Fatal(got, err)
	}
	if err := Handoff("", ""); err == nil {
		t.Fatal("relais")
	}
	if err := RequireRoot(); os.Geteuid() == 0 && err != nil || os.Geteuid() != 0 && err == nil {
		t.Fatal(err)
	}
}

func TestWalkAndScript(t *testing.T) {
	lang := Choice("Langue", ListItem("fr", "Français"), ListItem("en", "English"))
	win := Window()
	win.Walk("gobubblecup", Facts(Fact{Label: "Langue", Text: lang.Label}),
		Stage{Legend: "Langue", Kids: []Widget{lang}},
		Stage{Legend: "Note", Kids: []Widget{Note("suite")}},
	)
	s := DefaultStyle()
	s.Width = 40
	app := NewWindow(s, win)
	app = press(t, app, tea.KeyPressMsg{Code: tea.KeyRight})
	app = press(t, app, key(tea.KeyEnter))
	if win.Index() != 1 {
		t.Fatal(win.Index())
	}
	text := plain(app.View().Content)
	if !strings.Contains(text, "suite") || !strings.Contains(text, "Retour") || !strings.Contains(text, "Français") {
		t.Fatal(text)
	}

	job := Script(t.TempDir(), "host", "Hôte", "script/monitor-host.sh", nil, nil)
	if job.Memory.Get("host").Get("path").String() != "script/monitor-host.sh" {
		t.Fatal(job.Memory.Get("host").Get("path").String())
	}
	wide := DefaultStyle()
	wide.Width = 40
	idle := plain(job.Host().View(wide, nil, false))
	if !strings.Contains(idle, "• En attente") || !strings.Contains(idle, "script/monitor-host.sh") {
		t.Fatal(idle)
	}
}
