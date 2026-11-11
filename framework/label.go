package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// labelLangs est l'ensemble lu par Labels. L'ordre est fixe.
var labelLangs = []string{"en", "fr", "de", "es"}

// Labels lit dir/<langue>/label.txt pour en, fr, de et es.
// Une clé [LABEL_NOM] est stockée sous NOM. Une clé non demandée peut manquer.
// Un fichier illisible ou une ligne hors clé renvoie une erreur.
func Labels(dir string) (map[string]map[string]string, error) {
	out := make(map[string]map[string]string, len(labelLangs))
	for _, lang := range labelLangs {
		got, err := readLabels(filepath.Join(dir, lang, "label.txt"))
		if err != nil {
			return nil, err
		}
		out[lang] = got
	}
	return out, nil
}

// Catalog lit dir/<langue>/file pour en, fr, de et es.
// Une clé [NOM] est stockée telle quelle. Une ligne vide ou qui commence par # est ignorée.
// Une clé non demandée peut manquer.
// Un fichier illisible, une clé vide ou une ligne hors clé renvoie une erreur.
func Catalog(dir, file string) (map[string]map[string]string, error) {
	out := make(map[string]map[string]string, len(labelLangs))
	for _, lang := range labelLangs {
		got, err := readCatalog(filepath.Join(dir, lang, file))
		if err != nil {
			return nil, err
		}
		out[lang] = got
	}
	return out, nil
}

func readCatalog(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("catalogue %s illisible", path)
	}
	got := map[string]string{}
	key := ""
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			if name == "" {
				return nil, fmt.Errorf("%s : clé vide", path)
			}
			key = name
			continue
		}
		if key == "" {
			return nil, fmt.Errorf("%s : ligne hors clé", path)
		}
		got[key] = line
		key = ""
	}
	return got, nil
}

func readLabels(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("libellés %s illisibles", path)
	}
	got := map[string]string{}
	key := ""
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[LABEL_") && strings.HasSuffix(line, "]") {
			name := strings.TrimSuffix(strings.TrimPrefix(line, "[LABEL_"), "]")
			if name == "" {
				return nil, fmt.Errorf("%s : nom vide", path)
			}
			key = name
			continue
		}
		if key == "" {
			return nil, fmt.Errorf("%s : ligne hors clé", path)
		}
		got[key] = line
		key = ""
	}
	return got, nil
}
