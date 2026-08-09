# AGENTS.md — Leipzig Library Media Search

Go web app and CLI that scrapes the Leipzig city library (Stadtbibliothek Leipzig) WebOPAC catalog to search for movies and games. Exposes an HTTP API with server-side HTML rendering (HTMX-style partials) and a persistent watchlist.

**Module:** `github.com/gunni1/leipzig-library-media-search`  
**Go version:** 1.24 (go.mod: 1.24.4)  
**Key dependencies:** `goquery` (HTML scraping), `testify` (assertions), `pkg/errors`

---

## Build Commands

```bash
go build ./...                          # verify all packages compile
go build -o bin/web main.go             # web server binary
go build -o bin/cli cmd/cli/main.go     # CLI binary
make build-web                          # build ./... + go build -o bin/web (no tests)
make build-amd64                        # cross-compile for Linux amd64 -> bin/llms-amd64-linux
make build-arm64-linux                  # cross-compile for Linux arm64 -> bin/lib-api-arm64-linux
make build-cli                          # build CLI -> bin/cli
```

> **Note:** `make build-web` does NOT run tests. Run `go test ./...` separately.

Docker:
```bash
make docker-build   # builds AND pushes to Docker Hub (docker push is included)
# No docker-run make target — run manually:
docker run -p 3000:3000 --rm --name lib-api gunni1/leipzig-library-media-search:latest
```

The web server listens on port **3000** (default), not 8080. Flags: `-port`, `-data-dir`.

---

## Test Commands

```bash
go test ./...                                              # all tests
go test -v ./library-le/ -run TestParseGameCopiesResult    # single test
go test -v ./library-le/ -run TestDetermine                # pattern match
go test -v ./web/ -run TestArrangeByBranch
```

**Pattern:** `go test -v ./<package-dir>/ -run <TestFunctionName>`

Formatting: `gofmt -l .` to check, `gofmt -w .` to fix. No linter configured.

---

## Project Structure

```
main.go                  # web server entry: -port (default 3000), -data-dir flags
domain/media.go          # core types: Movie, Game, Media — no HTTP/scraping concerns
library-le/              # package libraryle — scraping, search, branch mapping
  client.go              # HTTP client, session management
  branches.go            # branch name <-> branch code mapping
  request.go             # HTTP request builders for WebOPAC API
  search.go              # HTML parsing, search result processing
  game-index.go          # game availability search
  testdata/              # HTML fixtures for unit tests
watchlist/store.go       # file-backed watchlist (was in-memory, now persists to data/)
web/server.go            # HTTP mux, route handlers, template rendering
web/templates/           # Go html/template files (embedded via //go:embed)
web/static/              # static assets (embedded via //go:embed)
cmd/cli/main.go          # CLI entry point
data/                    # watchlist persistence directory (JSON files)
scripts/                 # bootstrap.sh and other ops scripts
```

---

## Architecture Notes

- **Watchlist persistence:** `watchlist.NewFileStore(dataDir)` creates a file-backed store in the `data/` directory. The old in-memory store is gone. Docker mounts `/data` as a volume.
- **Sessions:** Cookie `wl_session` (UUID). Watchlist store is keyed by session ID.
- **Templates/static assets** are embedded at build time — changes require a rebuild.
- **`library-le/` directory uses package name `libraryle`** (hyphen in dir, no hyphen in package).

---

## Testing Conventions

- Tests are **white-box** (same package as production code).
- Use `testify/assert` with a **dot import** — assertions are unqualified:
  ```go
  import (
      "testing"
      . "github.com/stretchr/testify/assert"
  )
  func TestFoo(t *testing.T) { Equal(t, expected, actual) }
  ```
- HTML scraping tests use fixture files from `library-le/testdata/`. Load via `loadTestData(filePath)` (defined in `client_test.go`) then `asDoc()`.
- No one-character variable names.
- Comments may be English or German — both appear in the codebase.

---

## Import Grouping

```go
import (
    // 1. stdlib
    "fmt"

    // 2. third-party
    "github.com/PuerkitoBio/goquery"

    // 3. internal
    "github.com/gunni1/leipzig-library-media-search/domain"
)
```

---

## CI

`.github/workflows/go.yml` on push/PR to `main`:
```
go mod download -> go build -v ./... -> go test -v -json > testresults.json ./...
```
Test results uploaded as artifact `Results/testresults.json`.
