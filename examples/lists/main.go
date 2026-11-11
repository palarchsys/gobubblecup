// Single list, multiple list, dots, reload, and a button row pushed right.
//
// From the module root: go run ./examples/lists
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	colors := framework.ListUniq("Colors",
		framework.ListItem("cyan", "Cyan").WithDot("#00f0ff"),
		framework.ListItem("magenta", "Magenta").WithDot("#ff2d55"),
		framework.ListItem("yellow", "Yellow").WithDot("#ffee00"),
	)
	parts := framework.ListMult("Parts",
		framework.ListItem("helmet", "Helmet").WithDot("#c8ccd4"),
		framework.ListItem("jacket", "Jacket"),
		framework.ListItem("boots", "Boots").WithDot("#39ff14"),
	)

	// Reload restores the original rows and forgets the choice.
	// SetFocus moves the cursor without confirming the row.
	reload := framework.Button("Reload", framework.Reload{Target: colors})
	focus := framework.Button("Magenta", framework.Call{Run: func() {
		colors.SetFocus("magenta")
	}})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Catalog", colors, parts)),
		framework.RowEnd(reload, focus),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("color:", colors.Value())
	fmt.Println("parts:", parts.Values())
}
