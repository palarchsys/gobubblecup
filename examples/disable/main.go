// Disable grays an object and removes it from focus. It stays visible.
// On a category, the call reaches children that are still active.
//
// From the module root: go run ./examples/disable
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	colors := framework.ListUniq("Colors",
		framework.ListItem("cyan", "Cyan"),
		framework.ListItem("magenta", "Magenta"),
	)
	name := framework.TextInput("Name", framework.InputText, 24, true)
	// Already disabled before the category: Enable will not bring it back.
	code := framework.TextInput("Code", framework.InputSecret, 0, true)
	code.Disable()
	input := framework.Category("Input", colors, name, code)

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(input),
		framework.Row(
			framework.Button("Disable", framework.Call{Run: input.Disable}),
			framework.Button("Enable", framework.Call{Run: input.Enable}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("category disabled:", input.Disabled())
	fmt.Println("code disabled:", code.Disabled())
	fmt.Println("name:", name.Value())
}
