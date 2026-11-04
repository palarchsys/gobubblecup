// Two grids in a window. Next and Previous wrap around.
//
// From the module root: go run ./examples/windows
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	colors := framework.ListUniq("Colors",
		framework.ListItem("cyan", "Cyan"),
		framework.ListItem("magenta", "Magenta"),
	)
	name := framework.TextInput("Name", framework.InputText, 24, true)
	win := framework.Window()
	one := framework.Grid(
		framework.Row(framework.Title("Window 1")),
		framework.Row(colors),
		framework.Row(framework.Button("Next", win.Next())),
	)
	two := framework.Grid(
		framework.Row(framework.Title("Window 2")),
		framework.Row(name),
		framework.Row(framework.Button("Previous", win.Prev())),
	)
	win.Use(one, two)

	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.NewWindow(style, win)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("window:", win.Index())
	fmt.Println("color:", colors.Value())
	fmt.Println("name:", name.Value())
}
