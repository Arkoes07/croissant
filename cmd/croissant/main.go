package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/quiz/memorystore"
	"github.com/Arkoes07/croissant/internal/quiz/quizservice"
	"github.com/Arkoes07/croissant/internal/song"
	"github.com/Arkoes07/croissant/internal/song/jsonservice"
	"github.com/Arkoes07/croissant/internal/web"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// initialize song service from local JSON file
	// TODO: switch back to spotifyservice once API access is restored
	var songSvc song.Service
	{
		cfg := jsonservice.Config{
			FilePath:   filepath.Join(dir, "files", "songs", "songs.json"),
			SongsCount: 20,
		}
		songSvc = jsonservice.New(cfg)
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
		quizSvc = quizservice.New(songSvc, gen, store)
	}

	// initialize web server
	srv, err := web.New(quizSvc)
	if err != nil {
		log.Fatalf("[main] failed to init web server: %v\n", err)
	}

	addr := ":8080"
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("[main] server error: %v\n", err)
	}
}
