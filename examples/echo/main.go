// Exec writes the script output into the area, without leaving the screen.
//
// From the module root: go run ./examples/echo
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	message := framework.TextInput("Message", framework.InputText, 40, true)
	output := framework.TextArea("Output", "Waiting", false, 6, true)
	// The action reads the field on click, then lets Exec draw the output.
	echo := framework.Button("Echo", echoNow{read: message.Value, into: output})
	reload := framework.Button("Reload", framework.Reload{Target: output})

	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(message),
		framework.Row(output),
		framework.Row(echo, reload),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("output:", output.Value())
}

// echoNow rereads the field when the button is pressed.
type echoNow struct {
	read func() string
	into framework.Sink
}

func (e echoNow) Act() tea.Cmd {
	return framework.Exec{
		Script: "bash",
		Args:   []string{"examples/scripts/echo.sh", e.read()},
		Into:   e.into,
	}.Act()
}
