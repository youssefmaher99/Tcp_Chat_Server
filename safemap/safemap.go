package safemap

import (
	"sync"
)

type SafeSet[K comparable] struct {
	Smap map[K]struct{}
	mu   sync.Mutex
}

func (s *SafeSet[K]) Store(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Smap[k] = struct{}{}
}

func (s *SafeSet[K]) Load(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *SafeSet[K]) Delete(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Smap[k] = struct{}{}
}

// func (s *SafeMap[K, V]) Load(k K) {}
// func (s *SafeMap) Remove() {}
