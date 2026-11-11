// PNG, SVG glyph, gray, and a row pushed to the right.
// Launch does not read logo.png or mark.svg again.
//
// From the module root: go run ./examples/image
package main

import (
	_ "embed"
	"fmt"

	"github.com/palarchsys/gobubblecup/framework"
)

//go:embed logo.png
var logoPNG []byte

//go:embed mark.svg
var markSVG []byte

func init() {
	framework.IncludeImage("logo.png", logoPNG)
	if err := framework.IncludeSVG("mark.svg", markSVG); err != nil {
		panic(err)
	}
}

func main() {
	mark := framework.Glyph("mark.svg", 18)
	gray := false
	grid := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(framework.Image("logo.png", 36)),
		framework.RowEnd(mark, framework.Button("Gray", framework.Call{Run: func() {
			gray = !gray
			mark.SetGray(gray)
		}})),
	)
	style := framework.DefaultStyle()
	style.Width = 60
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grid)
	if err := app.Run(); err != nil {
		fmt.Println(err)
	}
}
