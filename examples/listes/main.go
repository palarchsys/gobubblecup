// Liste unique, liste multiple, pastilles et rechargement.
//
// Lancer depuis la racine du module : go run ./examples/listes
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	teintes := framework.ListUniq("Teintes",
		framework.ListItem("cyan", "Cyan").WithDot("#00f0ff"),
		framework.ListItem("magenta", "Magenta").WithDot("#ff2d55"),
		framework.ListItem("jaune", "Jaune").WithDot("#ffee00"),
	)
	pieces := framework.ListMult("Pièces",
		framework.ListItem("casque", "Casque").WithDot("#c8ccd4"),
		framework.ListItem("veste", "Veste"),
		framework.ListItem("bottes", "Bottes").WithDot("#39ff14"),
	)

	// Recharger rend les lignes d'origine et oublie le choix.
	// Le focus forcé place le curseur sans confirmer la ligne.
	recharger := framework.Button("Recharger", framework.Reload{Target: teintes})
	focus := framework.Button("Magenta", framework.Call{Run: func() {
		teintes.SetFocus("magenta")
	}})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Catalogue", teintes, pieces)),
		framework.Row(recharger, focus),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("teinte:", teintes.Value())
	fmt.Println("pièces:", pieces.Values())
}
