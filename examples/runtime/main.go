// Runtime lancé une fois. Trois scripts s'enchaînent. La liste relit leurs slots.
//
// Lancer depuis la racine du module : go run ./examples/runtime
// Les chemins des scripts partent de ce dossier.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	nom := framework.TextInput("Nom", framework.InputText, 24, true)
	controles := framework.Runtime(framework.RuntimeConfig{
		Functions: func(args framework.Store) {
			framework.Bash("un", "Contrôle 1", "examples/scripts/controle1.sh", []string{nom.Value()})
			framework.Bash("deux", "Contrôle 2", "examples/scripts/controle2.sh", nil)
			framework.Bash("trois", "Contrôle 3", "examples/scripts/controle3.sh", nil)
		},
	})
	// Init pose id, label et path. Aucun script ne part.
	controles.Init()

	// La fonction écrit dans une mémoire neuve. Elle ne réécrit pas le runtime.
	suivi := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{controles.Memory},
		Functions: func(args []framework.Store) any {
			forme := framework.Memory()
			for _, id := range []string{"un", "deux", "trois"} {
				forme.Set(id, ligneRuntime(args[0], id))
			}
			return framework.ReactiveObject(
				framework.ListItem,
				forme,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})
	liste := framework.ListUniq("Scripts", suivi)

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(nom),
		framework.Row(liste),
		framework.Row(
			framework.Button("Lancer", controles.Start()),
			framework.Button("Stop", framework.Call{Run: controles.Stop}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("cycle:", controles.Status())
	fmt.Println(controles.Log())
}

func ligneRuntime(memo framework.Store, id string) string {
	label := memo.Get(id).Get("label").String()
	status := memo.Get(id).Get("status").String()
	if status == "" {
		status = "en attente"
	}
	return label + " · " + status
}
