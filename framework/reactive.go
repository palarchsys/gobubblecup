package framework

// reactiveObject reconstruit des valeurs depuis les data d'une mémoire.
// Le constructeur est ListItem pour une liste, TabItem pour une barre d'onglets.
// fields associe un champ du constructeur à un champ de data : "label" -> "value".
type reactiveObject[T any] struct {
	ctor   func(id, label string) T
	memo   *memory
	fields map[string]string
	origin []memEntry
}

func cloneEntries(in []memEntry) []memEntry {
	out := make([]memEntry, len(in))
	for i, entry := range in {
		fields := make(map[string]string, len(entry.Fields))
		for key, value := range entry.Fields {
			fields[key] = value
		}
		out[i] = memEntry{ID: entry.ID, Fields: fields}
	}
	return out
}

// Reload rend à la mémoire les entrées qu'elle avait à la création de l'objet.
func (r *reactiveObject[T]) Reload() {
	if r == nil || r.memo == nil {
		return
	}
	r.memo.restore(cloneEntries(r.origin))
}

func ReactiveObject[T any](ctor func(id, label string) T, memo *memory, fields map[string]string) *reactiveObject[T] {
	if ctor == nil || memo == nil {
		panic("framework: objet réactif incomplet")
	}
	copied := map[string]string{}
	for key, value := range fields {
		copied[key] = value
	}
	return &reactiveObject[T]{ctor: ctor, memo: memo, fields: copied, origin: cloneEntries(memo.Entries())}
}

func (r *reactiveObject[T]) isListArg() {}

func (r *reactiveObject[T]) isTabArg() {}

func (r *reactiveObject[T]) Items() []T {
	if r == nil || r.memo == nil || r.ctor == nil {
		return nil
	}
	idField := r.fields["id"]
	labelField := r.fields["label"]
	dotField := r.fields["dot"]
	out := make([]T, 0)
	for _, entry := range r.memo.Entries() {
		built := r.ctor(entry.Fields[idField], entry.Fields[labelField])
		if dotField != "" {
			built = applyDot(built, entry.Fields[dotField])
		}
		out = append(out, built)
	}
	return out
}

// applyDot copie la couleur sur un Item. Les autres types restent intacts.
func applyDot[T any](value T, color string) T {
	item, ok := any(value).(Item)
	if !ok {
		return value
	}
	item.Dot = color
	return any(item).(T)
}

// writeDot écrit la couleur dans la mémoire, sous le champ mappé par "dot".
// Faux si la carte n'a pas ce champ, ou si aucun item ne porte cet id.
func (r *reactiveObject[T]) writeDot(id, color string) bool {
	if r == nil || r.memo == nil || id == "" {
		return false
	}
	field := r.fields["dot"]
	if field == "" {
		return false
	}
	idField := r.fields["id"]
	entryID := ""
	found := false
	for _, entry := range r.memo.Entries() {
		if entry.Fields[idField] == id {
			entryID = entry.ID
			found = true
			break
		}
	}
	if !found {
		return false
	}
	r.memo.mustWrite()
	r.memo.putData(entryID, map[string]string{field: color})
	notify()
	return true
}
