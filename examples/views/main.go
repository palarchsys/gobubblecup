// Tab bar read from memory. A missing id stays hidden.
//
// From the module root: go run ./examples/views
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
	memo := framework.Memory()
	memo.Set("cards", "Cards")
	memo.Set("input", "Input")

	// The "more" grid already exists. It appears when memory receives its id.
	bar := framework.TabList(framework.ReactiveObject(
		framework.TabItem,
		memo,
		map[string]string{"id": "id", "label": "value"},
	))
	bar.SetTabs(
		framework.Tab("cards", framework.Grid(
			framework.Row(colors),
		)),
		framework.Tab("input", framework.Grid(
			framework.Row(framework.TextInput("Name", framework.InputText, 24, true)),
		)),
		framework.Tab("more", framework.Grid(
			framework.Row(framework.Title("This tab comes from memory.")),
		)),
	)
	add := framework.Button("Tab", framework.Call{Run: func() {
		memo.Set("more", "More")
	}})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(bar),
		framework.Row(add),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("tab:", bar.SelectedID())
}
