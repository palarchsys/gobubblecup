// A screen described with Frame, Form, Status, Keys, and Bind.
// Labels come from locale/<lang>/label.txt. Legend phrases come from locale/<lang>/screen.txt.
//
// From the module root: go run ./examples/screen
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	labels, err := framework.Labels("examples/screen/locale")
	if err != nil {
		fmt.Println(err)
		return
	}
	phrases, err := framework.Catalog("examples/screen/locale", "screen.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	order := []string{"en", "fr", "de", "es"}
	titles := map[string]string{"en": "Account", "fr": "Compte", "de": "Konto", "es": "Cuenta"}
	lang := 1
	form := framework.Form(titles[order[lang]],
		framework.Field("LOGIN", framework.InputText),
		framework.Field("ADDRESS", framework.InputText),
		framework.Field("CODE", framework.InputSecret),
	)
	form.Use(labels, order[lang])
	status := framework.Status("Status : ", order[lang])
	leftKeys := []string{"↑", "↓", "←", "→"}
	leftIDs := []string{"UP", "DOWN", "LEFT", "RIGHT"}
	rightKeys := []string{"ctrl+l", "ctrl+c", "ctrl+q"}
	rightIDs := []string{"LANG", "PASTE", "QUIT"}
	legend := framework.Keys("", nil, nil)
	legend.Use(phrases, order[lang], "TITLE", leftKeys, leftIDs, rightKeys, rightIDs)
	grid := framework.Grid(
		framework.Row(framework.Frame("gobubblecup", framework.AlignCenter, "")),
		framework.Row(form),
		framework.Row(status),
		framework.Row(legend),
	)
	style := framework.DefaultStyle()
	style.Width = 68
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid).Bind(framework.KeyMap{
		Quit:       "ctrl+q",
		Paste:      "ctrl+c",
		HideLegend: true,
		On: map[string]func() framework.Cmd{
			"ctrl+l": func() framework.Cmd {
				lang = (lang + 1) % len(order)
				code := order[lang]
				form.Use(labels, code)
				form.SetTitle(titles[code])
				status.SetText(code)
				legend.Use(phrases, code, "TITLE", leftKeys, leftIDs, rightKeys, rightIDs)
				return nil
			},
		},
	})
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("language:", order[lang])
	fmt.Println("values:", form.Values())
}
