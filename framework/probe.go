package framework

import "strings"

func slotText(store Store, id string) (string, string) {
	if store == nil || id == "" {
		return "", ""
	}
	node := store.Get(id)
	return node.Get("status").String(), node.Get("return").String()
}

func contentLines(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "", "success", "warning", "error":
			continue
		default:
			out = append(out, line)
		}
	}
	return out
}

type hostInfo struct {
	name    string
	id      string
	version string
	install string
	known   bool
}

func readHost(body string) hostInfo {
	var info hostInfo
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "hostname":
			info.name = value
			info.known = true
		case "os":
			info.id = value
			info.known = true
		case "version":
			info.version = value
			info.known = true
		case "install":
			info.install = value
			info.known = true
		}
	}
	return info
}

func (h hostInfo) system() string {
	if !h.known {
		return "—"
	}
	name := h.id
	if strings.EqualFold(name, "ubuntu") {
		name = "Ubuntu"
	}
	switch {
	case name != "" && h.version != "":
		return name + " " + h.version
	case name != "":
		return name
	case h.version != "":
		return h.version
	default:
		return "—"
	}
}

func (h hostInfo) installText() string {
	if !h.known {
		return "—"
	}
	switch h.install {
	case "present":
		return "présente"
	case "absent":
		return "absente"
	case "":
		return "—"
	default:
		return h.install
	}
}

func (h hostInfo) nameText() string {
	if !h.known || h.name == "" {
		return "—"
	}
	return h.name
}

func hostLines(status, body string) []Line {
	info := readHost(body)
	switch status {
	case "pending":
		return []Line{{Mark: Info, Text: "En cours"}}
	case "":
		return []Line{{Mark: Info, Text: "En attente"}}
	}
	rows := []Line{{Mark: OK, Text: "Exécution"}}
	if info.name != "" {
		rows = append(rows, Line{Mark: Info, Text: "Hôte : " + info.name})
	}
	if system := info.system(); system != "—" {
		rows = append(rows, Line{Mark: Info, Text: "OS : " + system})
	}
	switch status {
	case "success":
		rows = append(rows, Line{Mark: OK, Text: "Installation présente"})
	case "warning":
		rows = append(rows, Line{Mark: Warn, Text: "Installation absente"})
	default:
		rows = append(rows, Line{Mark: Bad, Text: "Lecture refusée"})
	}
	return rows
}

func runLines(status, body string) []Line {
	switch status {
	case "pending":
		return []Line{{Mark: Info, Text: "En cours"}}
	case "":
		return []Line{{Mark: Info, Text: "En attente"}}
	}
	kind := Bad
	phrase := "Refusé"
	switch status {
	case "success":
		kind = OK
		phrase = "Terminé"
	case "warning":
		kind = Warn
		phrase = "Avertissement"
	}
	lines := contentLines(body)
	if len(lines) == 0 {
		return []Line{{Mark: kind, Text: phrase}}
	}
	rows := make([]Line, len(lines))
	for i, line := range lines {
		mark := Info
		if i == len(lines)-1 {
			mark = kind
		}
		rows[i] = Line{Mark: mark, Text: line}
	}
	return rows
}

// HostFacts relit le slot et montre l'hôte, le système et l'installation.
func HostFacts(store Store, id string) *facts {
	load := func() hostInfo {
		_, body := slotText(store, id)
		return readHost(body)
	}
	return Facts(
		Fact{Label: "Hôte", Text: func() string { return load().nameText() }},
		Fact{Label: "OS", Text: func() string { return load().system() }},
		Fact{Label: "Installation", Text: func() string { return load().installText() }},
	)
}

// HostReport dessine la sonde comme un arbre. path est le titre.
func HostReport(store Store, id, path string) *stepView {
	return Step(path, func() []Line {
		status, body := slotText(store, id)
		return hostLines(status, body)
	})
}

// RunReport dessine l'exécution d'un script. Les mots success, warning et error ne sont pas répétés.
func RunReport(store Store, id, path string) *stepView {
	return Step(path, func() []Line {
		status, body := slotText(store, id)
		return runLines(status, body)
	})
}
