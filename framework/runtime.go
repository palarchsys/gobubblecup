package framework

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// RuntimeConfig est l'appel nommé loop, args, functions.
// StatusRegex est optionnelle. Par défaut la sortie doit contenir success, warning ou error.
type RuntimeConfig struct {
	Loop        bool
	Timer       time.Duration
	Args        Store
	Functions   func(args Store)
	StatusRegex string
}

// runtime exécute ses bash dans l'ordre.
// Status() est le cycle : not_initialized, pending, running, terminated.
// Le status d'un slot est vide avant Start, puis pending, puis success, warning ou error.
type runtime struct {
	Memory   *memory
	writable *memory
	argsMem  *memory
	re       *regexp.Regexp
	loop     bool
	timer    time.Duration
	fn       func(args Store)
	mu       sync.Mutex
	phase    string
	log      string
	jobs     []*bashJob
	ctx      context.Context
	cancel   context.CancelFunc
	stop     chan struct{}
	done     chan struct{}
	resume   chan struct{}
	once     sync.Once
	running  bool
	paused   bool
}

func Runtime(cfg RuntimeConfig) *runtime {
	pattern := cfg.StatusRegex
	if strings.TrimSpace(pattern) == "" {
		pattern = `(success|warning|error)`
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		panic("framework: status_regex invalide")
	}
	timer := cfg.Timer
	if cfg.Loop && timer <= 0 {
		timer = 100 * time.Millisecond
	}
	store := Memory()
	args := concreteMemory(cfg.Args)
	ctx, cancel := context.WithCancel(context.Background())
	rt := &runtime{
		Memory:   store.ReadOnly(),
		writable: store,
		argsMem:  args,
		re:       compiled,
		loop:     cfg.Loop,
		timer:    timer,
		fn:       cfg.Functions,
		phase:    "not_initialized",
		ctx:      ctx,
		cancel:   cancel,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		resume:   make(chan struct{}, 1),
	}
	return rt
}

// Init enregistre les bash et remplit leur slot. Aucun script n'est lancé.
func (r *runtime) Init() {
	r.invoke()
}

func concreteMemory(store Store) *memory {
	if store == nil {
		return Memory()
	}
	memo, ok := store.(*memory)
	if !ok {
		panic("framework: mémoire attendue")
	}
	return memo
}

func NewRuntime(loop bool, args []string, statusRegex string) *runtime {
	memo := Memory()
	for i, arg := range args {
		memo.Set(fmt.Sprint(i+1), arg)
	}
	return Runtime(RuntimeConfig{Loop: loop, Args: memo, StatusRegex: statusRegex})
}

func (r *runtime) Status() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.phase
}

func (r *runtime) Log() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.log
}

func (r *runtime) Args() []string {
	out := make([]string, 0)
	for _, entry := range r.argsMem.Entries() {
		out = append(out, entry.Fields["value"])
	}
	return out
}

func (r *runtime) setPhase(value string) {
	r.mu.Lock()
	r.phase = value
	r.mu.Unlock()
}

func (r *runtime) addLog(line string) {
	r.mu.Lock()
	if r.log != "" && !strings.HasSuffix(r.log, "\n") {
		r.log += "\n"
	}
	r.log += line
	if !strings.HasSuffix(r.log, "\n") {
		r.log += "\n"
	}
	r.mu.Unlock()
	notify()
}

func (r *runtime) stopping() bool {
	select {
	case <-r.stop:
		return true
	default:
		return false
	}
}

func (r *runtime) guard(fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			r.setPhase("terminated")
			r.addLog(fmt.Sprint(rec))
		}
	}()
	fn()
}

func (r *runtime) invoke() {
	if r.fn == nil {
		return
	}
	withRuntime(r, func() { r.fn(r.argsMem) })
}

// Start renvoie l'action du bouton. Le lancement a lieu quand cette action est déclenchée.
func (r *runtime) Start() Action {
	return Call{Run: r.launch}
}

func (r *runtime) launch() {
	r.mu.Lock()
	running := r.running
	r.mu.Unlock()
	if running {
		r.Stop()
	}
	r.mu.Lock()
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.stop = make(chan struct{})
	r.done = make(chan struct{})
	r.once = sync.Once{}
	r.phase = "pending"
	r.running = true
	r.mu.Unlock()
	r.invoke()
	for _, job := range r.snapshot() {
		r.writable.setField(job.id, "status", "pending")
		r.writable.setField(job.id, "return", "")
	}
	go r.run()
}

func (r *runtime) run() {
	defer func() { close(r.done) }()
	r.guard(func() {
		for {
			if r.stopping() {
				return
			}
			jobs := r.snapshot()
			loop := r.loop
			if len(jobs) == 0 {
				r.setPhase("terminated")
				return
			}
			for _, job := range jobs {
				if r.stopping() {
					return
				}
				r.waitPause()
				if r.stopping() {
					return
				}
				r.setPhase("running")
				job.execute()
			}
			if !loop || r.stopping() {
				r.setPhase("terminated")
				return
			}
			r.setPhase("pending")
			r.sleepTimer()
			if r.stopping() {
				return
			}
			r.invoke()
		}
	})
}

func (r *runtime) snapshot() []*bashJob {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*bashJob(nil), r.jobs...)
}

func (r *runtime) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.once.Do(func() {
		select {
		case <-r.stop:
		default:
			close(r.stop)
		}
	})
	r.mu.Lock()
	running := r.running
	r.mu.Unlock()
	if running {
		<-r.done
	}
	r.mu.Lock()
	r.running = false
	r.paused = false
	r.phase = "not_initialized"
	jobs := append([]*bashJob(nil), r.jobs...)
	r.mu.Unlock()
	for _, job := range jobs {
		r.writable.setField(job.id, "status", "")
		r.writable.setField(job.id, "return", "")
	}
}

func (r *runtime) Pause() {
	r.mu.Lock()
	r.paused = true
	r.mu.Unlock()
}

func (r *runtime) Resume() {
	r.mu.Lock()
	was := r.paused
	r.paused = false
	r.mu.Unlock()
	if !was {
		return
	}
	select {
	case r.resume <- struct{}{}:
	default:
	}
}

func (r *runtime) sleepTimer() {
	delay := r.timer
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}
	clock := time.NewTimer(delay)
	defer clock.Stop()
	select {
	case <-clock.C:
	case <-r.stop:
	}
}

func (r *runtime) waitPause() {
	for {
		r.mu.Lock()
		paused := r.paused
		r.mu.Unlock()
		if !paused || r.stopping() {
			return
		}
		select {
		case <-r.resume:
			return
		case <-r.stop:
			return
		case <-time.After(15 * time.Millisecond):
		}
	}
}

type bashJob struct {
	id    string
	label string
	path  string
	args  []string
	rt    *runtime
}

func Bash(id, label, path string, args []string) {
	rt := currentRuntime()
	if rt == nil {
		panic("framework: bash hors runtime")
	}
	rt.track(id, label, path, args)
}

func (r *runtime) track(id, label, path string, args []string) {
	if id == "" {
		panic("framework: bash sans id")
	}
	if path == "" {
		panic("framework: bash sans chemin")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, job := range r.jobs {
		if job.id == id {
			job.label = label
			job.path = path
			job.args = append([]string(nil), args...)
			r.writable.setField(id, "label", label)
			r.writable.setField(id, "path", path)
			return
		}
	}
	r.jobs = append(r.jobs, &bashJob{
		id:    id,
		label: label,
		path:  path,
		args:  append([]string(nil), args...),
		rt:    r,
	})
	r.writable.setField(id, "id", id)
	r.writable.setField(id, "label", label)
	r.writable.setField(id, "path", path)
	if !r.writable.hasField(id, "status") {
		r.writable.setField(id, "status", "")
	}
	if !r.writable.hasField(id, "return") {
		r.writable.setField(id, "return", "")
	}
}

func (m *memory) hasField(id, field string) bool {
	if m == nil || m.head == nil || m.head.root == nil {
		return false
	}
	m.head.mu.RLock()
	defer m.head.mu.RUnlock()
	entry := m.head.root.kids[id]
	if entry == nil || entry.kids[field] == nil {
		return false
	}
	return entry.kids[field].has
}

func (b *bashJob) execute() {
	if b.rt.stopping() {
		return
	}
	b.rt.writable.setField(b.id, "status", "pending")
	argv := append([]string{b.path}, b.args...)
	cmd := exec.CommandContext(b.rt.ctx, "bash", argv...)
	out, err := cmd.CombinedOutput()
	if b.rt.stopping() && len(out) == 0 {
		return
	}
	text := string(out)
	b.rt.writable.setField(b.id, "return", text)
	status := classify(b.rt.re, text, err)
	b.rt.writable.setField(b.id, "status", status)
	b.rt.addLog(b.id + ": " + status)
}

func classify(re *regexp.Regexp, output string, err error) string {
	if re != nil {
		found := re.FindStringSubmatch(output)
		if found == nil {
			return "error"
		}
		word := found[0]
		if len(found) > 1 && found[1] != "" {
			word = found[1]
		}
		switch strings.ToLower(strings.TrimSpace(word)) {
		case "success", "warning", "error":
			return strings.ToLower(strings.TrimSpace(word))
		default:
			return "error"
		}
	}
	if err != nil {
		return "error"
	}
	return "success"
}

var runtimeSlot struct {
	mu sync.Mutex
	rt *runtime
}

func withRuntime(rt *runtime, fn func()) {
	runtimeSlot.mu.Lock()
	prev := runtimeSlot.rt
	runtimeSlot.rt = rt
	runtimeSlot.mu.Unlock()
	defer func() {
		runtimeSlot.mu.Lock()
		runtimeSlot.rt = prev
		runtimeSlot.mu.Unlock()
	}()
	fn()
}

func currentRuntime() *runtime {
	runtimeSlot.mu.Lock()
	defer runtimeSlot.mu.Unlock()
	return runtimeSlot.rt
}
