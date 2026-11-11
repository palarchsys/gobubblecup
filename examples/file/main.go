// Reading and writing notes.txt, shown by a watcher.
//
// From the module root: go run ./examples/file
// The file is created in the current directory. Git ignores it.
package main

import (
	"fmt"
	"strings"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	memo := framework.Memory()
	note := framework.TextArea("Note", "a note", true, 6, true)

	// The list does not read the area. It reads the "file" entry in memory.
	preview := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{memo},
		Functions: func(args []framework.Store) any {
			rows := framework.Memory()
			text := args[0].Get("file").Get("data").Get("value").String()
			rows.Set("file", oneLine(text))
			return framework.ReactiveObject(
				framework.ListItem,
				rows,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})

	// ReadFile and WriteFile use the "file" entry. The area is copied around them
	// because it is the text being edited, not that entry.
	load := framework.Button("Load", framework.Call{Run: func() {
		framework.ReadFile{Path: "notes.txt", Memory: memo}.Act()
		note.SetText(memo.Get("file").Get("data").Get("value").String())
	}})
	save := framework.Button("Save", framework.Call{Run: func() {
		memo.Set("file", note.Value())
		framework.WriteFile{Path: "notes.txt", Memory: memo}.Act()
	}})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(note),
		framework.Row(framework.ListUniq("File", preview)),
		framework.Row(load, save),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("file:", memo.Get("file").Get("data").Get("value").String())
}

func oneLine(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if text == "" {
		return "empty"
	}
	return text
}
