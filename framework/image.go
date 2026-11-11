package framework

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Une cellule de terminal est environ deux fois plus haute que large.
// La hauteur en lignes vaut donc la moitié du ratio de l'image.
const cellRatio = 0.5

// asciiPixels est la rampe de image2ascii : du plus clair au plus dense.
const asciiPixels = " .,:;i1tfLCG08@"

// picture est une image rendue en ASCII coloré.
// La largeur est en colonnes. La hauteur suit le ratio de la source.
type picture struct {
	gate
	path   string
	cols   int
	img    image.Image
	border bool
	gray   bool
}

// included garde les images compilées dans le binaire, indexées par leur nom.
var included = map[string]image.Image{}

// IncludeImage décode des octets déjà présents dans le binaire.
// Le nom est celui passé ensuite à Image. Le disque n'est pas lu.
func IncludeImage(name string, data []byte) {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("framework: image sans chemin")
	}
	if len(data) == 0 {
		panic("framework: image illisible")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic("framework: image illisible")
	}
	included[name] = img
}

func Image(path string, width int) *picture {
	path = strings.TrimSpace(path)
	if path == "" {
		panic("framework: image sans chemin")
	}
	if width < 1 {
		panic("framework: image sans largeur")
	}
	img := included[path]
	if img == nil {
		panic("framework: image absente du binaire")
	}
	return &picture{gate: gate{on: true}, path: path, cols: width, img: img, border: true}
}

// Glyph dessine l'image incluse, sans cadre. Il ne prend pas le focus.
func Glyph(name string, cols int) *picture {
	p := Image(name, cols)
	p.SetBorder(false)
	p.SetActive(false)
	return p
}

// SetBorder dessine le cadre. Sans cadre, la largeur reste celle demandée.
func (p *picture) SetBorder(on bool) {
	if p == nil || p.border == on {
		return
	}
	p.border = on
	notify()
}

// SetGray passe l'image en luminance. L'objet reste actif.
// Disable continue de griser le cadre et de retirer le focus.
func (p *picture) SetGray(on bool) {
	if p == nil || p.gray == on {
		return
	}
	p.gray = on
	notify()
}

func (p *picture) Handle(bool, string) (bool, tea.Cmd) {
	return false, nil
}

func (p *picture) View(s Style, focused Widget, inside bool) string {
	cols := p.cols
	var wash color.Color
	if p.Disabled() {
		s = s.dimmed()
		wash = s.Idle
	}
	if !p.border {
		if s.fit && s.Width > 0 && s.Width < cols {
			cols = s.Width
		}
		if cols < 1 {
			cols = 1
		}
		return p.ascii(cols, wash)
	}
	chrome := 2 + 2*s.padX()
	width := cols + chrome
	if s.fit && s.Width > chrome {
		width = s.Width
		if room := s.content(width); room < cols {
			cols = room
		}
	}
	if cols < 1 {
		cols = 1
	}
	body := padASCII(s, p.ascii(cols, wash), s.content(width))
	return s.box(body, focusedEdge(p, focused, inside), width, 0, "")
}

func (p *picture) ascii(cols int, wash color.Color) string {
	bounds := p.img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	if iw < 1 || ih < 1 {
		return ""
	}
	rows := int(math.Round(float64(cols) * float64(ih) / float64(iw) * cellRatio))
	if rows < 1 {
		rows = 1
	}
	// Un pixel de l'image réduite = un caractère, comme image2ascii.
	small := resizeLanczos(p.img, cols, rows)
	lines := make([]string, rows)
	precision := float64(255*3) / float64(len(asciiPixels)-1)
	for y := 0; y < rows; y++ {
		var b strings.Builder
		for x := 0; x < cols; x++ {
			rr, gg, bb, aa := small.At(x, y).RGBA()
			r := uint8(rr >> 8)
			g := uint8(gg >> 8)
			bl := uint8(bb >> 8)
			a := uint8(aa >> 8)
			value := float64((int(r)+int(g)+int(bl))*int(a)) / 255
			idx := int(math.Floor(value/precision + 0.5))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(asciiPixels) {
				idx = len(asciiPixels) - 1
			}
			ch := asciiPixels[idx : idx+1]
			if ch == " " {
				b.WriteString(" ")
				continue
			}
			var ink color.Color = lipgloss.Color(hex(r, g, bl))
			if wash != nil {
				ink = wash
			} else if p.gray {
				y := byte((int(r)*299 + int(g)*587 + int(bl)*114) / 1000)
				ink = lipgloss.Color(hex(y, y, y))
			}
			b.WriteString(lipgloss.NewStyle().Foreground(ink).Render(ch))
		}
		lines[y] = b.String()
	}
	return strings.Join(lines, "\n")
}

func resizeLanczos(src image.Image, width, height int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	sb := src.Bounds()
	sw, sh := float64(sb.Dx()), float64(sb.Dy())
	for y := 0; y < height; y++ {
		sy := (float64(y)+0.5)*sh/float64(height) - 0.5
		for x := 0; x < width; x++ {
			sx := (float64(x)+0.5)*sw/float64(width) - 0.5
			var r, g, b, a, wsum float64
			x0 := int(math.Floor(sx)) - 2
			x1 := int(math.Floor(sx)) + 3
			y0 := int(math.Floor(sy)) - 2
			y1 := int(math.Floor(sy)) + 3
			for yy := y0; yy <= y1; yy++ {
				wy := lanczos3(sy - float64(yy))
				if wy == 0 {
					continue
				}
				py := yy
				if py < 0 {
					py = 0
				}
				if py >= sb.Dy() {
					py = sb.Dy() - 1
				}
				for xx := x0; xx <= x1; xx++ {
					wx := lanczos3(sx - float64(xx))
					w := wx * wy
					if w == 0 {
						continue
					}
					px := xx
					if px < 0 {
						px = 0
					}
					if px >= sb.Dx() {
						px = sb.Dx() - 1
					}
					rr, gg, bb, aa := src.At(sb.Min.X+px, sb.Min.Y+py).RGBA()
					r += float64(rr>>8) * w
					g += float64(gg>>8) * w
					b += float64(bb>>8) * w
					a += float64(aa>>8) * w
					wsum += w
				}
			}
			if wsum == 0 {
				continue
			}
			dst.SetNRGBA(x, y, clampNRGBA(r/wsum, g/wsum, b/wsum, a/wsum))
		}
	}
	return dst
}

func lanczos3(x float64) float64 {
	if x < 0 {
		x = -x
	}
	if x >= 3 {
		return 0
	}
	if x == 0 {
		return 1
	}
	return 3 * math.Sin(math.Pi*x) * math.Sin(math.Pi*x/3) / (math.Pi * math.Pi * x * x)
}

func clampNRGBA(r, g, b, a float64) color.NRGBA {
	return color.NRGBA{R: byte(clamp255(r)), G: byte(clamp255(g)), B: byte(clamp255(b)), A: byte(clamp255(a))}
}

func clamp255(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v + 0.5
}

func padASCII(s Style, body string, width int) string {
	if width < 1 {
		return body
	}
	rows := strings.Split(body, "\n")
	for i, row := range rows {
		pad := width - lipgloss.Width(row)
		if pad <= 0 {
			continue
		}
		left := pad / 2
		rows[i] = s.shade(left) + row + s.shade(pad-left)
	}
	return strings.Join(rows, "\n")
}

func hex(r, g, b byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{'#', digits[r>>4], digits[r&15], digits[g>>4], digits[g&15], digits[b>>4], digits[b&15]})
}
