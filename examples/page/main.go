// Installer page: a choice, a field, a host probe, and one script.
// Suite and Back are drawn in French. The host phrases are French too.
//
// From the module root: go run ./examples/page
package main

import (
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	lang := framework.Choice("Language",
		framework.ListItem("en", "English"),
		framework.ListItem("fr", "Français"),
	)
	name := framework.TextInput("Name", framework.InputText, 24, false)
	host := framework.Script("", "host", "Host", "examples/scripts/host.sh", nil, nil)
	apply := framework.Script("", "apply", "Apply", "examples/scripts/apply.sh", nil, func() map[string]string {
		return map[string]string{
			"LANGUAGE": lang.Value(),
			"NAME":     name.Value(),
		}
	})
	facts := framework.Facts(
		framework.Fact{Label: "Language", Text: lang.Label},
		framework.Fact{Label: "Name", Text: name.Value},
	)
	win := framework.Window()
	win.Walk("gobubblecup", facts,
		framework.Stage{Legend: "Language", Kids: []framework.Widget{lang}},
		framework.Stage{Legend: "Name", Kids: []framework.Widget{name}},
		framework.Stage{Legend: "Host", Kids: []framework.Widget{
			host.Host(),
			framework.Gap(),
			framework.Bullet("Run", host.Start()),
		}},
		framework.Stage{Legend: "Apply", Kids: []framework.Widget{
			apply.Report(),
			framework.Gap(),
			framework.Bullet("Run", apply.Start()),
		}},
	)
	style := framework.DefaultStyle().WithMeasure(48, 2, 0)
	style.Height = style.Width * 3 / 4
	if err := framework.NewWindow(style, win).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("language:", lang.Value())
	fmt.Println("name:", name.Value())
}
