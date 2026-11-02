// Un watcher suit deux runtimes et reconstruit une seule liste.
//
// Lancer depuis la racine du module : go run ./examples/suivi
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	controles := framework.Runtime(framework.RuntimeConfig{
		Functions: func(args framework.Store) {
			framework.Bash("un", "Contrôle 1", "examples/scripts/controle1.sh", []string{"suivi"})
			framework.Bash("deux", "Contrôle 2", "examples/scripts/controle2.sh", nil)
			framework.Bash("trois", "Contrôle 3", "examples/scripts/controle3.sh", nil)
		},
	})
	horloge := framework.Runtime(framework.RuntimeConfig{
		Loop:  true,
		Timer: time.Second,
		Functions: func(args framework.Store) {
			framework.Bash("horloge", "Horloge", "examples/scripts/horloge.sh", nil)
		},
	})
	controles.Init()
	horloge.Init()

	// args[0] est les contrôles, args[1] l'horloge.
	// Le résultat est un objet réactif sur une mémoire neuve.
	suivi := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{controles.Memory, horloge.Memory},
		Functions: func(args []framework.Store) any {
			forme := framework.Memory()
			for _, id := range []string{"un", "deux", "trois"} {
				forme.Set(id, slot(args[0], id))
			}
			forme.Set("horloge", "Horloge · "+premiere(args[1].Get("horloge").Get("return").String()))
			return framework.ReactiveObject(
				framework.ListItem,
				forme,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.ListUniq("Suivi", suivi)),
		framework.Row(
			framework.Button("Contrôles", controles.Start()),
			framework.Button("Horloge", horloge.Start()),
			framework.Button("Stop", framework.Call{Run: func() {
				controles.Stop()
				horloge.Stop()
			}}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
	}
}

func slot(memo framework.Store, id string) string {
	label := memo.Get(id).Get("label").String()
	status := memo.Get(id).Get("status").String()
	if status == "" {
		status = "en attente"
	}
	return label + " · " + status
}

func premiere(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "en attente"
	}
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return text[:i]
	}
	return text
}
