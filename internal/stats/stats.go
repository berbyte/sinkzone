package stats

import (
	"sync"
	"time"
)

type DomainStats struct {
	Domain   string
	Count    int
	LastSeen time.Time
	Blocked  bool
}

type DNSStats struct {
	mu      sync.RWMutex
	domains map[string]*DomainStats
}

func NewDNSStats() *DNSStats {
	return &DNSStats{
		domains: make(map[string]*DomainStats),
	}
}

func (s *DNSStats) UpdateStats(domain string, blocked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stats, exists := s.domains[domain]; exists {
		stats.Count++
		stats.LastSeen = time.Now()
		stats.Blocked = blocked
	} else {
		s.domains[domain] = &DomainStats{
			Domain:   domain,
			Count:    1,
			LastSeen: time.Now(),
			Blocked:  blocked,
		}
	}
}

func (s *DNSStats) GetStats() map[string]*DomainStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the stats
	stats := make(map[string]*DomainStats)
	for domain, stat := range s.domains {
		stats[domain] = &DomainStats{
			Domain:   stat.Domain,
			Count:    stat.Count,
			LastSeen: stat.LastSeen,
			Blocked:  stat.Blocked,
		}
	}
	return stats
}
