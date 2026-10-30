package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMemoireEtValeurs(t *testing.T) {
	memo := Memory()
	memo.Set(Data{"id": "1", "data": Data{"id": "cyan", "value": "Cyan"}})
	memo.Set(Data{"id": "2", "data": Data{"id": "magenta", "value": "Magenta"}})
	if got := memo.Get("2").Get("data").Get("value").String(); got != "Magenta" {
		t.Fatalf("lecture %q", got)
	}
	if got := memo.Get("2").Get("data").Get("id").String(); got != "magenta" {
		t.Fatalf("id data %q", got)
	}
	child := Memory()
	child.Set("nom", "inner")
	memo.SetList("groupe", []*memory{child})
	if len(memo.List("groupe")) != 1 || memo.List("groupe")[0].readonly {
		t.Fatal("liste interne")
	}
	ro := memo.ReadOnly()
	if ro.Get("1").Get("data").Get("value").String() != "Cyan" {
		t.Fatal("lecture externe")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("écriture externe")
		}
	}()
	ro.Set("1", "non")
}

func TestListeReactive(t *testing.T) {
	memo := Memory()
	memo.Set(Data{"id": "1", "data": Data{"id": "cyan", "value": "Cyan"}})
	memo.Set(Data{"id": "2", "data": Data{"id": "magenta", "value": "Magenta"}})
	memo.Set(Data{"id": "3", "data": Data{"id": "jaune", "value": "Jaune"}})
	items := ReactiveObject(ListItem, memo, map[string]string{"id": "id", "label": "value"})
	list := ListUniq("Teintes", items)
	if list.Value() != "" {
		t.Fatal("rien n'est choisi")
	}
	view := plain(list.View(DefaultStyle(), list, true))
	if !strings.Contains(view, "Cyan") || !strings.Contains(view, "Jaune") {
		t.Fatal(view)
	}
	list.Handle(true, "down")
	list.Handle(true, "enter")
	if list.Value() != "magenta" {
		t.Fatalf("valeur %q", list.Value())
	}
	memo.Set("4", "Noir")
	view = plain(list.View(DefaultStyle(), list, false))
	if !strings.Contains(view, "Noir") || list.Value() != "magenta" {
		t.Fatalf("valeur %q\n%s", list.Value(), view)
	}

	multi := ListMult("Pièces", ListItem("casque", "Casque"), ListItem("veste", "Veste"))
	multi.Handle(true, "enter")
	multi.Handle(true, "down")
	multi.Handle(true, "enter")
	if strings.Join(multi.Values(), ",") != "casque,veste" {
		t.Fatalf("%v", multi.Values())
	}
}

func TestRuntimeBash(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/bash\necho \"$1\"\necho warning\n"
	if err := os.WriteFile(filepath.Join(dir, "job.sh"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	args := Memory()
	args.Set(Data{"id": "1", "data": Data{"id": "nom", "value": "bonjour"}})
	rt := Runtime(RuntimeConfig{
		Loop: false,
		Args: args,
		Functions: func(args Store) {
			Bash("job", "Contrôle", "job.sh", []string{args.Get("1").Get("data").Get("value").String()})
		},
	})
	if rt.Status() != "not_initialized" {
		t.Fatalf("avant start %q", rt.Status())
	}
	rt.Init()
	if rt.Memory.Get("job").Get("id").String() != "job" || rt.Memory.Get("job").Get("label").String() != "Contrôle" {
		t.Fatal("slot avant start")
	}
	if rt.Memory.Get("job").Get("path").String() != "job.sh" {
		t.Fatal(rt.Memory.Get("job").Get("path").String())
	}
	if rt.Memory.Get("job").Get("status").String() != "" || rt.Memory.Get("job").Get("return").String() != "" {
		t.Fatal("slot pas vide")
	}
	rt.Start().Act()
	waitStatus(t, rt, "job", "warning")
	if rt.Status() != "terminated" {
		t.Fatalf("cycle %q", rt.Status())
	}
	ret := rt.Memory.Get("job").Get("return").String()
	if !strings.Contains(ret, "bonjour") {
		t.Fatal(ret)
	}
	if !strings.Contains(rt.Log(), "job: warning") {
		t.Fatal(rt.Log())
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("mémoire publique")
			}
		}()
		rt.Memory.Set("etat", "non")
	}()
}

func TestReactiveWatcher(t *testing.T) {
	src := Memory()
	src.Set("a", "un")
	autre := Memory()
	autre.Set("b", "seul")
	appels := 0
	watcher := ReactiveWatcher(WatcherConfig{
		Watcher: []Store{src, autre},
		Functions: func(args []Store) any {
			appels++
			forme := Memory()
			forme.Set("ligne", args[0].Get("a").Get("data").Get("value").String())
			forme.Set("cote", args[1].Get("b").Get("data").Get("value").String())
			return ReactiveObject(ListItem, forme, map[string]string{"id": "id", "label": "value"})
		},
	})
	if appels != 1 {
		t.Fatalf("appel initial %d", appels)
	}
	list := ListUniq("L", watcher)
	if view := plain(list.View(DefaultStyle(), list, false)); !strings.Contains(view, "un") || !strings.Contains(view, "seul") {
		t.Fatal(view)
	}
	list.View(DefaultStyle(), list, false)
	if appels != 1 {
		t.Fatalf("recalcul inutile %d", appels)
	}
	src.Set("a", "deux")
	view := plain(list.View(DefaultStyle(), list, false))
	if appels != 2 || !strings.Contains(view, "deux") {
		t.Fatalf("appels %d\n%s", appels, view)
	}
	autre.Set("b", "croise")
	view = plain(list.View(DefaultStyle(), list, false))
	if appels != 3 || !strings.Contains(view, "croise") {
		t.Fatalf("croisement %d\n%s", appels, view)
	}
	if _, ok := watcher.Result().(*reactiveObject); !ok {
		t.Fatal("résultat")
	}
}

func TestTimerDeBoucle(t *testing.T) {
	defaut := Runtime(RuntimeConfig{Loop: true})
	if defaut.timer != 100*time.Millisecond {
		t.Fatalf("défaut %s", defaut.timer)
	}
	choisi := Runtime(RuntimeConfig{Loop: true, Timer: 250 * time.Millisecond})
	if choisi.timer != 250*time.Millisecond {
		t.Fatalf("choisi %s", choisi.timer)
	}
	sans := Runtime(RuntimeConfig{Loop: false})
	if sans.timer != 0 {
		t.Fatalf("sans boucle %s", sans.timer)
	}
}

func TestRuntimeBoucleEtPanique(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "loop.sh")
	if err := os.WriteFile(path, []byte("#!/bin/bash\necho success\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := Runtime(RuntimeConfig{
		Loop: true,
		Functions: func(args Store) {
			Bash("loop", "Boucle", path, nil)
		},
	})
	rt.Start().Act()
	seen := ""
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ret := rt.Memory.Get("loop").Get("return").String()
		state := rt.Status()
		if strings.Contains(ret, "success") && state != "not_initialized" {
			seen = ret
			rt.Stop()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(seen, "success") {
		t.Fatalf("boucle %q", seen)
	}
	if rt.Status() != "not_initialized" {
		t.Fatalf("après stop %q", rt.Status())
	}

	broken := Runtime(RuntimeConfig{})
	broken.guard(func() { panic("boom") })
	if !strings.Contains(broken.Log(), "boom") || broken.Status() != "terminated" {
		t.Fatalf("%q %q", broken.Status(), broken.Log())
	}
}

func TestRuntimesConcurrents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.sh")
	if err := os.WriteFile(path, []byte("#!/bin/bash\necho success\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeRT := func(id string) *runtime {
		return Runtime(RuntimeConfig{
			Functions: func(args Store) {
				Bash(id, id, path, nil)
			},
		})
	}
	one := makeRT("a")
	two := makeRT("b")
	one.Start().Act()
	two.Start().Act()
	waitStatus(t, one, "a", "success")
	waitStatus(t, two, "b", "success")
}

func waitStatus(t *testing.T, rt *runtime, id, status string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if rt.Memory.Get(id).Get("status").String() == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("slot %s status %q", id, rt.Memory.Get(id).Get("status").String())
}
