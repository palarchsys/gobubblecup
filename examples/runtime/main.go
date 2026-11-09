// Runtime run once. Three scripts run in order.
// The list reads each slot. The area shows the return of the focused line.
//
// From the module root: go run ./examples/runtime
// Script paths are relative to this directory.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	name := framework.TextInput("Name", framework.InputText, 24, true)
	checks := framework.Runtime(framework.RuntimeConfig{
		Functions: func(args framework.Store) {
			framework.Bash("one", "Check 1", "examples/scripts/check1.sh", []string{name.Value()})
			framework.Bash("two", "Check 2", "examples/scripts/check2.sh", nil)
			framework.Bash("three", "Check 3", "examples/scripts/check3.sh", nil)
		},
	})
	// Init writes id, label, and path. No script starts.
	checks.Init()

	link := framework.Memory()
	scripts := framework.ListUniq("Scripts", framework.ReactiveObject(
		framework.ListItem,
		checks.Memory,
		map[string]string{"id": "id", "label": "label", "dot": "status"},
	))
	scripts.Bind(link, "current")
	output := framework.TextView(checks.Memory, link, "current")

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(name),
		framework.Row(framework.Category("Run", scripts, output)),
		framework.Row(
			framework.Button("Start", checks.Start()),
			framework.Button("Stop", checks.Stop()),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("cycle:", checks.Status())
	fmt.Println(checks.Log())
}
