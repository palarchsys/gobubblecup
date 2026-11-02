// Runtime en boucle. Le script repart après une seconde. Pause tient entre deux passages.
//
// Lancer depuis la racine du module : go run ./examples/horloge
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	horloge := framework.Runtime(framework.RuntimeConfig{
		Loop:  true,
		Timer: time.Second,
		Functions: func(args framework.Store) {
			framework.Bash("horloge", "Horloge", "examples/scripts/horloge.sh", nil)
		},
	})
	horloge.Init()

	suivi := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{horloge.Memory},
		Functions: func(args []framework.Store) any {
			forme := framework.Memory()
			forme.Set("horloge", heure(args[0].Get("horloge").Get("return").String()))
			return framework.ReactiveObject(
				framework.ListItem,
				forme,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.ListUniq("Heure", suivi)),
		framework.Row(
			framework.Button("Lancer", horloge.Start()),
			framework.Button("Pause", framework.Call{Run: horloge.Pause}),
			framework.Button("Reprendre", framework.Call{Run: horloge.Resume}),
			framework.Button("Stop", framework.Call{Run: horloge.Stop}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
	}
}

func heure(retour string) string {
	retour = strings.TrimSpace(retour)
	if retour == "" {
		return "en attente"
	}
	if i := strings.IndexByte(retour, '\n'); i >= 0 {
		retour = retour[:i]
	}
	return retour
}
