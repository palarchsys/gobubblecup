// Deux onglets fixes. Tab passe à l'onglet suivant et revient au premier.
//
// Lancer depuis la racine du module : go run ./examples/onglets
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
	vues := framework.TabList(
		framework.TabItem("fiches", "Fiches"),
		framework.TabItem("saisie", "Saisie"),
	)
	vues.SetTabs(
		framework.Tab("fiches", framework.Grid(framework.Row(framework.Category("Catalogue", teintes)))),
		framework.Tab("saisie", framework.Grid(framework.Row(nom))),
	)

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(vues),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("onglet:", vues.SelectedID())
	fmt.Println("teinte:", teintes.Value())
	fmt.Println("nom:", nom.Value())
}
