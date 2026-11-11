package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindRoot cherche la racine d'un produit.
// env, s'il est posé et non vide dans l'environnement, gagne.
// Sinon le dossier install, puis le dépôt folder à côté d'un répertoire go/.
// Une racine sans script/<name> est ignorée.
func FindRoot(env, install, folder, script string) (string, error) {
	script = strings.TrimSpace(script)
	folder = strings.TrimSpace(folder)
	if script == "" || folder == "" {
		return "", fmt.Errorf("framework: racine incomplète")
	}
	if env = strings.TrimSpace(env); env != "" {
		if root := os.Getenv(env); root != "" {
			if hasScript(root, script) {
				return root, nil
			}
			return "", fmt.Errorf("%s %s ne contient pas script/%s", env, root, script)
		}
	}
	if install != "" && hasScript(install, script) {
		return install, nil
	}
	if root := siblingRoot(folder, script); root != "" {
		return root, nil
	}
	return "", fmt.Errorf("racine %s introuvable, avec script/%s", folder, script)
}

// RequireRoot refuse un programme qui n'est pas lancé par root.
func RequireRoot() error {
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf("ce programme écrit la configuration. Relancez-le avec sudo")
}

func hasScript(root, script string) bool {
	info, err := os.Stat(filepath.Join(root, "script", script))
	return err == nil && !info.IsDir()
}

func siblingRoot(folder, script string) string {
	var starts []string
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	seen := map[string]bool{}
	for _, start := range starts {
		dir := start
		for range 8 {
			if seen[dir] {
				break
			}
			seen[dir] = true
			if filepath.Base(dir) == "go" {
				candidate := filepath.Join(filepath.Dir(dir), folder)
				if hasScript(candidate, script) {
					return candidate
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}
