// Champ texte, nombre, secret, zone éditable et zone en lecture.
//
// Lancer depuis la racine du module : go run ./examples/saisie
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	// Le dernier argument est le cadre. Sans cadre, les champs se touchent.
	// La limite 0 ne borne pas le secret. Le nombre ne garde que les chiffres.
	nom := framework.TextInput("Nom", framework.InputText, 24, true)
	port := framework.TextInput("Port", framework.InputNumber, 5, true)
	code := framework.TextInput("Code", framework.InputSecret, 0, false)
	notes := framework.TextArea("Notes", "En attente", true, 6, true)
	lecture := framework.TextArea("Lecture", "Ce texte ne s'écrit pas.\nHaut et bas le font défiler.", false, 4, true)

	exemple := framework.Button("Exemple", framework.Call{Run: func() {
		nom.SetText("Ada")
		port.SetText("8080")
		code.SetText("secret")
	}})
	effacer := framework.Button("Recharger", framework.Reload{Target: notes})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Saisie", nom, port, code)),
		framework.Row(notes, lecture),
		framework.Row(exemple, effacer),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("nom:", nom.Value())
	fmt.Println("port:", port.Value())
	fmt.Println("code:", code.Value())
	fmt.Println("notes:", notes.Text())
}
