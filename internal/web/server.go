package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/song"
)

//go:embed templates
var templateFS embed.FS

// Server holds all dependencies for the HTTP layer.
type Server struct {
	songSvc   song.Service
	generator *quiz.Generator
	store     quiz.Store
	tmpl      *parsedTemplates
}

// parsedTemplates holds pre-parsed template sets, one per rendered view.
type parsedTemplates struct {
	home     *template.Template
	question *template.Template
	result   *template.Template
	answer   *template.Template
}

// New creates a Server, parsing all templates eagerly so startup fails fast
// on any template syntax error.
func New(songSvc song.Service, gen *quiz.Generator, store quiz.Store) (*Server, error) {
	tmpl, err := parseTemplates(templateFS)
	if err != nil {
		return nil, err
	}

	return &Server{
		songSvc:   songSvc,
		generator: gen,
		store:     store,
		tmpl:      tmpl,
	}, nil
}

// parseTemplates parses each page as its own template set so that each set has
// exactly one "content" block definition paired with the shared layout.
func parseTemplates(fsys fs.FS) (*parsedTemplates, error) {
	parse := func(files ...string) (*template.Template, error) {
		return template.ParseFS(fsys, files...)
	}

	home, err := parse("templates/layout.html", "templates/home.html")
	if err != nil {
		return nil, err
	}

	question, err := parse("templates/layout.html", "templates/question.html")
	if err != nil {
		return nil, err
	}

	result, err := parse("templates/layout.html", "templates/result.html")
	if err != nil {
		return nil, err
	}

	// answer.html is a standalone fragment — no layout wrapper.
	answer, err := parse("templates/answer.html")
	if err != nil {
		return nil, err
	}

	return &parsedTemplates{
		home:     home,
		question: question,
		result:   result,
		answer:   answer,
	}, nil
}

// Handler builds and returns the HTTP router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// /{$} matches only the root path; "GET /" would be a catch-all.
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("POST /quiz/new", s.handleNewQuiz)
	mux.HandleFunc("GET /quiz/{id}", s.handleQuestion)
	mux.HandleFunc("POST /quiz/{id}/answer", s.handleAnswer)
	mux.HandleFunc("GET /quiz/{id}/result", s.handleResult)

	return mux
}
