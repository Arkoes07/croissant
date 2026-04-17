package quiz

import (
	"errors"
	"time"

	"github.com/Arkoes07/croissant/internal/song"
)

// below are the known Quiz errors
var (
	ErrNotFound    = errors.New("quiz not found")
	ErrOutOfBounds = errors.New("question index out of bounds")
	ErrAlreadyDone = errors.New("quiz already finished")
)

// Question represents a single round in the quiz.
// The player hears AudioURL and must pick the correct song from Choices.
type Question struct {
	// Choices contains the correct song and distractors in shuffled order.
	Choices []song.Song
	// CorrectIdx is the index of the correct song within Choices.
	CorrectIdx int
	// AudioURL is the 30-second preview URL to play.
	AudioURL string
}

// Quiz represents an active game session.
type Quiz struct {
	// ID uniquely identifies this quiz session.
	ID string
	// Questions holds all rounds for this session.
	Questions []Question
	// CurrentIdx is the index of the question currently being answered.
	CurrentIdx int
	// Score is the number of correct answers so far.
	Score int
	// StartedAt records when the quiz was created.
	StartedAt time.Time
}

// IsDone reports whether all questions have been answered.
func (q *Quiz) IsDone() bool {
	return q.CurrentIdx >= len(q.Questions)
}

// Answer records the player's guess for the current question.
// It returns true if the guess was correct and advances to the next question.
// Returns ErrAlreadyDone if the quiz is finished, ErrOutOfBounds if idx is invalid.
func (q *Quiz) Answer(choiceIdx int) (correct bool, err error) {
	if q.IsDone() {
		return false, ErrAlreadyDone
	}

	current := q.Questions[q.CurrentIdx]
	if choiceIdx < 0 || choiceIdx >= len(current.Choices) {
		return false, ErrOutOfBounds
	}

	correct = choiceIdx == current.CorrectIdx
	if correct {
		q.Score++
	}
	q.CurrentIdx++

	return correct, nil
}

// Store defines persistence behaviour for quiz sessions.
type Store interface {
	// Save persists or updates a quiz.
	Save(quiz Quiz) error
	// Get retrieves a quiz by ID. Returns ErrNotFound if absent.
	Get(id string) (Quiz, error)
}

// AnswerResult carries everything the caller needs after an answer is recorded.
type AnswerResult struct {
	// Correct reports whether the player's guess was right.
	Correct bool
	// CorrectSong is the song that was the right answer.
	CorrectSong song.Song
	// Score is the updated total number of correct answers.
	Score int
	// Answered is the number of questions answered so far (after this one).
	Answered int
	// IsDone reports whether the quiz is now finished.
	IsDone bool
}

// Service is the port for the quiz domain. It orchestrates song fetching,
// question generation, persistence, and answer recording behind a single
// interface so callers (e.g. the HTTP layer) need only one dependency.
type Service interface {
	// NewQuiz fetches songs from the given playlist, generates questions,
	// persists the quiz, and returns it ready for the first question.
	NewQuiz(playlistID string) (Quiz, error)
	// GetQuiz retrieves an active quiz by ID. Returns ErrNotFound if absent.
	GetQuiz(id string) (Quiz, error)
	// Answer records the player's choice for the current question, persists
	// the updated quiz, and returns a result summary.
	Answer(id string, choiceIdx int) (AnswerResult, error)
}
