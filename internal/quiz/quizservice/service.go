package quizservice

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/song"
)

// service is the adapter that implements the quiz port.
type service struct {
	songSvc   song.Service
	generator *quiz.Generator
	store     quiz.Store
}

// New creates a quiz.Service that orchestrates song fetching, question
// generation, and quiz persistence.
func New(songSvc song.Service, gen *quiz.Generator, store quiz.Store) quiz.Service {
	return &service{
		songSvc:   songSvc,
		generator: gen,
		store:     store,
	}
}

// newID generates a random 8-byte hex string for quiz session IDs.
func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// NewQuiz fetches songs, generates questions, persists the quiz, and returns it.
func (s *service) NewQuiz() (quiz.Quiz, error) {
	songs, err := s.songSvc.GetSongs()
	if err != nil {
		// ErrCountMismatch means fewer preview-URL tracks than requested —
		// proceed if we still have enough songs for the generator.
		slog.Warn("GetSongs returned non-nil error", "err", err, "count", len(songs))
		if len(songs) == 0 {
			return quiz.Quiz{}, err
		}
	}

	questions, err := s.generator.Generate(songs)
	if err != nil {
		return quiz.Quiz{}, err
	}

	q := quiz.Quiz{
		ID:        newID(),
		Questions: questions,
		StartedAt: time.Now(),
	}
	if err := s.store.Save(q); err != nil {
		return quiz.Quiz{}, err
	}

	return q, nil
}

// GetQuiz retrieves an active quiz by ID.
func (s *service) GetQuiz(id string) (quiz.Quiz, error) {
	return s.store.Get(id)
}

// Answer records the player's choice for the current question, persists the
// updated quiz, and returns a result summary.
func (s *service) Answer(id string, choiceIdx int) (quiz.AnswerResult, error) {
	q, err := s.store.Get(id)
	if err != nil {
		return quiz.AnswerResult{}, err
	}

	if q.IsDone() {
		return quiz.AnswerResult{}, quiz.ErrAlreadyDone
	}

	// Capture the correct song before Answer() advances CurrentIdx.
	current := q.Questions[q.CurrentIdx]
	correctSong := current.Choices[current.CorrectIdx]

	correct, err := q.Answer(choiceIdx)
	if err != nil {
		return quiz.AnswerResult{}, err
	}

	if err := s.store.Save(q); err != nil {
		return quiz.AnswerResult{}, err
	}

	return quiz.AnswerResult{
		Correct:     correct,
		CorrectSong: correctSong,
		Score:       q.Score,
		Answered:    q.CurrentIdx,
		IsDone:      q.IsDone(),
	}, nil
}
