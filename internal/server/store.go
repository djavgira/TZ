package server

import (
	"sort"
	"sync"
	"time"

	tzv1 "github.com/djavgira/TZ/pkg/proto/tz/v1"
)

// HostSnapshot holds the latest metrics received from a single agent.
type HostSnapshot struct {
	HostID   string
	LastSeen time.Time
	Report   *tzv1.MetricReport
}

// Store is a thread-safe in-memory store of the latest metrics per host.
type Store struct {
	mu    sync.RWMutex
	hosts map[string]*HostSnapshot
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{
		hosts: make(map[string]*HostSnapshot),
	}
}

// Update merges a received MetricReport into the store.
func (s *Store) Update(report *tzv1.MetricReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.hosts[report.HostId] = &HostSnapshot{
		HostID:   report.HostId,
		LastSeen: time.Now(),
		Report:   report,
	}
}

// GetAll returns all host snapshots sorted by HostID.
func (s *Store) GetAll() []*HostSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*HostSnapshot, 0, len(s.hosts))
	for _, h := range s.hosts {
		result = append(result, h)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].HostID < result[j].HostID
	})
	return result
}

// Get returns a single host snapshot, or nil if not found.
func (s *Store) Get(hostID string) *HostSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hosts[hostID]
}

// HostCount returns the number of currently tracked hosts.
func (s *Store) HostCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.hosts)
}

// PruneStale removes hosts that haven't been updated within maxAge.
func (s *Store) PruneStale(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for id, h := range s.hosts {
		if h.LastSeen.Before(cutoff) {
			delete(s.hosts, id)
			removed++
		}
	}
	return removed
}
