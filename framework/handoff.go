package framework

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// Handoff efface l'écran, se place dans dir, puis remplace le processus.
// bin est le programme. argv est la ligne reçue par ce programme, argv[0] compris.
// Un bin vide arrête l'appel et rend une erreur. dir vide ne change pas de dossier.
func Handoff(dir, bin string, argv ...string) error {
	if strings.TrimSpace(bin) == "" {
		return fmt.Errorf("framework: relais vide")
	}
	fmt.Print("\033[2J\033[H")
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return err
		}
	}
	if len(argv) == 0 {
		argv = []string{bin}
	}
	return syscall.Exec(bin, argv, os.Environ())
}
