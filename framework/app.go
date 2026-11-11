package framework

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// KeyMap remplace les touches du cadre.
// Quit vide reste ctrl+c. Paste vide reste ctrl+v.
// On associe une touche de l'écran à une commande. Quitter et coller passent avant.
type KeyMap struct {
	Quit       string
	Paste      string
	HideLegend bool
	On         map[string]func() Cmd
}

// Cmd est la commande du cadre. Une fonction qui rend nil reste acceptée.
type Cmd = tea.Cmd

// Quit termine le programme.
func Quit() Cmd { return tea.Quit }

// App branche la grille sur Bubble Tea.
// La grille est un pointeur : les changements survivent à la copie reçue par Update.
type App struct {
	style         Style
	win           *window
	grid          *grid
	width, height int
	keys          KeyMap
	bound         bool
}

func New(style Style, grid *grid) App {
	if grid == nil {
		grid = Grid()
	}
	w := Window()
	w.Use(grid)
	return NewWindow(style, w)
}

// NewWindow affiche une fenêtre. Suivante et Précédente changent la grille.
func NewWindow(style Style, w *window) App {
	if w == nil || len(w.pages) == 0 {
		panic("framework: fenêtre vide")
	}
	for _, page := range w.pages {
		lists := page.tabLists()
		if len(lists) > 1 {
			panic("framework: une seule liste d'onglets par interface")
		}
		if len(lists) == 1 {
			if zone := lists[0].firstFocus(); zone != nil {
				page.focus = zone
				page.inside = false
			}
		}
		page.follow()
	}
	return App{style: style, win: w, grid: w.current()}
}

// Bind choisit les touches quitter et coller, et peut masquer la légende stock.
// Sans Bind, ctrl+c quitte, ctrl+v colle, et la légende reste.
// Le résultat remplace l'application.
func (a App) Bind(keys KeyMap) App {
	a.keys = keys
	a.bound = true
	return a
}

// Run lance le programme et rend l'erreur. L'écran alternatif reste celui de View.
func (a App) Run() error {
	_, err := tea.NewProgram(a).Run()
	return err
}

func (a App) quitKey() string {
	if a.bound && a.keys.Quit != "" {
		return a.keys.Quit
	}
	return "ctrl+c"
}

func (a App) pasteKey() string {
	if a.bound && a.keys.Paste != "" {
		return a.keys.Paste
	}
	return "ctrl+v"
}

func (a App) hideLegend() bool { return a.bound && a.keys.HideLegend }

func (a App) sync() App {
	if a.win != nil && len(a.win.pages) > 0 {
		a.grid = a.win.current()
	}
	return a
}

func (a App) Init() tea.Cmd { return WaitRefresh() }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	a = a.sync()
	switch msg := msg.(type) {
	case refreshMsg:
		return a.sync(), WaitRefresh()
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a.sync(), nil
	case tea.PasteMsg:
		a.grid.paste(msg.Content)
		return a.sync(), nil
	case tea.ClipboardMsg:
		if msg.Clipboard() == 'c' || msg.Clipboard() == 0 {
			a.grid.paste(msg.String())
		}
		return a.sync(), nil
	case tea.KeyPressMsg:
		key := msg.String()
		if key == a.quitKey() {
			cancelRuntimes()
			return a.sync(), tea.Quit
		}
		if key == a.pasteKey() && a.grid.canPaste() {
			return a.sync(), func() tea.Msg { return tea.ReadClipboard() }
		}
		if a.bound && key == a.pasteKey() {
			return a.sync(), nil
		}
		if a.bound {
			if fn, ok := a.keys.On[key]; ok {
				if fn == nil {
					return a.sync(), nil
				}
				return a.sync(), fn()
			}
		}
		cmd := a.grid.Key(key)
		return a.sync(), cmd
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
			if zone, ok := msg.Into.(Widget); ok && a.win != nil {
				a.win.reveal(zone)
			}
		}
		return a.sync(), nil
	default:
		return a.sync(), nil
	}
}

func (a App) View() tea.View {
	a = a.sync()
	s := a.style
	// Height du style est le format de l'écran. Les objets gardent la leur à zéro.
	format := s.Height
	s.Height = 0
	help := ""
	legend := 0
	if !a.hideLegend() {
		help = Legend().View(s, nil, false)
		legend = lineCount(help)
	}
	budget, fill := 0, false
	if format > 0 {
		budget = format - legend
		if budget < 1 {
			budget = 1
		}
		fill = true
	}
	if a.height > 0 {
		term := a.height - legend
		if term < 1 {
			term = 1
		}
		if budget == 0 || budget > term {
			if format == 0 {
				fill = false
			}
			budget = term
		}
	}
	rows := a.grid.layout(s, budget, fill)
	width := 0
	for _, row := range rows {
		if w := lineWidth(row); w > width {
			width = w
		}
	}
	if help != "" {
		if w := lipgloss.Width(help); w > width {
			width = w
		}
	}
	blocks := make([]string, 0, len(rows)+1)
	for _, row := range rows {
		blocks = append(blocks, widen(s, row, width))
	}
	if help != "" {
		blocks = append(blocks, widen(s, help, width))
	}
	body := inset(s, strings.Join(blocks, "\n"), s.margin())
	screenH := 0
	if format > 0 {
		screenH = format
		if a.height > 0 && screenH > a.height {
			screenH = a.height
		}
	}
	body = fillScreen(s, body, a.width, screenH)
	v := tea.NewView(body)
	v.BackgroundColor = s.paper()
	// L'écran alternatif efface le terminal au lancement. La sortie le rend.
	v.AltScreen = true
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
