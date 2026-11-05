// Looping runtime. The script starts again after one second. Pause holds between passes.
//
// From the module root: go run ./examples/clock
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	clock := framework.Runtime(framework.RuntimeConfig{
		Loop:  true,
		Timer: time.Second,
		Functions: func(args framework.Store) {
			framework.Bash("clock", "Clock", "examples/scripts/clock.sh", nil)
		},
	})
	clock.Init()

	watch := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{clock.Memory},
		Functions: func(args []framework.Store) any {
			rows := framework.Memory()
			rows.Set("clock", clockText(args[0].Get("clock").Get("return").String()))
			return framework.ReactiveObject(
				framework.ListItem,
				rows,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.ListUniq("Time", watch)),
		framework.Row(
			framework.Button("Start", clock.Start()),
			framework.Button("Pause", clock.Pause()),
			framework.Button("Resume", clock.Resume()),
			framework.Button("Stop", clock.Stop()),
		),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
	}
}

func clockText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "waiting"
	}
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		text = text[:i]
	}
	return text
}
