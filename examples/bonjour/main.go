// Programme d'exemple documenté dans le README.
// Depuis la racine du module : go build -o bonjour ./examples/bonjour
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	memo := framework.Memory()
	memo.Set("1", "Cyan")
	memo.Set("2", "Magenta")

	teintes := framework.ListUniq("Teintes", framework.ReactiveObject(
		framework.ListItem,
		memo,
		map[string]string{"id": "id", "label": "value"},
	))
	nom := framework.TextInput("Nom", framework.InputText, 24, true)
	ajouter := framework.Button("Ajouter", framework.Call{Run: func() {
		memo.Set("3", "Jaune")
	}})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Catalogue", teintes)),
		framework.Row(nom, ajouter),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("choix:", teintes.Value())
	fmt.Println("nom:", nom.Value())
}
