// Text field, number, secret, editable area, and read-only area.
//
// From the module root: go run ./examples/input
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	// The last argument is the border. Without it, the fields touch.
	// A limit of 0 does not bound the secret. The number keeps only digits.
	name := framework.TextInput("Name", framework.InputText, 24, true)
	port := framework.TextInput("Port", framework.InputNumber, 5, true)
	code := framework.TextInput("Code", framework.InputSecret, 0, false)
	notes := framework.TextArea("Notes", "Waiting", true, 6, true)
	read := framework.TextArea("Read", "This text cannot be edited.\nUp and down scroll it.", false, 4, true)

	example := framework.Button("Example", framework.Call{Run: func() {
		name.SetText("Ada")
		port.SetText("8080")
		code.SetText("secret")
	}})
	reload := framework.Button("Reload", framework.Reload{Target: notes})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Input", name, port, code)),
		framework.Row(notes, read),
		framework.Row(example, reload),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("name:", name.Value())
	fmt.Println("port:", port.Value())
	fmt.Println("code:", code.Value())
	fmt.Println("notes:", notes.Text())
}
