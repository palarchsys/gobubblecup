package framework

// StatusLine classe un script dont le mot success, warning ou error occupe toute une ligne.
const StatusLine = `(?m)^(success|warning|error)$`

// script prépare un seul bash et sait le dessiner.
type script struct {
	*runtime
	id   string
	path string
}

// Script enregistre le bash et ne le lance pas.
// env est relu au départ. Une fonction nulle n'ajoute pas de clés.
// La classification est StatusLine.
func Script(dir, id, label, path string, args []string, env func() map[string]string) *script {
	rt := Runtime(RuntimeConfig{
		Dir:         dir,
		StatusRegex: StatusLine,
		Functions: func(Store) {
			var values map[string]string
			if env != nil {
				values = env()
			}
			BashEnv(id, label, path, args, values)
		},
	})
	rt.Init()
	return &script{runtime: rt, id: id, path: path}
}

// Report dessine l'exécution. Le titre est le chemin.
func (s *script) Report() *stepView {
	if s == nil {
		return nil
	}
	return RunReport(s.Memory, s.id, s.path)
}

// Host dessine la sonde. Le titre est le chemin.
func (s *script) Host() *stepView {
	if s == nil {
		return nil
	}
	return HostReport(s.Memory, s.id, s.path)
}

// Facts aligne l'hôte, le système et l'installation lus dans le slot.
func (s *script) Facts() *facts {
	if s == nil {
		return nil
	}
	return HostFacts(s.Memory, s.id)
}
