// Example program documented in the README.
// From the module root: go build -o hello ./examples/hello
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	memo := framework.Memory()
	memo.Set("1", "Cyan")
	memo.Set("2", "Magenta")

	colors := framework.ListUniq("Colors", framework.ReactiveObject(
		framework.ListItem,
		memo,
		map[string]string{"id": "id", "label": "value"},
	))
	name := framework.TextInput("Name", framework.InputText, 24, true)
	add := framework.Button("Add", framework.Call{Run: func() {
		memo.Set("3", "Yellow")
	}})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Catalog", colors)),
		framework.Row(name, add),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("choice:", colors.Value())
	fmt.Println("name:", name.Value())
}
