// Runtime run once. Three scripts run in order. The list rereads their slots.
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

	// The function writes into a fresh memory. It does not write the runtime back.
	watch := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{checks.Memory},
		Functions: func(args []framework.Store) any {
			rows := framework.Memory()
			for _, id := range []string{"one", "two", "three"} {
				rows.Set(id, runtimeLine(args[0], id))
			}
			return framework.ReactiveObject(
				framework.ListItem,
				rows,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})
	list := framework.ListUniq("Scripts", watch)

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(name),
		framework.Row(list),
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

func runtimeLine(memo framework.Store, id string) string {
	label := memo.Get(id).Get("label").String()
	status := memo.Get(id).Get("status").String()
	if status == "" {
		status = "waiting"
	}
	return label + " · " + status
}
