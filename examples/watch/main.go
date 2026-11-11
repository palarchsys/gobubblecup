// A watcher follows two runtimes and rebuilds a single list.
// Check rows keep the script label. Their dot is the slot status.
//
// From the module root: go run ./examples/watch
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	checks := framework.Runtime(framework.RuntimeConfig{
		Functions: func(args framework.Store) {
			framework.Bash("one", "Check 1", "examples/scripts/check1.sh", []string{"watch"})
			framework.Bash("two", "Check 2", "examples/scripts/check2.sh", nil)
			framework.Bash("three", "Check 3", "examples/scripts/check3.sh", nil)
		},
	})
	clock := framework.Runtime(framework.RuntimeConfig{
		Loop:  true,
		Timer: time.Second,
		Functions: func(args framework.Store) {
			framework.Bash("clock", "Clock", "examples/scripts/clock.sh", nil)
		},
	})
	checks.Init()
	clock.Init()

	// args[0] is the checks, args[1] the clock.
	// The result is a reactive object on a fresh memory.
	watch := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{checks.Memory, clock.Memory},
		Functions: func(args []framework.Store) any {
			rows := framework.Memory()
			for _, id := range []string{"one", "two", "three"} {
				put(rows, id, args[0].Get(id).Get("label").String(), args[0].Get(id).Get("status").String())
			}
			put(rows, "clock", "Clock · "+firstLine(args[1].Get("clock").Get("return").String()), "")
			return framework.ReactiveObject(
				framework.ListItem,
				rows,
				map[string]string{"id": "id", "label": "value", "dot": "dot"},
			)
		},
	})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.ListUniq("Watch", watch)),
		framework.Row(
			framework.Button("Checks", checks.Start()),
			framework.Button("Clock", clock.Start()),
			framework.Button("Stop", framework.Call{Run: func() {
				checks.Stop().Act()
				clock.Stop().Act()
			}}),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
	}
}

func put(rows framework.Store, id, label, dot string) {
	data := framework.Data{"id": id, "value": label}
	if dot != "" {
		data["dot"] = dot
	}
	rows.Set(framework.Data{"id": id, "data": data})
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "waiting"
	}
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return text[:i]
	}
	return text
}
