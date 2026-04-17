# CLAUDE.md

Context for future Claude Code sessions working in this repo.

## What this project is

**Croissant** is a song-guessing quiz game. The server fetches songs (with 30-second
preview clips) from Deezer playlists via the Deezer public API (no credentials required),
then asks the user to guess each song's title from its audio preview.

The game is live at [croissant.alwisukra.com](https://croissant.alwisukra.com), self-hosted
on a Coolify instance using the Dockerfile build pack.

## Tech stack

- **Language:** Go 1.24
- **Deezer API:** public REST API, no auth needed; `GET /playlist/{id}/tracks`
- **HTTP / UI:** stdlib `net/http` + `html/template` + HTMX, with a custom dark-themed audio player
- **Config:** environment variables only (`PORT`, no credentials needed)

## Architecture

Hexagonal (ports & adapters). Each domain package defines a `Service` interface
(the port); subpackages provide concrete adapters.

### Current layout

```
croissant/
├── cmd/croissant/main.go              ← entrypoint, wires services together
├── Dockerfile                         ← multi-stage build (golang:1.24-alpine → alpine)
└── internal/
    ├── song/
    │   ├── song.go                    ← Song{Title, Artists, PreviewURL} + Service port
    │   │                                Service interface: GetSongs(playlistID string) ([]Song, error)
    │   └── deezerservice/service.go   ← Deezer adapter; fetches playlist, filters tracks with no preview
    ├── quiz/
    │   ├── quiz.go                    ← Quiz, Question, Generator, Store port, Service port
    │   ├── generator.go               ← builds questions by sampling distractors from song pool
    │   ├── memorystore/store.go       ← in-memory Store adapter
    │   └── quizservice/service.go     ← Quiz Service adapter; orchestrates song fetch + generation + persistence
    └── web/
        ├── server.go                  ← Server struct, template parsing, HTTP router
        ├── handlers.go                ← handleHome, handleNewQuiz, handleQuestion, handleAnswer, handleResult
        └── templates/
            ├── layout.html            ← shared HTML shell + all CSS
            ├── home.html              ← era selection (3 Deezer playlist cards)
            ├── question.html          ← custom audio player + 4-choice grid
            ├── answer.html            ← HTMX fragment: correct/wrong feedback
            └── result.html            ← final score + play again
```

`internal/` is intentional: this is an app, not a reusable library.

## Conventions

- Interfaces named `Service`; implementations are lowercase `service` structs returned by a `New(...)` constructor.
- Each adapter has a `Config` struct grouping its tunables; `New` validates and applies defaults.
- Errors exported as `Err...` vars (e.g. `song.ErrCountMismatch`).
- Doc comments on every exported type/func.
- `cmd/croissant/main.go` uses bare-block scoping (`{ ... }`) to keep wiring sections visually grouped.
- **Struct literals always use multi-line format**, even for two fields:
  ```go
  // correct
  &service{
      cfg:    cfg,
      client: client,
  }

  // wrong
  &service{cfg: cfg, client: client}
  ```

## Running it

```bash
go mod download
go run ./cmd/croissant
# open http://localhost:8080
```

No credentials or config files needed — Deezer's API is public.

## Key design decisions

- **`song.Service.GetSongs(playlistID string)`** — playlist ID is a runtime parameter, not a
  constructor config. This keeps the service stateless and scalable to any number of playlists
  without instantiating multiple services.
- **3 hardcoded Deezer playlists** are offered on the home page (2000s / 2010s / 2020s hits).
  Their IDs are validated server-side in `handlers.go` (`allowedPlaylists` map) before being
  passed to `quizSvc.NewQuiz(playlistID)`.
- **HTMX** is used only for the answer feedback swap (`hx-post`, `hx-target="#content"`).
  Everything else is plain form POSTs with redirects — no JS required for navigation.
- **In-memory store** — quiz sessions are not persisted across server restarts. Fine for now.

## Known gaps / things to watch

- **No tests yet.** Quiz generation logic (`quiz/generator.go`) is the easiest first target —
  pure functions over `[]song.Song`.
- **`ErrCountMismatch`** fires if a Deezer playlist yields fewer preview-URL tracks than
  `SongsCount`. The service requests `SongsCount * 3` tracks to reduce this risk.
- **Session state** is in-memory only — a server restart loses all active quizzes.
