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
	list.Reload()
	view = plain(list.View(DefaultStyle(), list, false))
	if strings.Contains(view, "Noir") || list.Value() != "" || !strings.Contains(view, "Cyan") {
		t.Fatalf("recharge %q\n%s", list.Value(), view)
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
}

func TestChampDuRuntime(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "job.sh", "echo ligne\nsleep 1\necho success\n")
	rt := Runtime(RuntimeConfig{
		Loop: false,
		Dir:  dir,
		Functions: func(args Store) {
			Bash("install.snapd", "Snapd", "job.sh", nil)
		},
	})
	rt.Init()
	if len(rt.jobs) != 1 || len(rt.Memory.Entries()) != 1 {
		t.Fatal("slot initial")
	}

	gen := rt.Memory.Generation()
	rt.Memory.Set(Data{
		"id":   "install.snapd",
		"data": Data{"group": "install.system.title"},
	})
	if rt.Memory.Generation() == gen {
		t.Fatal("génération")
	}
	rt.Memory.Set(Data{
		"id":   "install.snapd",
		"data": Data{"note": "deux"},
	})
	rt.Memory.Set("install.snapd", "note-courte")

	fields := fieldsOf(t, rt.Memory, "install.snapd")
	if fields["group"] != "install.system.title" || fields["note"] != "deux" || fields["status"] != "" {
		t.Fatalf("entrée %v", fields)
	}
	if fields["value"] != "note-courte" || fields["id"] != "install.snapd" {
		t.Fatalf("forme courte %v", fields)
	}
	slot := rt.Memory.Get("install.snapd")
	if slot.Get("status").String() != "" || slot.Get("data").Get("status").String() != "" {
		t.Fatal("status direct")
	}
	if slot.Get("data").Get("group").String() != "install.system.title" || slot.Get("data").Get("note").String() != "deux" {
		t.Fatal("data")
	}

	gen = rt.Memory.Generation()
	mustPanic(t, "framework: champ réservé", func() {
		rt.Memory.Set(Data{
			"id":   "install.snapd",
			"data": Data{"status": "warning", "group": "autre"},
		})
	})
	for _, key := range []string{"id", "label", "path", "status", "return"} {
		mustPanic(t, "framework: champ réservé", func() {
			rt.Memory.Set(key, "x")
		})
		mustPanic(t, "framework: champ réservé", func() {
			rt.Memory.Set(Data{"id": "install.snapd", "data": Data{key: "x"}})
		})
	}
	if rt.Memory.Generation() != gen {
		t.Fatal("génération après refus")
	}
	slot = rt.Memory.Get("install.snapd")
	if slot.Get("status").String() != "" || slot.Get("return").String() != "" {
		t.Fatal("champ réservé écrit")
	}
	if slot.Get("data").Get("group").String() != "install.system.title" {
		t.Fatal("group écrasé")
	}
	if slot.Get("id").String() != "install.snapd" || slot.Get("label").String() != "Snapd" || slot.Get("path").String() != "job.sh" {
		t.Fatal("champs directs")
	}

	entries := len(rt.Memory.Entries())
	jobs := len(rt.jobs)
	mustPanic(t, "framework: slot inconnu", func() {
		rt.Memory.Set(Data{"id": "absent", "data": Data{"group": "x"}})
	})
	mustPanic(t, "framework: slot inconnu", func() {
		rt.Memory.Set("autre", "x")
	})
	if len(rt.Memory.Entries()) != entries || len(rt.jobs) != jobs {
		t.Fatal("entrée fantôme")
	}
	if rt.Memory.Get("absent").Get("id").String() != "" || rt.Memory.Get("autre").Get("data").Get("value").String() != "" {
		t.Fatal("fantôme lisible")
	}
	if rt.Memory.Generation() != gen {
		t.Fatal("génération après id absent")
	}

	ro := rt.Memory.ReadOnly()
	mustPanic(t, "framework: mémoire en lecture seule", func() {
		ro.Set(Data{"id": "install.snapd", "data": Data{"group": "z"}})
	})
	if rt.Memory.Get("install.snapd").Get("data").Get("group").String() != "install.system.title" {
		t.Fatal("vue lecture seule")
	}

	rt.Start().Act()
	waitReturn(t, rt, "install.snapd", "ligne")
	if rt.Memory.Get("install.snapd").Get("data").Get("group").String() != "install.system.title" {
		t.Fatal("group pendant la ligne")
	}
	waitStatus(t, rt, "install.snapd", "success")
	ret := rt.Memory.Get("install.snapd").Get("return").String()
	if !strings.Contains(ret, "ligne") || !strings.Contains(ret, "success") {
		t.Fatal(ret)
	}
	fields = fieldsOf(t, rt.Memory, "install.snapd")
	if fields["group"] != "install.system.title" || fields["status"] != "success" {
		t.Fatalf("après script %v", fields)
	}
	if rt.Memory.Get("install.snapd").Get("status").String() != "success" {
		t.Fatal("status direct après script")
	}
	if rt.Memory.Get("install.snapd").Get("data").Get("status").String() != "" {
		t.Fatal("status sous data")
	}

	items := ReactiveObject(ListItem, rt.Memory, map[string]string{"id": "id", "label": "label"})
	mustPanic(t, "framework: mémoire de runtime", func() {
		items.Reload()
	})
	if rt.Memory.Get("install.snapd").Get("status").String() != "success" || rt.Memory.Get("install.snapd").Get("return").String() != ret {
		t.Fatal("reload a remplacé la racine")
	}
	if rt.Memory.Get("install.snapd").Get("data").Get("group").String() != "install.system.title" {
		t.Fatal("group après reload")
	}

	rt.Stop()
	if rt.Memory.Get("install.snapd").Get("status").String() != "" || rt.Memory.Get("install.snapd").Get("return").String() != "" {
		t.Fatal("stop")
	}
	if rt.Memory.Get("install.snapd").Get("data").Get("group").String() != "install.system.title" {
		t.Fatal("group après stop")
	}
	if rt.Memory.Get("install.snapd").Get("data").Get("note").String() != "deux" {
		t.Fatal("note après stop")
	}
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
	if _, ok := watcher.Result().(*reactiveObject[Item]); !ok {
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

func TestFileUnSlotPending(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "un.sh", "echo success\n")
	writeScript(t, dir, "deux.sh", "echo ligne\nsleep 1\necho success\n")
	writeScript(t, dir, "trois.sh", "echo success\n")
	rt := Runtime(RuntimeConfig{
		Dir: dir,
		Functions: func(Store) {
			Bash("un", "Un", "un.sh", nil)
			Bash("deux", "Deux", "deux.sh", nil)
			Bash("trois", "Trois", "trois.sh", nil)
		},
	})
	rt.Start().Act()
	waitReturn(t, rt, "deux", "ligne")
	if rt.Memory.Get("un").Get("status").String() != "success" {
		t.Fatalf("premier %q", rt.Memory.Get("un").Get("status").String())
	}
	if rt.Memory.Get("deux").Get("status").String() != "pending" {
		t.Fatalf("second %q", rt.Memory.Get("deux").Get("status").String())
	}
	if rt.Memory.Get("trois").Get("status").String() != "" {
		t.Fatalf("troisième %q", rt.Memory.Get("trois").Get("status").String())
	}
	waitStatus(t, rt, "trois", "success")
	if rt.Status() != "terminated" {
		t.Fatalf("cycle %q", rt.Status())
	}
}

func TestErreurArreteLaFile(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "un.sh", "echo success\n")
	writeScript(t, dir, "deux.sh", "echo error\nsleep 1\necho done\n")
	writeScript(t, dir, "trois.sh", "echo success\n")
	rt := Runtime(RuntimeConfig{
		Dir:  dir,
		Loop: true,
		Functions: func(Store) {
			Bash("un", "Un", "un.sh", nil)
			Bash("deux", "Deux", "deux.sh", nil)
			Bash("trois", "Trois", "trois.sh", nil)
		},
	})
	rt.Start().Act()
	waitReturn(t, rt, "deux", "error")
	if rt.Memory.Get("deux").Get("status").String() != "pending" {
		t.Fatal("la regex a classé avant la fin")
	}
	waitStatus(t, rt, "deux", "error")
	time.Sleep(200 * time.Millisecond)
	if rt.Memory.Get("trois").Get("status").String() != "" {
		t.Fatal("le script suivant a démarré")
	}
	if !strings.Contains(rt.Memory.Get("deux").Get("return").String(), "error") {
		t.Fatal("return illisible")
	}
	if rt.Status() != "terminated" {
		t.Fatalf("cycle %q", rt.Status())
	}
	if rt.Memory.Get("un").Get("status").String() != "success" {
		t.Fatal("le premier a été effacé")
	}
}

func TestWarningContinue(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "un.sh", "echo warning\n")
	writeScript(t, dir, "deux.sh", "echo success\n")
	rt := Runtime(RuntimeConfig{
		Dir: dir,
		Functions: func(Store) {
			Bash("un", "Un", "un.sh", nil)
			Bash("deux", "Deux", "deux.sh", nil)
		},
	})
	rt.Start().Act()
	waitStatus(t, rt, "deux", "success")
	if rt.Memory.Get("un").Get("status").String() != "warning" {
		t.Fatal("warning a arrêté la file")
	}
}

func TestBashEnvEtDir(t *testing.T) {
	t.Setenv("KEEP", "oui")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("bonjour\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeScript(t, dir, "env.sh", "echo \"MODE=$MODE\"\necho \"KEEP=$KEEP\"\necho success\n")
	writeScript(t, dir, "lire.sh", "cat note.txt\necho success\n")
	env := Runtime(RuntimeConfig{
		Functions: func(Store) {
			BashEnv("strict", "Strict", filepath.Join(dir, "env.sh"), nil, map[string]string{"MODE": "strict"})
			BashEnv("souple", "Souple", filepath.Join(dir, "env.sh"), nil, map[string]string{"MODE": "souple"})
		},
	})
	env.Start().Act()
	waitStatus(t, env, "souple", "success")
	strict := env.Memory.Get("strict").Get("return").String()
	souple := env.Memory.Get("souple").Get("return").String()
	if !strings.Contains(strict, "MODE=strict") || !strings.Contains(strict, "KEEP=oui") {
		t.Fatalf("env strict %q", strict)
	}
	if !strings.Contains(souple, "MODE=souple") || !strings.Contains(souple, "KEEP=oui") {
		t.Fatalf("env souple %q", souple)
	}

	lecture := Runtime(RuntimeConfig{
		Dir: dir,
		Functions: func(Store) {
			Bash("lire", "Lire", "lire.sh", nil)
		},
	})
	lecture.Start().Act()
	waitStatus(t, lecture, "lire", "success")
	if !strings.Contains(lecture.Memory.Get("lire").Get("return").String(), "bonjour") {
		t.Fatal(lecture.Memory.Get("lire").Get("return").String())
	}
}

func TestCancelEtStop(t *testing.T) {
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
	annule := makeRT()
	annule.Start().Act()
	waitReturn(t, annule, "deux", "deja")
	callWithin(t, 2*time.Second, annule.Cancel)
	if annule.Status() != "terminated" {
		t.Fatalf("cancel %q", annule.Status())
	}
	if annule.Memory.Get("un").Get("status").String() != "success" {
		t.Fatal("le premier a changé")
	}
	if annule.Memory.Get("deux").Get("status").String() != "error" {
		t.Fatalf("second %q", annule.Memory.Get("deux").Get("status").String())
	}
	if !strings.Contains(annule.Memory.Get("deux").Get("return").String(), "deja") {
		t.Fatal("le texte déjà reçu a disparu")
	}
	if annule.Memory.Get("trois").Get("status").String() != "" {
		t.Fatal("le troisième n'est pas vide")
	}
	callWithin(t, time.Second, annule.Cancel)

	stoppe := makeRT()
	stoppe.Start().Act()
	waitReturn(t, stoppe, "deux", "deja")
	callWithin(t, 2*time.Second, stoppe.Stop)
	if stoppe.Status() != "not_initialized" {
		t.Fatalf("stop %q", stoppe.Status())
	}
	for _, id := range []string{"un", "deux", "trois"} {
		if stoppe.Memory.Get(id).Get("status").String() != "" || stoppe.Memory.Get(id).Get("return").String() != "" {
			t.Fatalf("slot %s non vide", id)
		}
	}
}

func mustPanic(t *testing.T, msg string, fn func()) {
	t.Helper()
	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("panic attendu %q", msg)
		}
		text, ok := got.(string)
		if !ok || text != msg {
			t.Fatalf("panic %v", got)
		}
	}()
	fn()
}

func fieldsOf(t *testing.T, memo *memory, id string) map[string]string {
	t.Helper()
	for _, entry := range memo.Entries() {
		if entry.ID == id {
			return entry.Fields
		}
	}
	t.Fatalf("entrée %s absente", id)
	return nil
}

func writeScript(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/bash\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func waitReturn(t *testing.T, rt *runtime, id, part string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(rt.Memory.Get(id).Get("return").String(), part) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("return %s %q", id, rt.Memory.Get(id).Get("return").String())
}

func callWithin(t *testing.T, d time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatal("délai dépassé")
	}
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
