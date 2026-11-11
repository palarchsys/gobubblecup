package framework

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Rule est un contrôle de schéma. Le cadre ne choisit pas la phrase d'alerte.
type Rule struct {
	Name      string   `json:"name"`
	Pattern   string   `json:"pattern"`
	Equals    string   `json:"equals"`
	Min       int      `json:"minimum_length"`
	Max       int      `json:"maximum_length"`
	PortMin   int      `json:"port_minimum"`
	PortMax   int      `json:"port_maximum"`
	Forbidden []string `json:"forbidden"`
	Reject    []string `json:"reject"`
}

// LoadRules lit un document {"variables":[...]} et rend les règles dans l'ordre du fichier.
func LoadRules(path string) ([]Rule, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Variables []Rule `json:"variables"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	return doc.Variables, nil
}

// Fail rend le premier nom en échec. equal dit que deux champs diffèrent.
// ok est vrai quand tous les noms passent. Un nom absent est un échec.
func Fail(rules []Rule, names, values []string) (name string, equal, ok bool) {
	byName := make(map[string]Rule, len(rules))
	for _, rule := range rules {
		byName[rule.Name] = rule
	}
	for i, item := range names {
		rule, found := byName[item]
		if !found || i >= len(values) {
			return item, false, false
		}
		other := ""
		if rule.Equals != "" {
			for j, candidate := range names {
				if candidate == rule.Equals && j < len(values) {
					other = values[j]
					break
				}
			}
		}
		bad, same := rule.hit(values[i], other)
		if bad {
			return item, same, false
		}
	}
	return "", false, true
}

func (item Rule) hit(value, other string) (bad, equal bool) {
	if item.Pattern != "" && !ruleMatch(item.Pattern, value) {
		return true, false
	}
	n := utf8.RuneCountInString(value)
	if (item.Min > 0 && n < item.Min) || (item.Max > 0 && n > item.Max) {
		return true, false
	}
	if ruleContains(value, item.Forbidden) || ruleContains(value, item.Reject) {
		return true, false
	}
	if item.PortMax > 0 {
		_, port, cut := strings.Cut(value, ":")
		number, err := strconv.Atoi(port)
		if !cut || err != nil || number < item.PortMin || number > item.PortMax {
			return true, false
		}
	}
	if item.Equals != "" && value != other {
		return true, true
	}
	return false, false
}

func ruleMatch(pattern, value string) bool {
	pattern = strings.ReplaceAll(pattern, "[:space:]", " \t\r\n\v\f")
	re, err := regexp.Compile(pattern)
	return err == nil && re.MatchString(value)
}

func ruleContains(value string, tokens []string) bool {
	for _, token := range tokens {
		if token != "" && strings.Contains(value, token) {
			return true
		}
	}
	return false
}
