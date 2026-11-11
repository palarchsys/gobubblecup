package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// flagRow est une rangée de drapeaux alignée à droite. Elle ne prend pas le focus.
// Au départ aucun n'est choisi : tous sont en couleur.
type flagRow struct {
	codes    []string
	pics     []*picture
	selected string
	disabled bool
}

// Flags lit dir/<code>.svg dans l'ordre et aligne les glyphes à droite.
func Flags(dir string, order []string, cols int) (*flagRow, error) {
	if cols < 1 {
		return nil, fmt.Errorf("drapeau: largeur")
	}
	row := &flagRow{}
	for _, code := range order {
		code = strings.TrimSpace(code)
		if code == "" {
			return nil, fmt.Errorf("drapeau: code vide")
		}
		raw, err := os.ReadFile(filepath.Join(dir, code+".svg"))
		if err != nil {
			return nil, fmt.Errorf("drapeau %s introuvable", code)
		}
		name := "flag:" + code
		if err := IncludeSVG(name, raw); err != nil {
			return nil, fmt.Errorf("drapeau %s : %s", code, err.Error())
		}
		row.codes = append(row.codes, code)
		row.pics = append(row.pics, Glyph(name, cols))
	}
	return row, nil
}

// Select colore code et grise les autres. Un code absent ne change rien.
func (f *flagRow) Select(code string) {
	if f == nil {
		return
	}
	found := false
	for _, item := range f.codes {
		if item == code {
			found = true
			break
		}
	}
	if !found {
		return
	}
	f.selected = code
	for i, item := range f.codes {
		f.pics[i].SetGray(item != code)
	}
}

// Selected rend le code coloré. Vide tant qu'aucun n'est choisi.
func (f *flagRow) Selected() string {
	if f == nil {
		return ""
	}
	return f.selected
}

func (f *flagRow) Active() bool { return false }

func (f *flagRow) SetActive(bool) {}

func (f *flagRow) Disable() {
	if f == nil || f.disabled {
		return
	}
	f.disabled = true
	notify()
}

func (f *flagRow) Enable() {
	if f == nil || !f.disabled {
		return
	}
	f.disabled = false
	notify()
}

func (f *flagRow) Disabled() bool { return f != nil && f.disabled }

func (f *flagRow) Handle(bool, string) (bool, tea.Cmd) { return false, nil }

func (f *flagRow) View(s Style, focused Widget, inside bool) string {
	if f == nil || len(f.pics) == 0 {
		return ""
	}
	child := s
	if s.fit {
		child = s.preferred()
	}
	parts := make([]string, 0, len(f.pics)*2)
	for i, pic := range f.pics {
		if i > 0 {
			parts = append(parts, child.shade(1))
		}
		parts = append(parts, pic.View(child, focused, inside))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	if s.fit && s.Width > lineWidth(body) {
		body = padLeft(s, body, s.Width-lineWidth(body))
	}
	return body
}

func padLeft(s Style, block string, n int) string {
	if n <= 0 || block == "" {
		return block
	}
	side := s.shade(n)
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = side + line
	}
	return strings.Join(lines, "\n")
}
