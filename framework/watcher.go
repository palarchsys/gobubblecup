package framework

// WatcherConfig branche une fonction libre sur une ou plusieurs mémoires.
// À chaque modification de l'une d'elles, Functions est rappelée et son résultat
// remplace celui du ReactiveWatcher.
type WatcherConfig struct {
	Watcher   []Store
	Functions func(args []Store) any
}

// reactiveWatcher garde le dernier résultat calculé par le développeur.
// S'il retourne un ReactiveObject, une liste d'items, ou un autre watcher,
// Items expose ces lignes.
type reactiveWatcher struct {
	watch  []Store
	fn     func([]Store) any
	gen    []uint64
	ready  bool
	result any
}

func ReactiveWatcher(cfg WatcherConfig) *reactiveWatcher {
	if len(cfg.Watcher) == 0 || cfg.Functions == nil {
		panic("framework: watcher incomplet")
	}
	watch := make([]Store, len(cfg.Watcher))
	for i, store := range cfg.Watcher {
		if store == nil {
			panic("framework: mémoire suivie nulle")
		}
		watch[i] = store
	}
	w := &reactiveWatcher{watch: watch, fn: cfg.Functions}
	w.refresh()
	return w
}

func (w *reactiveWatcher) isListArg() {}

// Result renvoie le dernier résultat, recalculé si la mémoire suivie a changé.
func (w *reactiveWatcher) Result() any {
	if w == nil {
		return nil
	}
	w.refresh()
	return w.result
}

func (w *reactiveWatcher) Items() []Item {
	if w == nil {
		return nil
	}
	w.refresh()
	return itemsOf(w.result)
}

func (w *reactiveWatcher) refresh() {
	gen := generationsOf(w.watch)
	if w.ready && sameGen(gen, w.gen) {
		return
	}
	w.result = w.fn(w.watch)
	w.gen = gen
	w.ready = true
}

func generationsOf(list []Store) []uint64 {
	out := make([]uint64, len(list))
	for i, store := range list {
		out[i] = generationOf(store)
	}
	return out
}

func sameGen(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generationOf(store Store) uint64 {
	memo, ok := store.(*memory)
	if !ok {
		return 0
	}
	return memo.Generation()
}

func itemsOf(result any) []Item {
	switch v := result.(type) {
	case *reactiveObject[Item]:
		return v.Items()
	case *reactiveWatcher:
		return v.Items()
	case []Item:
		return append([]Item(nil), v...)
	default:
		return nil
	}
}
