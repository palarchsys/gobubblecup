// Two fixed tabs. Tab moves to the next one and wraps to the first.
//
// From the module root: go run ./examples/tabs
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
	bar := framework.TabList(
		framework.TabItem("cards", "Cards"),
		framework.TabItem("input", "Input"),
	)
	bar.SetTabs(
		framework.Tab("cards", framework.Grid(framework.Row(framework.Category("Catalog", colors)))),
		framework.Tab("input", framework.Grid(framework.Row(name))),
	)

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(bar),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("tab:", bar.SelectedID())
	fmt.Println("color:", colors.Value())
	fmt.Println("name:", name.Value())
}
