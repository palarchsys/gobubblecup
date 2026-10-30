package framework

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Action est déclenchée par un bouton.
// Chaque type ne porte que ses propres attributs.
type Action interface{ Act() tea.Cmd }

// Exec lance un script.
// Into, s'il est posé, reçoit la sortie sans quitter l'écran.
// Sans Into, le script prend le terminal le temps de son exécution.
type Exec struct {
	Script string
	Args   []string
	Into   *textArea
}

// ExecDone est le résultat d'un script.
type ExecDone struct {
	Into   *textArea
	Output string
	Err    error
}

func (e Exec) Act() tea.Cmd {
	script := strings.TrimSpace(e.Script)
	if script == "" {
		return nil
	}
	args := append([]string(nil), e.Args...)
	if e.Into == nil {
		return tea.ExecProcess(exec.Command(script, args...), func(err error) tea.Msg {
			return ExecDone{Err: err}
		})
	}
	into := e.Into
	return func() tea.Msg {
		out, err := exec.Command(script, args...).CombinedOutput()
		return ExecDone{Into: into, Output: string(out), Err: err}
	}
}

// Reload réinitialise une zone.
type Reload struct{ Target Reloader }

func (r Reload) Act() tea.Cmd {
	if r.Target != nil {
		r.Target.Reload()
	}
	return nil
}

// Call est une action libre, pour le code de l'application.
type Call struct{ Run func() }

func (c Call) Act() tea.Cmd {
	if c.Run != nil {
		c.Run()
	}
	return nil
}
