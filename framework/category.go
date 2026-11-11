package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Category est un contour dont le sous-titre est posé sur le trait haut.
// Elle contient d'autres objets, des listes par exemple, et ne prend pas le focus.
type category struct {
	legend   string
	kids     []Widget
	disabled bool
	held     []Widget
}

func Category(legend string, kids ...Widget) *category {
	for _, kid := range kids {
		if kid == nil {
			panic("framework: objet nul dans la catégorie")
		}
	}
	return &category{legend: legend, kids: kids}
}

func (c *category) Kids() []Widget { return c.kids }

// SetLegend remplace le titre posé sur le trait haut.
func (c *category) SetLegend(legend string) {
	if c == nil || c.legend == legend {
		return
	}
	c.legend = legend
	notify()
}

func (c *category) Active() bool {
	if c == nil || c.disabled {
		return false
	}
	for _, kid := range c.kids {
		if kid.Active() {
			return true
		}
	}
	return false
}

func (c *category) Disable() {
	if c == nil || c.disabled {
		return
	}
	c.disabled = true
	c.held = nil
	for _, kid := range c.kids {
		if kid.Disabled() {
			continue
		}
		kid.Disable()
		c.held = append(c.held, kid)
	}
	notify()
}

func (c *category) Enable() {
	if c == nil || !c.disabled {
		return
	}
	c.disabled = false
	for _, kid := range c.held {
		kid.Enable()
	}
	c.held = nil
	notify()
}

func (c *category) Disabled() bool { return c != nil && c.disabled }

func (c *category) SetActive(on bool) {
	for _, kid := range c.kids {
		kid.SetActive(on)
	}
}

func (c *category) Reload() {
	for _, kid := range c.kids {
		if zone, ok := kid.(Reloader); ok {
			zone.Reload()
		}
	}
}

func (c *category) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (c *category) View(s Style, focused Widget, inside bool) string {
	if c.Disabled() {
		s = s.dimmed()
	}
	padX, padY := s.padX(), s.padY()
	width := 0
	if s.fit {
		width = s.Width
	}
	childBase := s
	childBase.Height = 0
	childBase.fit = false
	if s.fit {
		border := s.content(width)
		if border < 1 {
			border = 1
		}
		childBase = s.placed(border, 0)
	}

	parts := make([]string, len(c.kids))
	heights := make([]int, len(c.kids))
	innerW := 0
	total := 0
	for i, kid := range c.kids {
		parts[i] = kid.View(childBase, focused, inside)
		if w := lineWidth(parts[i]); w > innerW {
			innerW = w
		}
		heights[i] = lineCount(parts[i])
		total += heights[i]
	}

	natural := innerW + 2*padX + 2
	if width < natural {
		width = natural
	}
	height := 0
	if s.fit {
		height = s.Height
	}
	extra := 0
	if height > 0 {
		room := height - 2 - 2*padY
		if room > total {
			extra = room - total
		}
	}
	target := s.content(width)
	if target < 1 {
		target = 1
	}
	if len(c.kids) > 0 && (target != innerW || extra > 0) {
		for i, kid := range c.kids {
			h := 0
			if i == len(c.kids)-1 && extra > 0 {
				h = heights[i] + extra
			}
			if lineWidth(parts[i]) == target && h == 0 {
				continue
			}
			parts[i] = kid.View(s.placed(target, h), focused, inside)
		}
	}
	body := ""
	if len(parts) > 0 {
		body = strings.Join(parts, "\n")
	}
	return s.box(body, frameIdle, width, height, c.legend)
}
