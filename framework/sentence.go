package framework

import (
	"os"
	"strings"
)

// Sentence lit un fichier de blocs. Le programme passe les noms et les textes.
type Sentence struct {
	Path string
	Mark string
	Hide string
	Both []string
}

// Text choisit le dernier bloc qui contient name.
// equal vrai exige que le corps contienne aussi chaque nom de Both.
// Sinon le corps contient name et ne contient pas tous ces noms.
// {1} devient Mark. {2} devient value, sauf quand name est Hide.
// Un fichier illisible ou un bloc absent rend name.
func (s Sentence) Text(name string, equal bool, value string) string {
	raw, err := os.ReadFile(s.Path)
	if err != nil {
		return name
	}
	chosen := ""
	for _, part := range strings.Split(string(raw), "\n[") {
		text := part
		if i := strings.IndexByte(text, '\n'); i >= 0 {
			text = strings.TrimSpace(text[i+1:])
		}
		if strings.Contains(text, name) && equal == holdsAll(text, s.Both) {
			chosen = text
		}
	}
	if chosen == "" {
		return name
	}
	chosen = strings.ReplaceAll(chosen, "{1}", s.Mark)
	if name != s.Hide {
		chosen = strings.ReplaceAll(chosen, "{2}", value)
	}
	return chosen
}

func holdsAll(text string, names []string) bool {
	for _, name := range names {
		if !strings.Contains(text, name) {
			return false
		}
	}
	return true
}

// Assign remplace les lignes clé=valeur, ajoute les clés absentes et garde le reste.
// Un fichier absent reçoit head. Le mode est posé après l'écriture.
// Les valeurs sont écrites telles quelles.
func Assign(path string, mode os.FileMode, head string, pairs [][2]string) error {
	raw, err := os.ReadFile(path)
	fresh := os.IsNotExist(err)
	if err != nil && !fresh {
		return err
	}
	body := string(raw)
	if fresh {
		body = head
	}
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	lines := strings.Split(body, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	seen := map[string]bool{}
	for i, line := range lines {
		for _, pair := range pairs {
			if strings.HasPrefix(line, pair[0]+"=") {
				lines[i] = pair[0] + "=" + pair[1]
				seen[pair[0]] = true
				break
			}
		}
	}
	for _, pair := range pairs {
		if seen[pair[0]] {
			continue
		}
		lines = append(lines, pair[0]+"="+pair[1])
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), mode); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

// Safe est faux dès que value contient un caractère de forbidden.
func Safe(value, forbidden string) bool {
	return !strings.ContainsAny(value, forbidden)
}
