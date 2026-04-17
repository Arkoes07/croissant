# Croissant 🥐

A song-guessing quiz game built with Go. Listen to 30-second audio previews and pick the correct song title from four choices. Songs are sourced from Deezer's public API — no API key required.

**Live demo:** [croissant.alwisukra.com](https://croissant.alwisukra.com)

## Features

- 10 questions per game, 4 choices each
- 3 era playlists to choose from: 2000s, 2010s, 2020s hits
- Custom dark-themed audio player
- Server-rendered HTML with HTMX for seamless partial updates
- Hexagonal architecture (ports & adapters)

## Tech Stack

- **Language:** Go 1.24
- **HTTP:** stdlib `net/http` (Go 1.22+ ServeMux)
- **UI:** `html/template` + HTMX + HTML5 `<audio>`
- **Songs:** [Deezer API](https://developers.deezer.com/api) (no credentials needed)

## Running Locally

```bash
# 1. Clone the repo
git clone https://github.com/Arkoes07/croissant.git
cd croissant

# 2. Download dependencies
go mod download

# 3. Run
go run ./cmd/croissant
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Configuration

| Variable | Default | Description                     |
|----------|---------|---------------------------------|
| `PORT`   | `8080`  | Port the HTTP server listens on |

## Deployment

The repo includes a `Dockerfile` for container-based deployments. It uses a multi-stage build with `golang:1.24-alpine`.

```bash
docker build -t croissant .
docker run -p 8080:8080 croissant
```

This project is self-hosted on a [Coolify](https://coolify.io) instance at [croissant.alwisukra.com](https://croissant.alwisukra.com) using the Dockerfile build pack.

## Architecture

Hexagonal (ports & adapters). Domain packages define `Service` interfaces; subpackages provide concrete adapters.

```
croissant/
├── cmd/croissant/main.go          ← entrypoint, wires everything together
└── internal/
    ├── song/                      ← Song model + Service port
    │   └── deezerservice/         ← Deezer API adapter
    ├── quiz/                      ← Quiz domain: Generator, Store, Service port
    │   ├── memorystore/           ← In-memory Store adapter
    │   └── quizservice/           ← Quiz Service adapter
    └── web/                       ← HTTP handlers, server, templates
        └── templates/             ← HTML templates (layout, home, question, answer, result)
```

## License

MIT