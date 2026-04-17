package web

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/song"
)

// newID generates a random 8-byte hex ID for a quiz session.
func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// isHTMX reports whether the request was made by HTMX.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.home.ExecuteTemplate(w, "layout", nil); err != nil {
		slog.Error("render home", "err", err)
	}
}

func (s *Server) handleNewQuiz(w http.ResponseWriter, r *http.Request) {
	songs, err := s.songSvc.GetSongs()
	if err != nil {
		// ErrCountMismatch means fewer preview-URL tracks than requested — proceed
		// if we have at least some songs; the generator will validate the count.
		slog.Warn("GetSongs returned non-nil error", "err", err, "count", len(songs))
		if len(songs) == 0 {
			http.Error(w, "failed to load songs from Spotify", http.StatusInternalServerError)
			return
		}
	}

	questions, err := s.generator.Generate(songs)
	if err != nil {
		slog.Error("generate questions", "err", err)
		http.Error(w, "not enough songs to build a quiz — try a bigger playlist", http.StatusInternalServerError)
		return
	}

	q := quiz.Quiz{
		ID:        newID(),
		Questions: questions,
		StartedAt: time.Now(),
	}
	if err := s.store.Save(q); err != nil {
		slog.Error("save quiz", "err", err)
		http.Error(w, "failed to save quiz session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/quiz/"+q.ID, http.StatusSeeOther)
}

// questionData is the template data for the question page.
type questionData struct {
	QuizID         string
	QuestionNum    int
	TotalQuestions int
	AudioURL       string
	Choices        []song.Song
	Score          int
}

func (s *Server) handleQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	q, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "quiz not found", http.StatusNotFound)
		return
	}

	if q.IsDone() {
		http.Redirect(w, r, "/quiz/"+id+"/result", http.StatusSeeOther)
		return
	}

	current := q.Questions[q.CurrentIdx]
	data := questionData{
		QuizID:         id,
		QuestionNum:    q.CurrentIdx + 1,
		TotalQuestions: len(q.Questions),
		AudioURL:       current.AudioURL,
		Choices:        current.Choices,
		Score:          q.Score,
	}

	if err := s.tmpl.question.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("render question", "err", err)
	}
}

// answerData is the template data for the answer feedback fragment.
type answerData struct {
	QuizID      string
	Correct     bool
	CorrectSong song.Song
	Score       int
	Answered    int
	IsDone      bool
}

func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	q, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "quiz not found", http.StatusNotFound)
		return
	}

	if q.IsDone() {
		http.Redirect(w, r, "/quiz/"+id+"/result", http.StatusSeeOther)
		return
	}

	// Capture the correct song before Answer() advances CurrentIdx.
	current := q.Questions[q.CurrentIdx]
	correctSong := current.Choices[current.CorrectIdx]

	choiceIdx, err := strconv.Atoi(r.FormValue("choiceIdx"))
	if err != nil {
		http.Error(w, "invalid choice", http.StatusBadRequest)
		return
	}

	correct, err := q.Answer(choiceIdx)
	if err != nil {
		http.Error(w, "could not record answer: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.Save(q); err != nil {
		slog.Error("save quiz after answer", "err", err)
		http.Error(w, "failed to save quiz session", http.StatusInternalServerError)
		return
	}

	// HTMX request → return the answer feedback fragment for in-place swap.
	if isHTMX(r) {
		data := answerData{
			QuizID:      id,
			Correct:     correct,
			CorrectSong: correctSong,
			Score:       q.Score,
			Answered:    q.CurrentIdx, // already advanced by Answer()
			IsDone:      q.IsDone(),
		}
		if err := s.tmpl.answer.ExecuteTemplate(w, "answer", data); err != nil {
			slog.Error("render answer fragment", "err", err)
		}
		return
	}

	// Non-HTMX fallback → redirect (no answer feedback shown).
	if q.IsDone() {
		http.Redirect(w, r, "/quiz/"+id+"/result", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/quiz/"+id, http.StatusSeeOther)
}

// resultData is the template data for the result page.
type resultData struct {
	Score   int
	Total   int
	Message string
}

func (s *Server) handleResult(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	q, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "quiz not found", http.StatusNotFound)
		return
	}

	total := len(q.Questions)
	data := resultData{
		Score:   q.Score,
		Total:   total,
		Message: scoreMessage(q.Score, total),
	}

	if err := s.tmpl.result.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("render result", "err", err)
	}
}

// scoreMessage returns a flavour message based on the player's performance.
func scoreMessage(score, total int) string {
	if total == 0 {
		return ""
	}
	pct := float64(score) / float64(total)
	switch {
	case pct == 1.0:
		return "Perfect score! You really know your music."
	case pct >= 0.7:
		return "Great job! You know your tunes."
	case pct >= 0.4:
		return "Not bad! Keep listening."
	default:
		return "Time to expand that playlist."
	}
}
