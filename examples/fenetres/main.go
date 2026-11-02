// Deux grilles dans une fenêtre. Suivante et Précédente font le tour.
//
// Lancer depuis la racine du module : go run ./examples/fenetres
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
	nom := framework.TextInput("Nom", framework.InputText, 24, true)
	fenetre := framework.Window()
	une := framework.Grid(
		framework.Row(framework.Title("Fenêtre 1")),
		framework.Row(teintes),
		framework.Row(framework.Button("Suivante", fenetre.Next())),
	)
	deux := framework.Grid(
		framework.Row(framework.Title("Fenêtre 2")),
		framework.Row(nom),
		framework.Row(framework.Button("Précédente", fenetre.Prev())),
	)
	fenetre.Use(une, deux)

	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.NewWindow(style, fenetre)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fenêtre:", fenetre.Index())
	fmt.Println("teinte:", teintes.Value())
	fmt.Println("nom:", nom.Value())
}
