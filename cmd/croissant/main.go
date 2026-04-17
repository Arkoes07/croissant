package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Arkoes07/croissant/internal/quiz"
	"github.com/Arkoes07/croissant/internal/quiz/memorystore"
	"github.com/Arkoes07/croissant/internal/secret"
	"github.com/Arkoes07/croissant/internal/secret/jsonsecret"
	"github.com/Arkoes07/croissant/internal/song"
	"github.com/Arkoes07/croissant/internal/song/spotifyservice"
	"github.com/Arkoes07/croissant/internal/web"
)

func main() {
	// generate path to secret directory
	var secretDir string
	{
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		secretDir = filepath.Join(dir, "files", "secret")
	}

	// initialize secret service
	var secretSvc secret.Service
	{
		cfg := jsonsecret.Config{
			FilePath: filepath.Join(secretDir, "secret.json"),
		}
		secretSvc = jsonsecret.New(cfg)
	}

	// parse Spotify credentials
	sec, err := secretSvc.Parse()
	if err != nil {
		log.Fatalf("[main] failed to parse secret: %v\n", err)
	}

	// initialize song service
	// TODO: move config values to a config file
	var songSvc song.Service
	{
		cfg := spotifyservice.Config{
			PlaylistID: "37i9dQZF1DXcBWIGoYBM5M",
			SongsCount: 20,
		}
		songSvc, err = spotifyservice.New(sec.Spotify.ClientID, sec.Spotify.ClientSecret, cfg)
		if err != nil {
			log.Fatalf("[main] failed to init song service: %v\n", err)
		}
	}

	// initialize quiz generator (10 questions, 4 choices each)
	gen, err := quiz.NewGenerator(quiz.GeneratorConfig{
		QuestionCount: 10,
		ChoiceCount:   4,
	})
	if err != nil {
		log.Fatalf("[main] failed to init generator: %v\n", err)
	}

	// initialize in-memory quiz store
	store := memorystore.New()

	// initialize web server
	srv, err := web.New(songSvc, gen, store)
	if err != nil {
		log.Fatalf("[main] failed to init web server: %v\n", err)
	}

	addr := ":8080"
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("[main] server error: %v\n", err)
	}
}
