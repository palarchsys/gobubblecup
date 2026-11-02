// Barre relue à chaque dessin. Seules les valeurs 0 et 1 comptent.
//
// Lancer depuis la racine du module : go run ./examples/progression
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	un := framework.Variable()
	deux := framework.Variable()
	trois := framework.Variable()
	etapes := framework.Variables(un, deux, trois)
	barre := framework.ProgressBar(etapes)

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(barre),
		framework.Row(
			framework.Button("Un", framework.Call{Run: func() { un.Set(1) }}),
			framework.Button("Deux", framework.Call{Run: func() { deux.Set(1) }}),
			framework.Button("Trois", framework.Call{Run: func() { trois.Set(1) }}),
		),
		framework.Row(
			framework.Button("Rien", framework.Call{Run: func() { un.Set(0) }}),
			framework.Button("Hors", framework.Call{Run: func() { trois.Set(2) }}),
			framework.Button("Étape", framework.Call{Run: func() {
				etapes.Add(framework.Variable())
			}}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("un:", un.Value(), "deux:", deux.Value(), "trois:", trois.Value())
}
