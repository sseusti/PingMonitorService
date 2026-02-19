package jobs

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]Job),
	}
}

func (s *Store) Create(total int) Job {
	j := Job{
		ID:        NewID(),
		Status:    Running,
		CreatedAt: time.Now().UTC(),
		Total:     total,
		Done:      0,
	}

	s.mu.Lock()
	s.jobs[j.ID] = j
	s.mu.Unlock()

	return j
}

func (s *Store) Get(id string) (Job, bool) {
	s.mu.RLock()
	j, ok := s.jobs[id]
	s.mu.RUnlock()

	return j, ok
}

func (s *Store) SetStatus(id string, status Status) bool {
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		return false
	}

	j.Status = status
	s.jobs[id] = j
	s.mu.Unlock()

	return true
}

func (s *Store) SetRunning(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return false
	}

	j.Status = Running
	s.jobs[id] = j

	return true
}

func (s *Store) SetResults(id string, results []Result) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return false
	}

	j.Status = Done
	j.Done = len(results)
	j.Results = results
	t := time.Now().UTC()
	j.FinishedAt = &t
	s.jobs[id] = j

	return true
}

func (s *Store) SetFailed(id string, err error) bool {
	return s.SetFailedWithResults(id, err, nil)
}

func (s *Store) SetFailedWithResults(id string, err error, results []Result) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return false
	}

	j.Status = Failed
	j.Error = err.Error()
	if results != nil {
		j.Results = results
		j.Done = len(results)
	}
	t := time.Now().UTC()
	j.FinishedAt = &t
	s.jobs[id] = j

	return true
}

func NewID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}
