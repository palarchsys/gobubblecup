package framework

import (
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Action est déclenchée par un bouton.
// Chaque type ne porte que ses propres attributs.
type Action interface{ Act() tea.Cmd }

// Sink reçoit un texte. Un champ et une zone de texte le font.
type Sink interface{ SetText(string) }

// Exec lance un script.
// Into, s'il est posé, reçoit la sortie sans quitter l'écran.
// Sans Into, le script prend le terminal le temps de son exécution.
type Exec struct {
	Script string
	Args   []string
	Into   Sink
}

// ExecDone est le résultat d'un script.
type ExecDone struct {
	Into   Sink
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

// ReadFile charge un fichier dans la variable, sous l'id "file".
type ReadFile struct {
	Path string
	Into Store
}

func (r ReadFile) Act() tea.Cmd {
	if strings.TrimSpace(r.Path) == "" || r.Into == nil {
		panic("framework: lecture de fichier incomplète")
	}
	text, err := os.ReadFile(r.Path)
	value := string(text)
	if err != nil {
		value = err.Error()
	}
	r.Into.Set("file", value)
	return nil
}

// WriteFile enregistre la variable "file" dans un fichier.
type WriteFile struct {
	Path string
	From Store
}

func (w WriteFile) Act() tea.Cmd {
	if strings.TrimSpace(w.Path) == "" || w.From == nil {
		panic("framework: écriture de fichier incomplète")
	}
	value := w.From.Get("file").Get("data").Get("value").String()
	if err := os.WriteFile(w.Path, []byte(value), 0o644); err != nil {
		w.From.Set("file", err.Error())
	}
	return nil
}
