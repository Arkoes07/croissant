package memorystore

import (
	"sync"

	"github.com/Arkoes07/croissant/internal/quiz"
)

// store is an in-memory implementation of quiz.Store.
// Safe for concurrent use.
type store struct {
	mu     sync.RWMutex
	quizzes map[string]quiz.Quiz
}

// New creates a new in-memory store.
func New() quiz.Store {
	return &store{
		quizzes: make(map[string]quiz.Quiz),
	}
}

// Save persists or overwrites a quiz by its ID.
func (s *store) Save(q quiz.Quiz) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quizzes[q.ID] = q
	return nil
}

// Get retrieves a quiz by ID. Returns quiz.ErrNotFound if absent.
func (s *store) Get(id string) (quiz.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q, ok := s.quizzes[id]
	if !ok {
		return quiz.Quiz{}, quiz.ErrNotFound
	}
	return q, nil
}
