package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Arkoes07/croissant/internal/quiz"
)

//go:embed templates
var templateFS embed.FS

// Server holds all dependencies for the HTTP layer.
type Server struct {
	quizSvc  quiz.Service
	tmpl     *parsedTemplates
	basePath string // e.g. "/croissant"; empty string means serve at root
}

// parsedTemplates holds pre-parsed template sets, one per rendered view.
type parsedTemplates struct {
	home     *template.Template
	question *template.Template
	result   *template.Template
	answer   *template.Template
}

// New creates a Server, parsing all templates eagerly so startup fails fast
// on any template syntax error. basePath is an optional URL prefix the app is
// mounted under (e.g. "/croissant"); pass "" to serve at the root.
func New(quizSvc quiz.Service, basePath string) (*Server, error) {
	tmpl, err := parseTemplates(templateFS)
	if err != nil {
		return nil, err
	}

	return &Server{
		quizSvc:  quizSvc,
		tmpl:     tmpl,
		basePath: basePath,
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
	// inner mux always uses root-relative paths so handlers stay simple.
	inner := http.NewServeMux()
	inner.HandleFunc("GET /{$}", s.handleHome)
	inner.HandleFunc("POST /quiz/new", s.handleNewQuiz)
	inner.HandleFunc("GET /quiz/{id}", s.handleQuestion)
	inner.HandleFunc("POST /quiz/{id}/answer", s.handleAnswer)
	inner.HandleFunc("GET /quiz/{id}/result", s.handleResult)

	if s.basePath == "" {
		return inner
	}

	// Mount inner under basePath, redirecting the bare path to the slash form.
	outer := http.NewServeMux()
	outer.HandleFunc("GET "+s.basePath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.basePath+"/", http.StatusMovedPermanently)
	})
	outer.Handle(s.basePath+"/", http.StripPrefix(s.basePath, inner))
	return outer
}
