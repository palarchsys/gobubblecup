// Mémoire, objet réactif et pastille relue dans la mémoire.
//
// Lancer depuis la racine du module : go run ./examples/memoire
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	// Set(Data) range un enregistrement. Set(id, texte) range id et value.
	// L'objet réactif est créé après ces écritures : Recharger y revient.
	memo := framework.Memory()
	memo.Set(framework.Data{"id": "1", "data": framework.Data{"id": "cyan", "value": "Cyan", "dot": "#00f0ff"}})
	memo.Set(framework.Data{"id": "2", "data": framework.Data{"id": "magenta", "value": "Magenta", "dot": "#ff2d55"}})
	memo.Set("3", "Jaune")

	teintes := framework.ListUniq("Teintes", framework.ReactiveObject(
		framework.ListItem,
		memo,
		map[string]string{"id": "id", "label": "value", "dot": "dot"},
	))

	ajouter := framework.Button("Ajouter", framework.Call{Run: func() {
		memo.Set("4", "Noir")
	}})
	pastille := framework.Button("Pastille", framework.Call{Run: func() {
		teintes.SetDot("cyan", "#ffee00")
	}})
	recharger := framework.Button("Recharger", framework.Reload{Target: teintes})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Mémoire", teintes)),
		framework.Row(ajouter, pastille, recharger),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("choix:", teintes.Value())
	fmt.Println("cyan:", memo.Get("1").Get("data").Get("value").String())
	fmt.Println("pastille:", memo.Get("1").Get("data").Get("dot").String())
}
