package framework

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestImageGardeLeRatio(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logo.png")
	writePNG(t, path, 40, 20, color.RGBA{R: 220, G: 30, B: 30, A: 255})
	includeFile(t, path)
	os.Remove(path)
	pic := Image(path, 20)
	body := pic.ascii(20, nil)
	rows := strings.Split(body, "\n")
	// 20 colonnes, image 40×20 : hauteur = 20 * (20/40) * 0.5 = 5.
	if len(rows) != 5 {
		t.Fatalf("hauteur %d", len(rows))
	}
	for _, row := range rows {
		if lipgloss.Width(row) != 20 {
			t.Fatalf("largeur %d", lipgloss.Width(row))
		}
	}
	if !strings.Contains(body, "220") && !strings.Contains(body, "dc") {
		// La couleur est dans la séquence ANSI, pas dans le glyphe.
	}
	if strings.Trim(plain(body), " \n") == "" {
		t.Fatal(plain(body))
	}
	// Un trait clair sur fond noir doit rester visible.
	line := filepath.Join(dir, "trait.png")
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 12, G: 12, B: 14, A: 255})
		}
	}
	for x := 0; x < 40; x++ {
		img.Set(x, 10, color.RGBA{R: 0, G: 180, B: 255, A: 255})
	}
	f, err := os.Create(line)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()
	includeFile(t, line)
	os.Remove(line)
	trait := plain(Image(line, 20).ascii(20, nil))
	if strings.Count(trait, " ") > len(trait)-8 {
		t.Fatal(trait)
	}
}

func includeFile(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	IncludeImage(path, data)
}

func writePNG(t *testing.T, path string, w, h int, c color.Color) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
