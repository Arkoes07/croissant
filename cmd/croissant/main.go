package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/quiz/memorystore"
	"github.com/Arkoes07/croissant/internal/quiz/quizservice"
	"github.com/Arkoes07/croissant/internal/song"
	"github.com/Arkoes07/croissant/internal/song/deezerservice"
	"github.com/Arkoes07/croissant/internal/web"
)

func main() {
	// initialize one Deezer song service per selectable playlist
	songSvcs := map[string]song.Service{
		"13650084141": deezerservice.New(deezerservice.Config{PlaylistID: "13650084141", SongsCount: 20}), // 2020s Hits
		"14917741483": deezerservice.New(deezerservice.Config{PlaylistID: "14917741483", SongsCount: 20}), // 2010s Hits
		"248297032":   deezerservice.New(deezerservice.Config{PlaylistID: "248297032", SongsCount: 20}),   // 2000s Hits
	}

	// initialize quiz generator (10 questions, 4 choices each)
	gen, err := quiz.NewGenerator(quiz.GeneratorConfig{
		QuestionCount: 10,
		ChoiceCount:   4,
	})
	if err != nil {
		log.Fatalf("[main] failed to init generator: %v\n", err)
	}

	// initialize quiz service (wires song fetching, generation, and persistence)
	var quizSvc quiz.Service
	{
		store := memorystore.New()
		quizSvc = quizservice.New(songSvcs, gen, store)
	}

	// initialize web server
	basePath := os.Getenv("BASE_PATH") // e.g. "/croissant"; empty for root
	srv, err := web.New(quizSvc, basePath)
	if err != nil {
		log.Fatalf("[main] failed to init web server: %v\n", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("[main] server error: %v\n", err)
	}
}
