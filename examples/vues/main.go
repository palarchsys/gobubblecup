// Barre d'onglets relue dans une mémoire. Un id absent reste masqué.
//
// Lancer depuis la racine du module : go run ./examples/vues
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	teintes := framework.ListUniq("Teintes",
		framework.ListItem("cyan", "Cyan"),
		framework.ListItem("magenta", "Magenta"),
	)
	memo := framework.Memory()
	memo.Set("fiches", "Fiches")
	memo.Set("saisie", "Saisie")

	// La grille « plus » existe déjà. Elle apparaît quand la mémoire reçoit son id.
	vues := framework.TabList(framework.ReactiveObject(
		framework.TabItem,
		memo,
		map[string]string{"id": "id", "label": "value"},
	))
	vues.SetTabs(
		framework.Tab("fiches", framework.Grid(
			framework.Row(teintes),
		)),
		framework.Tab("saisie", framework.Grid(
			framework.Row(framework.TextInput("Nom", framework.InputText, 24, true)),
		)),
		framework.Tab("plus", framework.Grid(
			framework.Row(framework.Title("Cet onglet vient de la mémoire.")),
		)),
	)
	ajouter := framework.Button("Onglet", framework.Call{Run: func() {
		memo.Set("plus", "Plus")
	}})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(vues),
		framework.Row(ajouter),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("onglet:", vues.SelectedID())
}
