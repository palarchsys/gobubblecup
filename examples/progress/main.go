// Bar reread on each draw. Only the values 0 and 1 count.
//
// From the module root: go run ./examples/progress
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	one := framework.Variable()
	two := framework.Variable()
	three := framework.Variable()
	steps := framework.Variables(one, two, three)
	bar := framework.ProgressBar(steps)

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(bar),
		framework.Row(
			framework.Button("One", framework.Call{Run: func() { one.Set(1) }}),
			framework.Button("Two", framework.Call{Run: func() { two.Set(1) }}),
			framework.Button("Three", framework.Call{Run: func() { three.Set(1) }}),
		),
		framework.Row(
			framework.Button("None", framework.Call{Run: func() { one.Set(0) }}),
			framework.Button("Out", framework.Call{Run: func() { three.Set(2) }}),
			framework.Button("Step", framework.Call{Run: func() {
				steps.Add(framework.Variable())
			}}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("one:", one.Value(), "two:", two.Value(), "three:", three.Value())
}
