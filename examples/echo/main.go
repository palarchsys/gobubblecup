// Exec écrit la sortie du script dans la zone, sans quitter l'écran.
//
// Lancer depuis la racine du module : go run ./examples/echo
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	message := framework.TextInput("Message", framework.InputText, 40, true)
	sortie := framework.TextArea("Sortie", "En attente", false, 6, true)
	// L'action lit le champ au clic, puis laisse Exec rendre la sortie à l'écran.
	echo := framework.Button("Echo", echoMaintenant{lire: message.Value, vers: sortie})
	recharger := framework.Button("Recharger", framework.Reload{Target: sortie})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(message),
		framework.Row(sortie),
		framework.Row(echo, recharger),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("sortie:", sortie.Text())
}

// echoMaintenant relit le champ quand le bouton est pressé.
type echoMaintenant struct {
	lire func() string
	vers framework.Sink
}

func (e echoMaintenant) Act() tea.Cmd {
	return framework.Exec{
		Script: "bash",
		Args:   []string{"examples/scripts/echo.sh", e.lire()},
		Into:   e.vers,
	}.Act()
}
