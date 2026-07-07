# MCP Server for Leipzig Library Search

**Date:** 2026-07-07  
**Status:** Design Phase

## Overview

Expose the Leipzig city library catalog search functions as an MCP (Model Context Protocol) server so that LLM frontends (e.g., OpenWebUI, Claude Desktop) can call them as tools. A user asking an LLM "What Nintendo Switch games are available at Gohlis right now?" will have the LLM call the appropriate MCP tool to fetch live results from the library catalog.

## Problem Statement

- LLMs have no access to real-time library catalog data from Leipzig's WebOPAC
- OpenWebUI and other LLM frontends support MCP tools as a way to extend LLM capabilities
- The existing codebase already has clean catalog search functions; we need to expose them via MCP

## Solution Approach

Build a new standalone MCP server binary that:
1. Reuses the existing `library-le` scraping implementation via the `catalog.Client` interface
2. Wraps the 4 core search methods as MCP tools
3. Runs independently on port 8081 via HTTP/SSE transport
4. Resolves human-readable branch names to codes internally
5. Returns results as JSON for LLM consumption

## Architecture

```
cmd/mcp/
├── main.go          ← Entry point, flags, server initialization
├── tools.go         ← Tool definitions, handlers, branch resolution
└── tools_test.go    ← Unit tests with mock catalog.Client

[Shared]
├── library-le/      ← Existing WebOPAC scraping (unchanged)
├── catalog/         ← catalog.Client interface (unchanged)
└── domain/          ← Domain types: Movie, Game, Media (unchanged)
```

The MCP server is a completely separate binary from the existing web server. Both can run simultaneously on different ports (web: 3000, MCP: 8081).

## Transport

**HTTP/SSE** via `github.com/mark3labs/mcp-go`

Rationale:
- stdio transport is for local CLI tools (Claude Desktop)
- HTTP/SSE is required for web-based LLM frontends like OpenWebUI
- Clients connect to `http://<host>:8081/sse`

## The 4 MCP Tools

### Tool 1: `search_movies`

Search for a movie title across all Leipzig library branches.

**Input:**
- `title` (string, required) — movie title to search for

**Output:** JSON array of available copies
```json
[
  {
    "title": "Inception",
    "branch": "Gohlis",
    "platform": "bluray",
    "isAvailable": true
  }
]
```

**Wraps:** `catalog.Client.FindMovies(title string) ([]domain.Media, error)`

---

### Tool 2: `search_games`

Search for a game by title and platform across all branches.

**Input:**
- `title` (string, required) — game title to search for
- `platform` (string, required) — one of: `switch 1`, `switch 2`, `playstation`, `xbox`, `pc`

**Output:** JSON array of available copies
```json
[
  {
    "title": "Zelda: Breath of the Wild",
    "branch": "Gohlis",
    "platform": "switch 1",
    "isAvailable": true
  }
]
```

**Wraps:** `catalog.Client.FindGames(title, platform string) ([]domain.Media, error)`

---

### Tool 3: `list_available_games`

List all games currently available to borrow at a specific branch.

**Input:**
- `branch` (string, required) — branch name (see Valid Branches table below)
- `platform` (string, required) — one of: `switch 1`, `switch 2`, `playstation`, `xbox`, `pc`

**Output:** JSON array of available games
```json
[
  {
    "title": "Zelda: Breath of the Wild",
    "branch": "Gohlis",
    "platform": "switch 1"
  }
]
```

**Wraps:** `catalog.Client.FindAvailableGames(branchCode int, platform string) ([]domain.Game, error)`

---

### Tool 4: `get_return_date`

Get the expected return date for a checked-out item.

**Input:**
- `branch` (string, required) — branch name (see Valid Branches table below)
- `platform` (string, required) — one of: `switch 1`, `switch 2`, `playstation`, `xbox`, `pc`, `dvd`, `bluray`
- `title` (string, required) — exact title of the item

**Output:** Plain text date string
```
15.07.2026
```

**Wraps:** `catalog.Client.RetrieveReturnDate(branchCode int, platform, title string) (string, error)`

---

## Valid Branch Names

All branch-accepting tools (tools 3 and 4) accept these case-insensitive strings:

| Branch Name | Display Name |
|---|---|
| `stadtbibliothek` | Main library |
| `plagwitz` | Plagwitz |
| `wiederitzsch` | Wiederitzsch |
| `böhlitz` | Böhlitz-Ehrenberg |
| `lützschena` | Lützschena |
| `holzhausen` | Holzhausen |
| `südvorstadt` | Südvorstadt |
| `gohlis` | Gohlis |
| `volkmarsdorf` | Volkmarsdorf |
| `schönefeld` | Schönefeld |
| `paunsdorf` | Paunsdorf |
| `reudnitz` | Reudnitz |
| `mockau` | Mockau |
| `grünau-mitte` | Grünau-Mitte |
| `grünau-nord` | Grünau-Nord |
| `grünau-süd` | Grünau-Süd |

The server resolves branch names to integer codes using the existing `GetBranchCode()` function from `library-le/branches.go`. If an unknown branch name is provided, the tool returns a clear error listing all valid options.

## Error Handling

| Scenario | Response |
|---|---|
| Unknown branch name | Tool returns error: `"unknown branch 'atlantis'. Valid branches: stadtbibliothek, plagwitz, ..."` |
| Missing required parameter | Tool returns error: `"<param> is required"` |
| Upstream scraping error (network, HTML change) | Tool returns error: `"search failed: <underlying error>"` |
| No results found | Tool returns empty JSON array (not an error) |

## Configuration

| Flag | Default | Description |
|---|---|---|
| `--port` | `8081` | Port the MCP server listens on |

## Build & Deployment

### Local Development

```bash
make build-mcp   # Builds cmd/mcp/main.go → bin/mcp-server
./bin/mcp-server --port 8081
```

### Docker

```bash
docker build -f Dockerfile.mcp -t leipzig-library-mcp:latest .
docker run -p 8081:8081 leipzig-library-mcp:latest
```

### Integration with OpenWebUI

1. Start the MCP server: `./bin/mcp-server --port 8081`
2. In OpenWebUI → Settings → Tools → Add MCP Server
3. Enter URL: `http://<host>:8081/sse`
4. Label: `Leipzig Library`
5. The LLM will now have access to all 4 tools

## Technical Details

### Dependencies

- `github.com/mark3labs/mcp-go v0.55.1` — MCP server library
- Standard library only otherwise (existing code uses same)

### Type Mapping

- `domain.Media.IsAvailable` is `bool` → JSON `"isAvailable": true/false`
- `domain.Game.IsAvailable` is `string` → JSON `"isAvailable": "ausleihbar"`
- Tool results are JSON-serialized with `encoding/json`

### Session Management

Each tool call creates a fresh `libraryle.Client` instance. Since each `FindMovies`, `FindGames`, etc. call internally refreshes its own session, there's no concurrency concern.

## Out of Scope

- Authentication / API keys (follow existing web server's stateless posture)
- Watchlist integration (MCP is stateless; watchlist is session-based)
- stdio transport (not needed for OpenWebUI; can be added later)
- WebSocket transport (HTTP/SSE is sufficient)

## Success Criteria

- [ ] All 4 tools registered and callable via MCP
- [ ] Branch name resolution works with clear error messages
- [ ] JSON results serialize correctly
- [ ] Unit tests cover happy path and error cases
- [ ] Server starts on configurable port and listens on `/sse`
- [ ] Dockerfile.mcp builds successfully
- [ ] Makefile has `build-mcp` and `docker-build-mcp` targets
