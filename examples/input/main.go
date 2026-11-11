// Form, fixed columns, filler, placeholders, number field, editable area, and read-only area.
//
// From the module root: go run ./examples/input
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	// Form aligns the labels and moves up and down between its fields.
	// The number field stays a TextInput beside the form.
	account := framework.Form("Account",
		framework.Field("NAME", framework.InputText),
		framework.Field("CODE", framework.InputSecret),
	)
	account.SetLabels(map[string]string{"NAME": "Name", "CODE": "Code"})
	port := framework.TextInput("Port", framework.InputNumber, 5, false).Placeholder("8080")
	host := framework.TextInput("Host", framework.InputText, 0, false).Span(8, 10).Placeholder("name")
	state := framework.Label("local", 8)
	notes := framework.TextArea("Notes", "Waiting", true, 6, true)
	read := framework.TextArea("Read", "This text cannot be edited.\nUp and down scroll it.", false, 4, true)

	example := framework.Button("Example", framework.Call{Run: func() {
		account.SetTitle("Compte")
		account.SetLabels(map[string]string{"NAME": "Nom", "CODE": "Code"})
		port.SetText("8080")
		port.SetTitle("Port")
	}})
	reload := framework.Button("Reload", framework.Reload{Target: notes})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Input", account, port)),
		framework.Row(host, framework.Fill(), state),
		framework.Row(notes, read),
		framework.Row(example, reload),
	)
	style := framework.DefaultStyle()
	style.Width = 36
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("account:", account.Values())
	fmt.Println("port:", port.Value())
	fmt.Println("host:", host.Value())
	fmt.Println("notes:", notes.Value())
}
