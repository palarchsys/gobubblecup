// Memory, a reactive object, and a dot read back from memory.
//
// From the module root: go run ./examples/memory
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	// Set(Data) stores a record. Set(id, text) stores id and value.
	// The reactive object is created after these writes: Reload returns to them.
	memo := framework.Memory()
	memo.Set(framework.Data{"id": "1", "data": framework.Data{"id": "cyan", "value": "Cyan", "dot": "#00f0ff"}})
	memo.Set(framework.Data{"id": "2", "data": framework.Data{"id": "magenta", "value": "Magenta", "dot": "#ff2d55"}})
	memo.Set("3", "Yellow")

	colors := framework.ListUniq("Colors", framework.ReactiveObject(
		framework.ListItem,
		memo,
		map[string]string{"id": "id", "label": "value", "dot": "dot"},
	))

	add := framework.Button("Add", framework.Call{Run: func() {
		memo.Set("4", "Black")
	}})
	dot := framework.Button("Dot", framework.Call{Run: func() {
		colors.SetDot("cyan", "#ffee00")
	}})
	reload := framework.Button("Reload", framework.Reload{Target: colors})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Category("Memory", colors)),
		framework.Row(add, dot, reload),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("choice:", colors.Value())
	fmt.Println("cyan:", memo.Get("1").Get("data").Get("value").String())
	fmt.Println("dot:", memo.Get("1").Get("data").Get("dot").String())
}
