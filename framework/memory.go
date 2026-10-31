package framework

import (
	"sync"

	tea "charm.land/bubbletea/v2"
)

// Data est l'enregistrement pseudo `{"id": "...", "data": {...}}`.
// Les feuilles de data sont des chaînes.
type Data map[string]any

// memory est un arbre. Chaque entrée a un id.
// Set("etat", "en attente") range data.id et data.value.
// Set(Data{"id": "1", "data": Data{"id": "cyan", "value": "Cyan"}}) range ces champs.
// Get se chaîne : Get("1").Get("data").Get("value").String().
// head est partagé avec la vue lecture seule : le compteur gen suit les écritures.
type memory struct {
	head     *memHead
	readonly bool
}

type memHead struct {
	mu    sync.RWMutex
	gen   uint64
	root  *box
	lists map[string][]*memory
}

type box struct {
	text  string
	has   bool
	order []string
	kids  map[string]*box
}

func newBox() *box { return &box{kids: map[string]*box{}} }

func (b *box) childOrCreate(key string) *box {
	if b.kids == nil {
		b.kids = map[string]*box{}
	}
	if child, ok := b.kids[key]; ok {
		return child
	}
	child := newBox()
	b.kids[key] = child
	b.order = append(b.order, key)
	return child
}

// Node est une vue figée d'un nœud. Une clé absente donne un nœud vide.
type Node struct {
	text string
	kids map[string]*Node
}

func (n *Node) Get(key string) *Node {
	if n == nil || n.kids == nil {
		return &Node{}
	}
	if child, ok := n.kids[key]; ok && child != nil {
		return child
	}
	return &Node{}
}

func (n *Node) String() string {
	if n == nil {
		return ""
	}
	return n.text
}

func (b *box) view() *Node {
	if b == nil {
		return &Node{}
	}
	n := &Node{text: b.text, kids: map[string]*Node{}}
	for key, child := range b.kids {
		n.kids[key] = child.view()
	}
	return n
}

// Store est la mémoire vue par les scripts du runtime et par le code appelant.
type Store interface {
	Get(string) *Node
	Set(any, ...any)
}

func Memory() *memory {
	return &memory{head: &memHead{root: newBox(), lists: map[string][]*memory{}}}
}

func NewMemory() *memory { return Memory() }

func (m *memory) ReadOnly() *memory {
	if m == nil || m.readonly {
		return m
	}
	return &memory{head: m.head, readonly: true}
}

func (m *memory) Generation() uint64 {
	if m == nil || m.head == nil {
		return 0
	}
	m.head.mu.RLock()
	defer m.head.mu.RUnlock()
	return m.head.gen
}

func (m *memory) Set(a any, rest ...any) {
	m.mustWrite()
	switch {
	case len(rest) == 1:
		id, ok1 := a.(string)
		value, ok2 := rest[0].(string)
		if !ok1 || !ok2 || id == "" {
			panic("framework: mémoire Set")
		}
		m.putData(id, map[string]string{"id": id, "value": value})
	case len(rest) == 0:
		raw, ok := a.(Data)
		if !ok {
			panic("framework: mémoire Set")
		}
		id, _ := raw["id"].(string)
		if id == "" {
			panic("framework: mémoire sans id")
		}
		fields := map[string]string{}
		if data, ok := raw["data"].(Data); ok {
			for key, value := range data {
				text, ok := value.(string)
				if !ok {
					panic("framework: data illisible")
				}
				fields[key] = text
			}
		}
		m.putData(id, fields)
	default:
		panic("framework: mémoire Set")
	}
	notify()
}

func (m *memory) putData(id string, fields map[string]string) {
	m.head.mu.Lock()
	defer m.head.mu.Unlock()
	data := m.head.root.childOrCreate(id).childOrCreate("data")
	for key, value := range fields {
		leaf := data.childOrCreate(key)
		leaf.text = value
		leaf.has = true
	}
	m.head.gen++
}

func (m *memory) setField(id, field, value string) {
	if m == nil || m.readonly {
		panic("framework: mémoire en lecture seule")
	}
	m.head.mu.Lock()
	leaf := m.head.root.childOrCreate(id).childOrCreate(field)
	leaf.text = value
	leaf.has = true
	m.head.gen++
	m.head.mu.Unlock()
	notify()
}

func (m *memory) Get(key string) *Node {
	if m == nil || m.head.root == nil {
		return &Node{}
	}
	m.head.mu.RLock()
	defer m.head.mu.RUnlock()
	child := m.head.root.kids[key]
	if child == nil {
		return &Node{}
	}
	return child.view()
}

type memEntry struct {
	ID     string
	Fields map[string]string
}

func (m *memory) Entries() []memEntry {
	if m == nil || m.head.root == nil {
		return nil
	}
	m.head.mu.RLock()
	defer m.head.mu.RUnlock()
	out := make([]memEntry, 0, len(m.head.root.order))
	for _, id := range m.head.root.order {
		entry := m.head.root.kids[id]
		fields := map[string]string{}
		for _, key := range entry.order {
			leaf := entry.kids[key]
			if leaf != nil && leaf.has && len(leaf.kids) == 0 {
				fields[key] = leaf.text
			}
		}
		if data := entry.kids["data"]; data != nil {
			for _, key := range data.order {
				leaf := data.kids[key]
				if leaf != nil && leaf.has {
					fields[key] = leaf.text
				}
			}
		}
		if len(fields) == 0 {
			continue
		}
		out = append(out, memEntry{ID: id, Fields: fields})
	}
	return out
}

func (m *memory) SetList(key string, list []*memory) {
	m.mustWrite()
	m.head.mu.Lock()
	if m.head.lists == nil {
		m.head.lists = map[string][]*memory{}
	}
	m.head.lists[key] = append([]*memory(nil), list...)
	m.head.gen++
	m.head.mu.Unlock()
	notify()
}

func (m *memory) List(key string) []*memory {
	if m == nil {
		return nil
	}
	m.head.mu.RLock()
	defer m.head.mu.RUnlock()
	src := m.head.lists[key]
	out := make([]*memory, len(src))
	for i, child := range src {
		if m.readonly {
			out[i] = child.ReadOnly()
		} else {
			out[i] = child
		}
	}
	return out
}

// restore remplace les entrées data par la photo donnée.
// Les ajouts faits après cette photo disparaissent.
func (m *memory) restore(entries []memEntry) {
	m.mustWrite()
	root := newBox()
	for _, entry := range entries {
		if entry.ID == "" {
			continue
		}
		data := root.childOrCreate(entry.ID).childOrCreate("data")
		for key, value := range entry.Fields {
			leaf := data.childOrCreate(key)
			leaf.text = value
			leaf.has = true
		}
	}
	m.head.mu.Lock()
	m.head.root = root
	m.head.gen++
	m.head.mu.Unlock()
	notify()
}

func (m *memory) mustWrite() {
	if m == nil || m.readonly {
		panic("framework: mémoire en lecture seule")
	}
}

type refreshMsg struct{}

var refresh = make(chan struct{}, 1)

func notify() {
	select {
	case refresh <- struct{}{}:
	default:
	}
}

func WaitRefresh() tea.Cmd {
	return func() tea.Msg {
		<-refresh
		return refreshMsg{}
	}
}
