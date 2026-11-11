package framework

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"strconv"
	"strings"
)

// IncludeSVG rasterise un SVG et le range sous name, comme IncludeImage.
// rect est rempli. Un path n'accepte que M, L, H et V, y compris plusieurs sous-tracés.
// g est ignoré. defs et clipPath sont sautés. Un autre élément ou un autre ordre renvoie une erreur.
func IncludeSVG(name string, data []byte) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("svg: nom vide")
	}
	if len(data) == 0 {
		return fmt.Errorf("svg: taille absente")
	}
	img, err := rasterSVG(data)
	if err != nil {
		return err
	}
	included[name] = img
	return nil
}

func rasterSVG(raw []byte) (image.Image, error) {
	dec := xml.NewDecoder(strings.NewReader(string(raw)))
	var img *image.NRGBA
	skip := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			name := tok.Name.Local
			if skip > 0 {
				if name == "defs" || name == "clipPath" {
					skip++
				}
				continue
			}
			switch name {
			case "svg":
				w, err := svgNeed(attr(tok, "width"), "largeur")
				if err != nil {
					return nil, err
				}
				h, err := svgNeed(attr(tok, "height"), "hauteur")
				if err != nil {
					return nil, err
				}
				if w < 1 || h < 1 {
					return nil, fmt.Errorf("svg: taille vide")
				}
				img = image.NewNRGBA(image.Rect(0, 0, int(math.Round(w)), int(math.Round(h))))
			case "g":
			case "defs", "clipPath":
				skip++
			case "rect":
				if img == nil {
					return nil, fmt.Errorf("svg: taille absente")
				}
				if err := drawRect(img, tok); err != nil {
					return nil, err
				}
			case "path":
				if img == nil {
					return nil, fmt.Errorf("svg: taille absente")
				}
				if err := drawPath(img, tok); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("svg: élément %s", name)
			}
		case xml.EndElement:
			name := tok.Name.Local
			if skip > 0 && (name == "defs" || name == "clipPath") {
				skip--
			}
		}
	}
	if img == nil {
		return nil, fmt.Errorf("svg: taille absente")
	}
	return img, nil
}

func drawRect(img *image.NRGBA, tok xml.StartElement) error {
	fill, err := parseHex(attr(tok, "fill"))
	if err != nil {
		return err
	}
	x, err := optFloat(attr(tok, "x"))
	if err != nil {
		return err
	}
	y, err := optFloat(attr(tok, "y"))
	if err != nil {
		return err
	}
	w, err := svgNeed(attr(tok, "width"), "largeur")
	if err != nil {
		return err
	}
	h, err := svgNeed(attr(tok, "height"), "hauteur")
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		cy := float64(py) + 0.5
		if cy < y || cy >= y+h {
			continue
		}
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			cx := float64(px) + 0.5
			if cx < x || cx >= x+w {
				continue
			}
			img.SetNRGBA(px, py, fill)
		}
	}
	return nil
}

func drawPath(img *image.NRGBA, tok xml.StartElement) error {
	ink, err := parseHex(attr(tok, "stroke"))
	if err != nil {
		return err
	}
	width, err := svgNeed(attr(tok, "stroke-width"), "épaisseur")
	if err != nil {
		return err
	}
	segs, err := parsePath(attr(tok, "d"))
	if err != nil {
		return err
	}
	for _, seg := range segs {
		stroke(img, seg, width, ink)
	}
	return nil
}

type segment struct{ x0, y0, x1, y1 float64 }

func parsePath(d string) ([]segment, error) {
	i := 0
	var cx, cy float64
	started := false
	var out []segment
	for {
		i = skipSep(d, i)
		if i >= len(d) {
			return out, nil
		}
		cmd := d[i]
		if cmd < 'A' || cmd > 'Z' {
			return nil, fmt.Errorf("svg: ordre %q", string(cmd))
		}
		i++
		switch cmd {
		case 'M':
			x, y, next, err := readPair(d, i)
			if err != nil {
				return nil, err
			}
			cx, cy, i = x, y, next
			started = true
			for numberFollows(d, i) {
				x, y, next, err = readPair(d, i)
				if err != nil {
					return nil, err
				}
				out = append(out, segment{cx, cy, x, y})
				cx, cy, i = x, y, next
			}
		case 'L':
			if !started {
				return nil, fmt.Errorf("svg: L sans M")
			}
			for {
				x, y, next, err := readPair(d, i)
				if err != nil {
					return nil, err
				}
				out = append(out, segment{cx, cy, x, y})
				cx, cy, i = x, y, next
				if !numberFollows(d, i) {
					break
				}
			}
		case 'H':
			if !started {
				return nil, fmt.Errorf("svg: H sans M")
			}
			x, next, err := readNum(d, i)
			if err != nil {
				return nil, err
			}
			out = append(out, segment{cx, cy, x, cy})
			cx, i = x, next
		case 'V':
			if !started {
				return nil, fmt.Errorf("svg: V sans M")
			}
			y, next, err := readNum(d, i)
			if err != nil {
				return nil, err
			}
			out = append(out, segment{cx, cy, cx, y})
			cy, i = y, next
		default:
			return nil, fmt.Errorf("svg: ordre %q", string(cmd))
		}
	}
}

func stroke(img *image.NRGBA, seg segment, width float64, ink color.NRGBA) {
	r := width / 2
	bounds := img.Bounds()
	minX := int(math.Floor(math.Min(seg.x0, seg.x1) - r))
	maxX := int(math.Ceil(math.Max(seg.x0, seg.x1) + r))
	minY := int(math.Floor(math.Min(seg.y0, seg.y1) - r))
	maxY := int(math.Ceil(math.Max(seg.y0, seg.y1) + r))
	if minX < bounds.Min.X {
		minX = bounds.Min.X
	}
	if minY < bounds.Min.Y {
		minY = bounds.Min.Y
	}
	if maxX >= bounds.Max.X {
		maxX = bounds.Max.X - 1
	}
	if maxY >= bounds.Max.Y {
		maxY = bounds.Max.Y - 1
	}
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if dist(float64(px)+0.5, float64(py)+0.5, seg) <= r {
				img.SetNRGBA(px, py, ink)
			}
		}
	}
}

func dist(px, py float64, seg segment) float64 {
	dx := seg.x1 - seg.x0
	dy := seg.y1 - seg.y0
	len2 := dx*dx + dy*dy
	if len2 == 0 {
		return math.Hypot(px-seg.x0, py-seg.y0)
	}
	t := ((px-seg.x0)*dx + (py-seg.y0)*dy) / len2
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(px-(seg.x0+t*dx), py-(seg.y0+t*dy))
}

func attr(tok xml.StartElement, name string) string {
	for _, a := range tok.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func parseHex(s string) (color.NRGBA, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") {
		return color.NRGBA{}, fmt.Errorf("svg: couleur %s", s)
	}
	hex := s[1:]
	var r, g, b uint8
	switch len(hex) {
	case 6:
		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("svg: couleur %s", s)
		}
		r = uint8(v >> 16)
		g = uint8(v >> 8)
		b = uint8(v)
	default:
		return color.NRGBA{}, fmt.Errorf("svg: couleur %s", s)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
}

func svgNeed(s, what string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, fmt.Errorf("svg: %s absente", what)
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("svg: %s", what)
	}
	return v, nil
}

func optFloat(s string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return svgNeed(s, "nombre")
}

func skipSep(d string, i int) int {
	for i < len(d) {
		switch d[i] {
		case ' ', '\t', '\n', '\r', ',':
			i++
		default:
			return i
		}
	}
	return i
}

func numberFollows(d string, i int) bool {
	i = skipSep(d, i)
	if i >= len(d) {
		return false
	}
	c := d[i]
	return c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9')
}

func readPair(d string, i int) (float64, float64, int, error) {
	x, i, err := readNum(d, i)
	if err != nil {
		return 0, 0, i, err
	}
	y, i, err := readNum(d, i)
	if err != nil {
		return 0, 0, i, err
	}
	return x, y, i, nil
}

func readNum(d string, i int) (float64, int, error) {
	i = skipSep(d, i)
	start := i
	if i < len(d) && (d[i] == '+' || d[i] == '-') {
		i++
	}
	digits := false
	for i < len(d) && d[i] >= '0' && d[i] <= '9' {
		i++
		digits = true
	}
	if i < len(d) && d[i] == '.' {
		i++
		for i < len(d) && d[i] >= '0' && d[i] <= '9' {
			i++
			digits = true
		}
	}
	if !digits {
		return 0, start, fmt.Errorf("svg: nombre")
	}
	v, err := strconv.ParseFloat(d[start:i], 64)
	if err != nil {
		return 0, start, err
	}
	return v, i, nil
}
