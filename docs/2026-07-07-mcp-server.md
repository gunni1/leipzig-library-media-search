# MCP Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a standalone HTTP/SSE MCP server that exposes the Leipzig library catalog search as 4 callable tools for LLM frontends like OpenWebUI.

**Architecture:** New Go binary at `cmd/mcp/main.go` using `mark3labs/mcp-go`. Tool handlers and registration in `cmd/mcp/tools.go`, tested with a mock `catalog.Client` in `cmd/mcp/tools_test.go`. Separate `Dockerfile.mcp` for containerization.

**Tech Stack:** Go 1.25+, `github.com/mark3labs/mcp-go v0.55.1`, existing `library-le` and `catalog` packages, standard library.

---

## File Structure

| File | Action | Purpose |
|---|---|---|
| `cmd/mcp/main.go` | Create | Server entry point: flags, logging, initialization |
| `cmd/mcp/tools.go` | Create | Tool definitions, handlers, branch resolution, utilities |
| `cmd/mcp/tools_test.go` | Create | Unit tests with mock catalog.Client |
| `Makefile` | Modify | Add `build-mcp` and `docker-build-mcp` targets |
| `Dockerfile.mcp` | Create | Multi-stage Docker build (same pattern as existing Dockerfile) |
| `go.mod` / `go.sum` | Modify | Add `github.com/mark3labs/mcp-go v0.55.1` |

---

## Task 1: Add mcp-go Dependency

**Files:**
- `go.mod`, `go.sum`

- [ ] **Step 1: Fetch the dependency**

```bash
go get github.com/mark3labs/mcp-go@v0.55.1
```

This updates `go.mod` and `go.sum` automatically.

- [ ] **Step 2: Verify it's present**

```bash
grep "mark3labs/mcp-go" go.mod
```

Expected: `github.com/mark3labs/mcp-go v0.55.1`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add github.com/mark3labs/mcp-go v0.55.1"
```

---

## Task 2: Create tools.go with search_movies Tool

**Files:**
- Create: `cmd/mcp/tools.go`

- [ ] **Step 1: Create the file structure**

Create `cmd/mcp/tools.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gunni1/leipzig-library-media-search/catalog"
	libraryle "github.com/gunni1/leipzig-library-media-search/library-le"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const validBranchesList = "stadtbibliothek, plagwitz, wiederitzsch, böhlitz, lützschena, holzhausen, südvorstadt, gohlis, volkmarsdorf, schönefeld, paunsdorf, reudnitz, mockau, grünau-mitte, grünau-nord, grünau-süd"

// RegisterTools registers all 4 Leipzig library MCP tools on the given server.
func RegisterTools(s *server.MCPServer, client catalog.Client) {
	s.AddTool(searchMoviesTool(), searchMoviesHandler(client))
}

// searchMoviesTool defines the MCP tool for searching movies.
func searchMoviesTool() mcp.Tool {
	return mcp.NewTool("search_movies",
		mcp.WithDescription("Search for movies available to borrow at Leipzig city library branches. Returns copies per branch with availability status and format (dvd or bluray)."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Movie title to search for")),
	)
}

// searchMoviesHandler returns the handler function for the search_movies tool.
func searchMoviesHandler(client catalog.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		results, err := client.FindMovies(title)
		if err != nil {
			return mcp.NewToolResultErrorf("search failed: %v", err), nil
		}

		return mcp.NewToolResultText(toJSON(results)), nil
	}
}

// toJSON marshals v to a compact JSON string.
func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// resolveBranch converts a human-readable branch name to its integer code.
// Returns an error with a list of valid names if the name is unknown.
func resolveBranch(name string) (int, error) {
	code, ok := libraryle.GetBranchCode(name)
	if !ok {
		return 0, fmt.Errorf("unknown branch %q. Valid branches: %s", name, validBranchesList)
	}
	return code, nil
}
```

- [ ] **Step 2: Build to verify no syntax errors**

```bash
go build ./cmd/mcp/
```

Expected: no errors; a `mcp` executable is created (which we can delete).

- [ ] **Step 3: Clean up and commit**

```bash
rm -f mcp
git add cmd/mcp/tools.go
git commit -m "feat(mcp): add search_movies tool definition"
```

---

## Task 3: Create tools_test.go with Mock and Tests

**Files:**
- Create: `cmd/mcp/tools_test.go`

- [ ] **Step 1: Create the test file with mock client**

Create `cmd/mcp/tools_test.go`:

```go
package main

import (
	"context"
	"errors"
	"testing"

	"github.com/gunni1/leipzig-library-media-search/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCatalogClient satisfies catalog.Client for testing.
type mockCatalogClient struct {
	findMoviesResult         []domain.Media
	findMoviesErr            error
	findGamesResult          []domain.Media
	findGamesErr             error
	findAvailableGamesResult []domain.Game
	findAvailableGamesErr    error
	retrieveReturnDateResult string
	retrieveReturnDateErr    error
}

func (m *mockCatalogClient) FindMovies(title string) ([]domain.Media, error) {
	return m.findMoviesResult, m.findMoviesErr
}

func (m *mockCatalogClient) FindGames(title, platform string) ([]domain.Media, error) {
	return m.findGamesResult, m.findGamesErr
}

func (m *mockCatalogClient) FindAvailableGames(branchCode int, platform string) ([]domain.Game, error) {
	return m.findAvailableGamesResult, m.findAvailableGamesErr
}

func (m *mockCatalogClient) RetrieveReturnDate(branchCode int, platform, title string) (string, error) {
	return m.retrieveReturnDateResult, m.retrieveReturnDateErr
}

// callHandler invokes a ToolHandlerFunc with the given arguments map.
func callHandler(t *testing.T, handler server.ToolHandlerFunc, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	return result
}

// textContent extracts the first text content from a result, or returns empty string.
func textContent(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return ""
	}
	return tc.Text
}

// Tests for search_movies

func TestSearchMovies_ReturnsResults(t *testing.T) {
	mock := &mockCatalogClient{
		findMoviesResult: []domain.Media{
			{Title: "Inception", Branch: "Gohlis", Platform: "bluray", IsAvailable: true},
		},
	}
	result := callHandler(t, searchMoviesHandler(mock), map[string]any{"title": "inception"})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "Inception")
}

func TestSearchMovies_ClientError(t *testing.T) {
	mock := &mockCatalogClient{findMoviesErr: errors.New("network timeout")}
	result := callHandler(t, searchMoviesHandler(mock), map[string]any{"title": "inception"})
	assert.True(t, result.IsError)
	assert.Contains(t, textContent(result), "search failed")
}

func TestSearchMovies_MissingTitle(t *testing.T) {
	mock := &mockCatalogClient{}
	result := callHandler(t, searchMoviesHandler(mock), map[string]any{})
	assert.True(t, result.IsError)
}

func TestSearchMovies_EmptyResults(t *testing.T) {
	mock := &mockCatalogClient{findMoviesResult: []domain.Media{}}
	result := callHandler(t, searchMoviesHandler(mock), map[string]any{"title": "nonexistent"})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "[]")
}
```

- [ ] **Step 2: Run the search_movies tests**

```bash
go test ./cmd/mcp/ -run TestSearchMovies -v
```

Expected: all 4 tests PASS.

- [ ] **Step 3: Commit**

```bash
git add cmd/mcp/tools_test.go
git commit -m "feat(mcp): add mock client and search_movies tests"
```

---

## Task 4: Add search_games Tool

**Files:**
- Modify: `cmd/mcp/tools.go`
- Modify: `cmd/mcp/tools_test.go`

- [ ] **Step 1: Update RegisterTools in tools.go**

In `tools.go`, replace the `RegisterTools` function:

```go
func RegisterTools(s *server.MCPServer, client catalog.Client) {
	s.AddTool(searchMoviesTool(), searchMoviesHandler(client))
	s.AddTool(searchGamesTool(), searchGamesHandler(client))
}
```

- [ ] **Step 2: Add tool definition and handler to tools.go**

Append to `tools.go`:

```go
// searchGamesTool defines the MCP tool for searching games.
func searchGamesTool() mcp.Tool {
	return mcp.NewTool("search_games",
		mcp.WithDescription("Search for games available to borrow at Leipzig city library branches by title and platform."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Game title to search for")),
		mcp.WithString("platform", mcp.Required(), mcp.Description("Platform to filter by. Valid values: switch 1, switch 2, playstation, xbox")),
	)
}

// searchGamesHandler returns the handler function for the search_games tool.
func searchGamesHandler(client catalog.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		platform, err := req.RequireString("platform")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		results, err := client.FindGames(title, platform)
		if err != nil {
			return mcp.NewToolResultErrorf("search failed: %v", err), nil
		}

		return mcp.NewToolResultText(toJSON(results)), nil
	}
}
```

- [ ] **Step 3: Add tests to tools_test.go**

Append to `tools_test.go`:

```go
// Tests for search_games

func TestSearchGames_ReturnsResults(t *testing.T) {
	mock := &mockCatalogClient{
		findGamesResult: []domain.Media{
			{Title: "Zelda", Branch: "Gohlis", Platform: "switch 1", IsAvailable: true},
		},
	}
	result := callHandler(t, searchGamesHandler(mock), map[string]any{
		"title":    "zelda",
		"platform": "switch 1",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "Zelda")
}

func TestSearchGames_ClientError(t *testing.T) {
	mock := &mockCatalogClient{findGamesErr: errors.New("scrape error")}
	result := callHandler(t, searchGamesHandler(mock), map[string]any{
		"title":    "zelda",
		"platform": "switch 1",
	})
	assert.True(t, result.IsError)
}

func TestSearchGames_MissingTitle(t *testing.T) {
	result := callHandler(t, searchGamesHandler(&mockCatalogClient{}), map[string]any{
		"platform": "switch 1",
	})
	assert.True(t, result.IsError)
}

func TestSearchGames_MissingPlatform(t *testing.T) {
	result := callHandler(t, searchGamesHandler(&mockCatalogClient{}), map[string]any{
		"title": "zelda",
	})
	assert.True(t, result.IsError)
}

func TestSearchGames_EmptyResults(t *testing.T) {
	mock := &mockCatalogClient{findGamesResult: []domain.Media{}}
	result := callHandler(t, searchGamesHandler(mock), map[string]any{
		"title":    "nonexistent",
		"platform": "switch 1",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "[]")
}
```

- [ ] **Step 4: Run all tests**

```bash
go test ./cmd/mcp/ -run TestSearch -v
```

Expected: all 9 tests PASS (4 search_movies + 5 search_games).

- [ ] **Step 5: Commit**

```bash
git add cmd/mcp/tools.go cmd/mcp/tools_test.go
git commit -m "feat(mcp): add search_games tool with tests"
```

---

## Task 5: Add list_available_games Tool

**Files:**
- Modify: `cmd/mcp/tools.go`
- Modify: `cmd/mcp/tools_test.go`

- [ ] **Step 1: Update RegisterTools in tools.go**

In `tools.go`, update `RegisterTools`:

```go
func RegisterTools(s *server.MCPServer, client catalog.Client) {
	s.AddTool(searchMoviesTool(), searchMoviesHandler(client))
	s.AddTool(searchGamesTool(), searchGamesHandler(client))
	s.AddTool(listAvailableGamesTool(), listAvailableGamesHandler(client))
}
```

- [ ] **Step 2: Add tool definition and handler to tools.go**

Append to `tools.go`:

```go
// listAvailableGamesTool defines the MCP tool for listing available games at a branch.
func listAvailableGamesTool() mcp.Tool {
	return mcp.NewTool("list_available_games",
		mcp.WithDescription("List all games currently available to borrow at a specific Leipzig library branch and platform. Valid branches: "+validBranchesList),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Library branch name (case-insensitive), e.g., gohlis, plagwitz, stadtbibliothek")),
		mcp.WithString("platform", mcp.Required(), mcp.Description("Platform to filter by. Valid values: switch 1, switch 2, playstation, xbox")),
	)
}

// listAvailableGamesHandler returns the handler function for the list_available_games tool.
func listAvailableGamesHandler(client catalog.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		branch, err := req.RequireString("branch")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		platform, err := req.RequireString("platform")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		branchCode, err := resolveBranch(branch)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		results, err := client.FindAvailableGames(branchCode, platform)
		if err != nil {
			return mcp.NewToolResultErrorf("search failed: %v", err), nil
		}

		return mcp.NewToolResultText(toJSON(results)), nil
	}
}
```

- [ ] **Step 3: Add tests to tools_test.go**

Append to `tools_test.go`:

```go
// Tests for list_available_games

func TestListAvailableGames_ReturnsResults(t *testing.T) {
	mock := &mockCatalogClient{
		findAvailableGamesResult: []domain.Game{
			{Title: "Zelda", Branch: "Gohlis", Platform: "switch 1", IsAvailable: "ausleihbar"},
		},
	}
	result := callHandler(t, listAvailableGamesHandler(mock), map[string]any{
		"branch":   "gohlis",
		"platform": "switch 1",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "Zelda")
}

func TestListAvailableGames_UnknownBranch(t *testing.T) {
	result := callHandler(t, listAvailableGamesHandler(&mockCatalogClient{}), map[string]any{
		"branch":   "atlantis",
		"platform": "switch 1",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, textContent(result), "Valid branches")
}

func TestListAvailableGames_ClientError(t *testing.T) {
	mock := &mockCatalogClient{findAvailableGamesErr: errors.New("connection failed")}
	result := callHandler(t, listAvailableGamesHandler(mock), map[string]any{
		"branch":   "gohlis",
		"platform": "switch 1",
	})
	assert.True(t, result.IsError)
}

func TestListAvailableGames_EmptyResults(t *testing.T) {
	mock := &mockCatalogClient{findAvailableGamesResult: []domain.Game{}}
	result := callHandler(t, listAvailableGamesHandler(mock), map[string]any{
		"branch":   "gohlis",
		"platform": "switch 1",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "[]")
}
```

- [ ] **Step 4: Run all tests**

```bash
go test ./cmd/mcp/ -v
```

Expected: all 13 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/mcp/tools.go cmd/mcp/tools_test.go
git commit -m "feat(mcp): add list_available_games tool with tests"
```

---

## Task 6: Add get_return_date Tool

**Files:**
- Modify: `cmd/mcp/tools.go`
- Modify: `cmd/mcp/tools_test.go`

- [ ] **Step 1: Update RegisterTools in tools.go**

In `tools.go`, update `RegisterTools`:

```go
func RegisterTools(s *server.MCPServer, client catalog.Client) {
	s.AddTool(searchMoviesTool(), searchMoviesHandler(client))
	s.AddTool(searchGamesTool(), searchGamesHandler(client))
	s.AddTool(listAvailableGamesTool(), listAvailableGamesHandler(client))
	s.AddTool(getReturnDateTool(), getReturnDateHandler(client))
}
```

- [ ] **Step 2: Add tool definition and handler to tools.go**

Append to `tools.go`:

```go
// getReturnDateTool defines the MCP tool for retrieving return dates.
func getReturnDateTool() mcp.Tool {
	return mcp.NewTool("get_return_date",
		mcp.WithDescription("Get the expected return date for a checked-out item at a specific Leipzig library branch. Returns a date in DD.MM.YYYY format. Valid branches: "+validBranchesList),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Library branch name (case-insensitive), e.g., gohlis, plagwitz, stadtbibliothek")),
		mcp.WithString("platform", mcp.Required(), mcp.Description("Item platform or format. Valid values: switch 1, switch 2, playstation, xbox, dvd, bluray")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Exact title of the item")),
	)
}

// getReturnDateHandler returns the handler function for the get_return_date tool.
func getReturnDateHandler(client catalog.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		branch, err := req.RequireString("branch")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		platform, err := req.RequireString("platform")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		branchCode, err := resolveBranch(branch)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		date, err := client.RetrieveReturnDate(branchCode, platform, title)
		if err != nil {
			return mcp.NewToolResultErrorf("retrieval failed: %v", err), nil
		}

		return mcp.NewToolResultText(date), nil
	}
}
```

- [ ] **Step 3: Add tests to tools_test.go**

Append to `tools_test.go`:

```go
// Tests for get_return_date

func TestGetReturnDate_ReturnsDate(t *testing.T) {
	mock := &mockCatalogClient{retrieveReturnDateResult: "15.07.2026"}
	result := callHandler(t, getReturnDateHandler(mock), map[string]any{
		"branch":   "gohlis",
		"platform": "switch 1",
		"title":    "Zelda",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, textContent(result), "15.07.2026")
}

func TestGetReturnDate_UnknownBranch(t *testing.T) {
	result := callHandler(t, getReturnDateHandler(&mockCatalogClient{}), map[string]any{
		"branch":   "narnia",
		"platform": "switch 1",
		"title":    "Zelda",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, textContent(result), "Valid branches")
}

func TestGetReturnDate_ClientError(t *testing.T) {
	mock := &mockCatalogClient{retrieveReturnDateErr: errors.New("item not found")}
	result := callHandler(t, getReturnDateHandler(mock), map[string]any{
		"branch":   "gohlis",
		"platform": "switch 1",
		"title":    "Zelda",
	})
	assert.True(t, result.IsError)
}

func TestGetReturnDate_MissingBranch(t *testing.T) {
	result := callHandler(t, getReturnDateHandler(&mockCatalogClient{}), map[string]any{
		"platform": "switch 1",
		"title":    "Zelda",
	})
	assert.True(t, result.IsError)
}
```

- [ ] **Step 4: Run all tests**

```bash
go test ./cmd/mcp/ -v
```

Expected: all 17 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/mcp/tools.go cmd/mcp/tools_test.go
git commit -m "feat(mcp): add get_return_date tool with tests"
```

---

## Task 7: Create main.go Entry Point

**Files:**
- Create: `cmd/mcp/main.go`

- [ ] **Step 1: Create main.go**

Create `cmd/mcp/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"log"

	libraryle "github.com/gunni1/leipzig-library-media-search/library-le"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	port := flag.Int("port", 8081, "MCP server port")
	flag.Parse()

	// Create a fresh Leipzig library client for each tool call
	client := libraryle.Client{}

	// Initialize MCP server
	mcpServer := server.NewMCPServer("Leipzig Library", "1.0.0")

	// Register all tools
	RegisterTools(mcpServer, client)

	// Set up SSE transport
	addr := fmt.Sprintf(":%d", *port)
	baseURL := fmt.Sprintf("http://localhost:%d", *port)
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(baseURL))

	log.Printf("MCP server listening on %s/sse\n", baseURL)
	log.Fatal(sseServer.Start(addr))
}
```

- [ ] **Step 2: Build and verify**

```bash
go build ./cmd/mcp/
```

Expected: builds successfully; `mcp` executable created.

- [ ] **Step 3: Test that it compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Clean up and commit**

```bash
rm -f mcp
git add cmd/mcp/main.go
git commit -m "feat(mcp): add MCP server entry point with main.go"
```

---

## Task 8: Add Makefile Targets

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add build targets to Makefile**

Append to `Makefile`:

```makefile
build-mcp:
	go build ./...
	go test ./...
	go build -o bin/mcp-server cmd/mcp/main.go

docker-build-mcp:
	docker build -f Dockerfile.mcp -t gunni1/leipzig-library-mcp-server:latest .
	docker push gunni1/leipzig-library-mcp-server:latest
```

- [ ] **Step 2: Test the build target**

```bash
make build-mcp
```

Expected: tests pass, `bin/mcp-server` binary created.

- [ ] **Step 3: Verify binary works**

```bash
./bin/mcp-server --help
```

Expected: prints usage information (from Go's flag package).

- [ ] **Step 4: Commit**

```bash
git add Makefile
git commit -m "chore(mcp): add build-mcp and docker-build-mcp Makefile targets"
```

---

## Task 9: Create Dockerfile.mcp

**Files:**
- Create: `Dockerfile.mcp`

- [ ] **Step 1: Create Dockerfile.mcp**

Create `Dockerfile.mcp`:

```dockerfile
# Stage 1: Build
# Cross-compilation: build on host arch, Go cross-compiles to TARGETARCH.
FROM golang:1.26 AS build

ARG TARGETARCH=arm64
ENV GOOS=linux
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o bin/mcp-server cmd/mcp/main.go

# Stage 2: Run
FROM --platform=linux/arm64 gcr.io/distroless/static-debian13
WORKDIR /
COPY --from=build /app/bin/mcp-server /mcp-server

EXPOSE 8081
USER nonroot:nonroot

ENTRYPOINT ["/mcp-server"]
```

- [ ] **Step 2: Build the Docker image**

```bash
docker build -f Dockerfile.mcp -t leipzig-library-mcp-server:dev .
```

Expected: image builds without errors; final image is distroless (small, fast, secure).

- [ ] **Step 3: Verify it runs**

```bash
docker run --rm -p 8081:8081 leipzig-library-mcp-server:dev &
sleep 2
curl http://localhost:8081/sse -I
kill %1
```

Expected: HTTP 200 response from the SSE endpoint (or 405 if not GET; either means the server started).

- [ ] **Step 4: Commit**

```bash
git add Dockerfile.mcp
git commit -m "feat(mcp): add Dockerfile.mcp for containerization"
```

---

## Summary

**9 tasks completed:**
1. Add mcp-go dependency
2. Create tools.go with search_movies tool
3. Create tools_test.go with mock client and tests
4. Add search_games tool
5. Add list_available_games tool
6. Add get_return_date tool
7. Create main.go entry point
8. Add Makefile targets
9. Create Dockerfile.mcp

**17 unit tests:**
- 4 tests for search_movies
- 5 tests for search_games
- 4 tests for list_available_games
- 4 tests for get_return_date

**Result:** Fully functional MCP server binary that:
- Exposes 4 tools via HTTP/SSE
- Reuses existing library-le scraping
- Handles branch name resolution with clear error messages
- Returns JSON results for LLM consumption
- Containerizes via Docker
- Builds with `make build-mcp`
