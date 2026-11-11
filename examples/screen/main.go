// A screen described with Frame, Flags, Form, Status, Keys, and Bind.
// Labels come from locale/<lang>/label.txt. Flags come from flags/<code>.svg.
//
// From the module root: go run ./examples/screen
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	labels, err := framework.Labels("examples/screen/locale")
	if err != nil {
		fmt.Println(err)
		return
	}
	order := []string{"en", "fr", "de", "es"}
	flags, err := framework.Flags("examples/screen/flags", order, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	titles := map[string]string{"en": "Account", "fr": "Compte", "de": "Konto", "es": "Cuenta"}
	lang := 1
	flags.Select(order[lang])
	form := framework.Form(titles[order[lang]],
		framework.Field("LOGIN", framework.Text),
		framework.Field("ADDRESS", framework.Text),
		framework.Field("CODE", framework.Secret),
	)
	form.Use(labels, order[lang])
	status := framework.Status("Status : ", order[lang])
	grid := framework.Grid(
		framework.Row(framework.Frame("gobubblecup", framework.AlignCenter, "")),
		framework.Row(flags),
		framework.Row(form),
		framework.Row(status),
		framework.Row(framework.Keys("Keys",
			[][2]string{{"↑", "Previous field"}, {"↓", "Next field"}, {"←", "Cursor left"}, {"→", "Cursor right"}},
			[][2]string{{"ctrl+l", "Language"}, {"ctrl+c", "Paste"}, {"ctrl+q", "Quit"}},
		)),
	)
	style := framework.DefaultStyle()
	style.Width = 68
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid).Bind(framework.KeyMap{
		Quit:       "ctrl+q",
		Paste:      "ctrl+c",
		HideLegend: true,
		On: map[string]func() tea.Cmd{
			"ctrl+l": func() tea.Cmd {
				lang = (lang + 1) % len(order)
				code := order[lang]
				flags.Select(code)
				form.Use(labels, code)
				form.SetTitle(titles[code])
				status.SetText(code)
				return nil
			},
		},
	})
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("language:", order[lang])
	fmt.Println("flag:", flags.Selected())
	fmt.Println("values:", form.Values())
}
