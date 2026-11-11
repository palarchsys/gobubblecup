package framework

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RuntimeConfig est l'appel nommé loop, args, functions.
// StatusRegex est optionnelle. Par défaut la sortie doit contenir success, warning ou error.
// Dir vide : le bash part du dossier de lancement. Sinon c'est son répertoire de travail.
type RuntimeConfig struct {
	Loop        bool
	Timer       time.Duration
	Args        Store
	Functions   func(args Store)
	StatusRegex string
	Dir         string
}

// runtime exécute ses bash dans l'ordre.
// Status() est le cycle : not_initialized, pending, running, terminated.
// Le status d'un slot reste vide tant que son script n'a pas démarré,
// puis pending pendant qu'il tourne, puis success, warning ou error.
type runtime struct {
	Memory   *memory
	writable *memory
	argsMem  *memory
	re       *regexp.Regexp
	loop     bool
	timer    time.Duration
	dir      string
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
		Memory:   &memory{head: store.head, runtime: true},
		writable: store,
		argsMem:  args,
		re:       compiled,
		loop:     cfg.Loop,
		timer:    timer,
		dir:      cfg.Dir,
		fn:       cfg.Functions,
		phase:    "not_initialized",
		ctx:      ctx,
		cancel:   cancel,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		resume:   make(chan struct{}, 1),
	}
	rt.enlist()
	return rt
}

// Les runtimes créés restent ici. Ctrl+C les arrête tous, sauf si Bind a choisi une autre touche.
var runtimeList struct {
	mu    sync.Mutex
	items []*runtime
}

func (r *runtime) enlist() {
	runtimeList.mu.Lock()
	runtimeList.items = append(runtimeList.items, r)
	runtimeList.mu.Unlock()
}

// cancelRuntimes tue les bash encore en course et laisse les slots lisibles.
func cancelRuntimes() {
	runtimeList.mu.Lock()
	items := append([]*runtime(nil), runtimeList.items...)
	runtimeList.mu.Unlock()
	for _, rt := range items {
		rt.cancelNow()
	}
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

func NewRuntime(cfg RuntimeConfig) *runtime { return Runtime(cfg) }

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
		r.stopNow()
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
	go r.run()
}

// Cancel renvoie l'action du bouton. Le script en cours s'arrête quand elle part.
// Le slot encore pending passe à error, avec le return déjà reçu.
// Si le runtime n'est pas en course, l'action revient tout de suite.
func (r *runtime) Cancel() Action { return Call{Run: r.cancelNow} }

func (r *runtime) cancelNow() {
	r.mu.Lock()
	running := r.running
	r.mu.Unlock()
	if !running {
		return
	}
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
	<-r.done
	r.mu.Lock()
	r.running = false
	r.paused = false
	r.phase = "terminated"
	r.mu.Unlock()
}

func (r *runtime) run() {
	defer func() {
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
		close(r.done)
	}()
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
				if r.stopping() {
					return
				}
				if r.writable.Get(job.id).Get("status").String() == "error" {
					r.setPhase("terminated")
					return
				}
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

// Stop renvoie l'action du bouton. Elle vide status et return quand elle part.
func (r *runtime) Stop() Action { return Call{Run: r.stopNow} }

func (r *runtime) stopNow() {
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

// Pause renvoie l'action du bouton. Elle retient le runtime entre deux scripts.
func (r *runtime) Pause() Action { return Call{Run: r.pauseNow} }

func (r *runtime) pauseNow() {
	r.mu.Lock()
	r.paused = true
	r.mu.Unlock()
}

// Resume renvoie l'action du bouton. Sans pause, elle ne fait rien.
func (r *runtime) Resume() Action { return Call{Run: r.resumeNow} }

func (r *runtime) resumeNow() {
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
	env   map[string]string
	rt    *runtime
}

func Bash(id, label, path string, args []string) {
	trackBash(id, label, path, args, nil)
}

// BashEnv est un Bash dont les clés recouvrent l'environnement hérité.
// Une map nulle ou vide équivaut à Bash. Le même id remplace le job.
func BashEnv(id, label, path string, args []string, env map[string]string) {
	trackBash(id, label, path, args, env)
}

func trackBash(id, label, path string, args []string, env map[string]string) {
	rt := currentRuntime()
	if rt == nil {
		panic("framework: bash hors runtime")
	}
	rt.track(id, label, path, args, env)
}

func copyEnv(env map[string]string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	out := make(map[string]string, len(env))
	for key, value := range env {
		out[key] = value
	}
	return out
}

func (r *runtime) track(id, label, path string, args []string, env map[string]string) {
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
			job.env = copyEnv(env)
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
		env:   copyEnv(env),
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
	b.rt.writable.setField(b.id, "return", "")
	argv := append([]string{b.path}, b.args...)
	cmd := exec.CommandContext(b.rt.ctx, "bash", argv...)
	if b.rt.dir != "" {
		cmd.Dir = b.rt.dir
	}
	if len(b.env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range b.env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err == nil {
			return nil
		}
		return cmd.Process.Kill()
	}
	cmd.WaitDelay = 200 * time.Millisecond
	buf := &lineBuf{job: b}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Start(); err != nil {
		b.rt.writable.setField(b.id, "return", err.Error())
		b.finish("error")
		return
	}
	err := cmd.Wait()
	buf.flush()
	if b.rt.stopping() || b.rt.ctx.Err() != nil {
		b.finish("error")
		return
	}
	text := b.rt.writable.Get(b.id).Get("return").String()
	b.finish(classify(b.rt.re, text, err))
}

func (b *bashJob) finish(status string) {
	b.rt.writable.setField(b.id, "status", status)
	b.rt.addLog(b.id + ": " + status)
}

func (b *bashJob) appendReturn(text string) {
	cur := b.rt.writable.Get(b.id).Get("return").String()
	b.rt.writable.setField(b.id, "return", cur+text)
}

// lineBuf ajoute chaque ligne complète à return, dans l'ordre d'arrivée.
type lineBuf struct {
	mu  sync.Mutex
	buf []byte
	job *bashJob
}

func (w *lineBuf) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i+1])
		w.buf = append([]byte(nil), w.buf[i+1:]...)
		w.job.appendReturn(line)
	}
	return len(p), nil
}

func (w *lineBuf) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) == 0 {
		return
	}
	w.job.appendReturn(string(w.buf))
	w.buf = nil
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
