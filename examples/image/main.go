// Image compilée dans le binaire. Le lancement ne relit pas le fichier.
//
// Lancer depuis la racine du module : go run ./examples/image
package main

import (
	_ "embed"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

//go:embed logo.png
var logoPNG []byte

func init() {
	framework.IncludeImage("logo.png", logoPNG)
}

func main() {
	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Image("logo.png", 48)),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
	}
}
