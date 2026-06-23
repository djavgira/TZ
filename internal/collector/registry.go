package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/djavgira/TZ/pkg/metrics"
)

// SnapshotHandler is called after a successful collection cycle with the
// collector name and the populated snapshot. The snapshot may only have
// the subset of metrics that the specific collector fills.
type SnapshotHandler func(name string, snap *metrics.Snapshot)

type collectorEntry struct {
	collector Collector
	interval  time.Duration
	cancel    context.CancelFunc
}

// Registry manages the set of active collectors.
type Registry struct {
	mu         sync.RWMutex
	collectors map[string]*collectorEntry

	// Tracks ready state for /ready endpoint.
	muReady sync.RWMutex
	ready   map[string]bool

	// Called after every successful collection.
	onSnapshot SnapshotHandler
}

// NewRegistry creates an empty collector registry.
func NewRegistry(onSnapshot SnapshotHandler) *Registry {
	return &Registry{
		collectors:  make(map[string]*collectorEntry),
		ready:       make(map[string]bool),
		onSnapshot:  onSnapshot,
	}
}

// Register adds a collector. Returns an error if a collector with the same
// Name() is already registered.
func (r *Registry) Register(c Collector, interval time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := c.Name()
	if _, exists := r.collectors[name]; exists {
		return fmt.Errorf("collector %q already registered", name)
	}

	r.collectors[name] = &collectorEntry{
		collector: c,
		interval:  interval,
	}
	r.ready[name] = false

	return nil
}

// StartAll launches every registered collector in its own goroutine.
func (r *Registry) StartAll(parentCtx context.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, entry := range r.collectors {
		ctx, cancel := context.WithCancel(parentCtx)
		entry.cancel = cancel

		go r.runCollector(ctx, name, entry)
	}
}

// runCollector is the per-collector goroutine.
func (r *Registry) runCollector(ctx context.Context, name string, entry *collectorEntry) {
	// CPU is special: cpu.Percent blocks for the interval.
	if name == "cpu" {
		r.runBlockingCollector(ctx, name, entry)
		return
	}

	ticker := time.NewTicker(entry.interval)
	defer ticker.Stop()

	// Run first collection immediately
	r.runCollectionCycle(ctx, name, entry)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runCollectionCycle(ctx, name, entry)
		}
	}
}

// runBlockingCollector handles collectors where Collect() blocks for the
// sampling duration (e.g., CPU).
func (r *Registry) runBlockingCollector(ctx context.Context, name string, entry *collectorEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.runCollectionCycle(ctx, name, entry)
		}
	}
}

// runCollectionCycle executes one Collect() call and handles errors.
// After a successful collection, the snapshot is passed to r.onSnapshot.
func (r *Registry) runCollectionCycle(ctx context.Context, name string, entry *collectorEntry) {
	defer func() {
		if rc := recover(); rc != nil {
			// Panic recovery prevents a single collector crash from
			// taking down the entire agent.
		}
	}()

	snap := &metrics.Snapshot{}
	if err := entry.collector.Collect(ctx, snap); err != nil {
		return
	}

	// Mark as ready after first successful collection
	r.muReady.Lock()
	r.ready[name] = true
	r.muReady.Unlock()

	// Notify the handler
	if r.onSnapshot != nil {
		r.onSnapshot(name, snap)
	}
}

// StopAll cancels all collector goroutines.
func (r *Registry) StopAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entry := range r.collectors {
		if entry.cancel != nil {
			entry.cancel()
		}
	}
}

// Statuses returns a map of collector name -> health status.
func (r *Registry) Statuses() map[string]error {
	r.muReady.RLock()
	defer r.muReady.RUnlock()

	statuses := make(map[string]error)
	for name, ready := range r.ready {
		if !ready {
			statuses[name] = fmt.Errorf("collector %q not ready", name)
		}
	}
	return statuses
}

// AllReady returns true if all registered collectors have completed at
// least one successful collection cycle.
func (r *Registry) AllReady() bool {
	r.muReady.RLock()
	defer r.muReady.RUnlock()

	for _, ready := range r.ready {
		if !ready {
			return false
		}
	}
	return len(r.ready) > 0
}

// IsReady implements server.ReadyChecker.
func (r *Registry) IsReady() bool { return r.AllReady() }

// HealthStatuses implements server.HealthChecker.
func (r *Registry) HealthStatuses() map[string]error { return r.Statuses() }

// CollectAll calls Collect() on every registered collector synchronously,
// merging results into a single Snapshot. Used by agent mode for gRPC push.
func (r *Registry) CollectAll(ctx context.Context) (*metrics.Snapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snap := &metrics.Snapshot{}
	for name, entry := range r.collectors {
		if err := entry.collector.Collect(ctx, snap); err != nil {
			return nil, fmt.Errorf("collector %q: %w", name, err)
		}
		// Mark ready after each successful collection
		r.muReady.Lock()
		r.ready[name] = true
		r.muReady.Unlock()
	}
	return snap, nil
}

// Count returns the number of registered collectors.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.collectors)
}
