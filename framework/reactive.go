package framework

// reactiveObject reconstruit des items depuis les data d'une mémoire.
// fields associe un champ de ListItem à un champ de data : "label" -> "value".
type reactiveObject struct {
	ctor   func(id, label string) Item
	memo   *memory
	fields map[string]string
}

func ReactiveObject(ctor func(id, label string) Item, memo *memory, fields map[string]string) *reactiveObject {
	if ctor == nil || memo == nil {
		panic("framework: objet réactif incomplet")
	}
	copied := map[string]string{}
	for key, value := range fields {
		copied[key] = value
	}
	return &reactiveObject{ctor: ctor, memo: memo, fields: copied}
}

func NewReactiveObject(ctor func(id, label string) Item, memo *memory, fields map[string]string) *reactiveObject {
	return ReactiveObject(ctor, memo, fields)
}

func (r *reactiveObject) isListArg() {}

func (r *reactiveObject) Items() []Item {
	if r == nil || r.memo == nil || r.ctor == nil {
		return nil
	}
	idField := r.fields["id"]
	labelField := r.fields["label"]
	out := make([]Item, 0)
	for _, entry := range r.memo.Entries() {
		out = append(out, r.ctor(entry.Fields[idField], entry.Fields[labelField]))
	}
	return out
}
