// Disable grise un objet et le retire du focus. Il reste visible.
// Sur une catégorie, l'appel descend aux enfants encore actifs.
//
// Lancer depuis la racine du module : go run ./examples/griser
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
	// Déjà gris avant la catégorie : Rétablir ne le rendra pas.
	code := framework.TextInput("Code", framework.InputSecret, 0, true)
	code.Disable()
	saisie := framework.Category("Saisie", teintes, nom, code)

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(saisie),
		framework.Row(
			framework.Button("Griser", framework.Call{Run: saisie.Disable}),
			framework.Button("Rétablir", framework.Call{Run: saisie.Enable}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("catégorie grise:", saisie.Disabled())
	fmt.Println("code gris:", code.Disabled())
	fmt.Println("nom:", nom.Value())
}
