package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// App branche la grille sur Bubble Tea.
// La grille est un pointeur : les changements survivent à la copie reçue par Update.
type App struct {
	style         Style
	grid          *grid
	width, height int
}

func New(style Style, grid *grid) App {
	if grid == nil {
		grid = NewGrid()
	}
	lists := grid.tabLists()
	if len(lists) > 1 {
		panic("framework: une seule liste d'onglets par interface")
	}
	if len(lists) == 1 {
		if w := lists[0].firstFocus(); w != nil {
			grid.focus = w
			grid.inside = false
		}
	}
	return App{style: style, grid: grid}
}

func (a App) Init() tea.Cmd { return tea.Batch(tea.ClearScreen, WaitRefresh()) }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshMsg:
		return a, WaitRefresh()
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" || (key == "q" && !a.grid.Inside()) {
			return a, tea.Quit
		}
		return a, a.grid.Key(key)
	case ExecDone:
		if msg.Into != nil {
			text := strings.TrimRight(msg.Output, "\n")
			if msg.Err != nil {
				if text != "" {
					text += "\n"
				}
				text += msg.Err.Error()
			}
			msg.Into.SetText(text)
		}
		return a, nil
	default:
		return a, nil
	}
}

func (a App) View() tea.View {
	s := a.style
	rows := a.grid.render(s)
	width := 0
	for _, row := range rows {
		if w := lineWidth(row); w > width {
			width = w
		}
	}
	help := NewLegend().View(s, nil, false)
	if w := lipgloss.Width(help); w > width {
		width = w
	}
	blocks := make([]string, 0, len(rows)+1)
	for _, row := range rows {
		blocks = append(blocks, widen(s, row, width))
	}
	blocks = append(blocks, widen(s, help, width))
	body := inset(s, strings.Join(blocks, "\n"), s.margin())
	body = fillScreen(s, body, a.width, a.height)
	v := tea.NewView(body)
	v.BackgroundColor = s.paper()
	return v
}

func widen(s Style, block string, width int) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		if w := lipgloss.Width(line); w < width {
			line += s.shade(width - w)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func inset(s Style, block string, margin int) string {
	if margin <= 0 {
		return block
	}
	lines := strings.Split(block, "\n")
	inner := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > inner {
			inner = w
		}
	}
	side := s.shade(margin)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		pad := ""
		if w := lipgloss.Width(line); w < inner {
			pad = s.shade(inner - w)
		}
		out = append(out, side+line+pad+side)
	}
	return strings.Join(out, "\n")
}

func fillScreen(s Style, block string, width, height int) string {
	if width < 1 && height < 1 {
		return block
	}
	lines := strings.Split(block, "\n")
	if width > 0 {
		for i, line := range lines {
			if w := lipgloss.Width(line); w < width {
				lines[i] = line + s.shade(width-w)
			}
		}
	}
	rowW := width
	if rowW < 1 && len(lines) > 0 {
		rowW = lipgloss.Width(lines[0])
	}
	for height > len(lines) {
		lines = append(lines, s.shade(rowW))
	}
	return strings.Join(lines, "\n")
}
