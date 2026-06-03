package project

import (
	"context"
	"sync"
	"time"
)

// ProjectSupplier exposes a function that returns the current project to
// persist. It is satisfied by *app.App without creating an import cycle.
type ProjectSupplier func() *Project

// AutoSaver periodically persists the active project. The actual project
// pointer is fetched on every tick via the supplied getter so the saver
// always saves the latest edits without holding a stale reference.
type AutoSaver struct {
	manager  *Manager
	getter   ProjectSupplier
	interval time.Duration

	mu      sync.Mutex
	enabled bool
	stopCh  chan struct{}
}

func NewAutoSaver(m *Manager, intervalSec int) *AutoSaver {
	return &AutoSaver{
		manager:  m,
		interval: time.Duration(intervalSec) * time.Second,
		enabled:  true,
	}
}

// SetProjectSupplier wires the function used to fetch the live project.
// Must be called before Start (or any time before a tick fires).
func (a *AutoSaver) SetProjectSupplier(g ProjectSupplier) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.getter = g
}

func (a *AutoSaver) SetEnabled(v bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = v
}

func (a *AutoSaver) Start(ctx context.Context) {
	a.mu.Lock()
	if a.stopCh != nil {
		a.mu.Unlock()
		return
	}
	a.stopCh = make(chan struct{})
	stopCh := a.stopCh
	a.mu.Unlock()

	go a.loop(ctx, stopCh)
}

func (a *AutoSaver) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.stopCh != nil {
		close(a.stopCh)
		a.stopCh = nil
	}
}

func (a *AutoSaver) loop(ctx context.Context, stopCh chan struct{}) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-stopCh:
			return
		case <-ticker.C:
			a.tick()
		}
	}
}

func (a *AutoSaver) tick() {
	a.mu.Lock()
	enabled := a.enabled
	getter := a.getter
	a.mu.Unlock()

	if !enabled || getter == nil || a.manager == nil {
		return
	}

	p := getter()
	if p == nil {
		return
	}

	if err := a.manager.SaveProject(*p); err != nil {
		// Best-effort: log via the Wails event system if available.
		// The frontend listens to "project:autosaved" but not to errors;
		// surfacing a toast there can be added in v1.0.
		return
	}
}
