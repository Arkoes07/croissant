# CLAUDE.md

Context for future Claude Code sessions working in this repo.

## What this project is

**Croissant** is a song-guessing quiz game. The server fetches songs (with 30-second
preview clips) from a Spotify playlist via the Spotify Web API, then asks the user
to guess each song's title from its audio preview.

The repo started as a small proof-of-concept that just fetched and printed songs.
The active direction is to grow it into a playable web-based quiz game.

## Tech stack

- **Language:** Go 1.24 (`go.mod` and local toolchain `1.24.5` are aligned).
- **Spotify API client:** `github.com/zmb3/spotify/v2` v2.4.3; `TokenURL` lives in sub-package `github.com/zmb3/spotify/v2/auth`.
- **Auth:** OAuth2 client-credentials flow via `golang.org/x/oauth2/clientcredentials`; v2 client created with `spotify.New(authConfig.Client(ctx))`.
- **HTTP / UI (planned):** stdlib `net/http` + `html/template` + HTMX, with HTML5 `<audio>` for preview playback.
- **Config (current):** local JSON file at `files/secret/secret.json` (template: `files/secret/secret-template.json`).

## Architecture

Hexagonal (ports & adapters). Each domain package defines a `Service` interface
(the port); subpackages provide concrete adapters.

### Current layout

```
croissant/
├── cmd/croissant/main.go            ← entrypoint, wires services together
├── files/secret/
│   ├── secret-template.json         ← template; copy to secret.json and fill in
│   └── secret.json                  ← (gitignored) real Spotify credentials
└── internal/
    ├── secret/
    │   ├── secret.go                ← Secret model + Service port
    │   └── jsonsecret/service.go    ← JSON-file adapter (uses os.ReadFile)
    └── song/
        ├── song.go                  ← Song{Title, Artists, PreviewURL} + Service port + ErrCountMismatch
        └── spotifyservice/service.go ← Spotify v2 adapter; fetches a playlist, filters tracks with no preview
```

`internal/` is intentional: this is an app, not a reusable library.

### Planned layout (when quiz/web work begins)

```
croissant/
├── cmd/croissant/main.go
├── internal/
│   ├── config/                      ← env / file config loader
│   ├── secret/      (+ jsonsecret/) ← already exists
│   ├── song/        (+ spotifyservice/) ← already exists
│   ├── quiz/                        ← NEW: Quiz, Question, Generator, Store
│   │   └── memorystore/
│   └── web/                         ← NEW: handlers, middleware, templates, static
│       ├── templates/
│       └── static/
└── files/secret/
```

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
# 1. Set up Spotify credentials
cp files/secret/secret-template.json files/secret/secret.json
# edit secret.json and fill in client_id / client_secret from
# https://developer.spotify.com/dashboard

# 2. Build & run
go mod download
go run ./cmd/croissant
```

The current entrypoint (`cmd/croissant/main.go`) logs the first 10 songs from
the hardcoded playlist `37i9dQZF1DXcBWIGoYBM5M` (Spotify's "Today's Top Hits").
The `// TODO: will remove soon, only for demo purpose` comment marks that block
for replacement when the quiz/web layer lands.

## Known gaps / things to watch

- **Hardcoded config** (`PlaylistID`, `SongsCount`) lives in `cmd/croissant/main.go` — flagged with a TODO, should move to a `config` package reading env vars.
- **`ErrCountMismatch`** fires if a playlist yields fewer preview-URL-bearing tracks than `SongsCount`. Newer Spotify playlists increasingly have `PreviewURL == ""`, so quiz generation should expect to scan more tracks than it needs.
- **No tests yet.** Quiz logic (once added) is the easy first target — pure functions over `[]song.Song`.

## Quiz game design (target)

- HTTP server using stdlib `net/http` (Go 1.22+ ServeMux is sufficient).
- Server-rendered HTML with `html/template`; HTMX for partial swaps (no SPA).
- Audio playback via `<audio src="{preview_url}">` directly in the browser.
- Session state (current quiz, score, question index) keyed by signed cookie; `quiz.Store` interface with an in-memory implementation first.
- A `quiz.Generator` builds `Question{Correct, Choices}` from a pool of songs by sampling distractors.
- Routes (sketch): `GET /`, `POST /quiz/new`, `GET /quiz/{id}`, `POST /quiz/{id}/answer`, `GET /quiz/{id}/result`.
