# Catalog Abstraction Layer — Design Spec

**Date:** 2026-06-14
**Goal:** Hide all HTML processing behind a Go interface so callers can query the library catalog without knowing about webOPAC scraping internals.
**Drivers:** Testability, separation of concerns, extensibility (swap catalog backends).

---

## 1. Package Structure & Dependency Flow

New package `catalog/` defines the `Client` interface. No HTTP, no HTML, no goquery lives there.

```
catalog/
  client.go   ← interface definition only
```

`library-le/` is restructured internally into clear layers but stays one package:

```
library-le/
  client.go   ← session management (unchanged)
  request.go  ← HTTP query building (unchanged)
  parser.go   ← ALL goquery/CSS-selector HTML parsing logic (new, extracted)
  opac.go     ← implements catalog.Client; wires session + HTTP + parser
  branches.go ← static branch lookup (unchanged)
```

Dependency flow (no cycles):

```
domain  ←  catalog  ←  library-le  ←  web
                                   ←  cmd/cli
```

`web/` and `cmd/cli` import `catalog` for the interface type and `library-le` only to
construct the concrete implementation — the same pattern as `database/sql` + a driver.

`GetBranchCode` is not HTML processing; it stays as a public function in `library-le`.
`web/` may import `library-le` for this utility without breaking the abstraction.

---

## 2. The `catalog.Client` Interface

```go
package catalog

import "github.com/gunni1/leipzig-library-media-search/domain"

type Client interface {
    FindMovies(title string) ([]domain.Media, error)
    FindGames(title, platform string) ([]domain.Media, error)
    FindAvailableGames(branchCode int, platform string) ([]domain.Game, error)
    RetrieveReturnDate(branchCode int, platform, title string) (string, error)
}
```

Notes:
- All methods return explicit errors. Callers distinguish "no results" (empty slice, nil error)
  from "failure" (non-nil error).
- Typo fixed: `FindAvailabelGames` → `FindAvailableGames`.
- `RetrieveReturnDate` already returned error; signature is unchanged.
- `GetBranchCode` is not on the interface — static lookup, not a catalog query.

---

## 3. `library-le` Internal Restructuring

### `parser.go` — pure HTML parsing

All `goquery` logic moves here. Functions take `*goquery.Document` or `io.Reader` and
return domain types or errors. No HTTP, no sessions, no side effects.

Functions that move here from `search.go` and `game-index.go`:
- `isSingleResultPage(doc *goquery.Document) bool`
- `extractTitles(doc *goquery.Document) []searchResult`
- `parseMediaCopiesPage(title string, doc *goquery.Document) ([]domain.Media, error)`
- `parseGameSearchResult(r io.Reader) ([]domain.Game, error)`
- `findReturnDateInCopiesPage(doc *goquery.Document) (string, error)`
- `determinePlatform(doc *goquery.Document) string`
- `isMediaAvailable(copy *goquery.Selection) bool`
- `isGameAvailable(node *goquery.Selection) bool`
- `extractDate(text string) (string, error)`
- `clearTitle(title string) string`
- `removeBranchSuffix(branchName string) string`
- `filterExactTitle(title string, results []searchResult) []searchResult`

All CSS selectors and regexes live in `parser.go` — one file to update when the
catalog's HTML structure changes.

### `opac.go` — implements `catalog.Client`

Thin orchestration layer only: create session → build request → fetch → call parser →
return result or error. Contains no HTML logic.

The `searchResult` struct (title + URL pair) is a plain data type; it stays as a private
type accessible to both `parser.go` and `opac.go` within the package. The HTTP-fetching
methods currently on `searchResult` (`loadMediaCopies`, `loadReturnDate`) move to
`opac.go` as package-level functions since they involve HTTP.

### Files removed

`search.go` and `game-index.go` are deleted once their logic is fully redistributed to
`parser.go` and `opac.go`.

---

## 4. Error Handling

Current behavior: errors are logged, `nil` is returned, callers cannot distinguish
failure from empty results.

New behavior:
- `parser.go` returns `error` when goquery fails or expected HTML structure is missing.
- `opac.go` wraps errors with context: `fmt.Errorf("FindMovies: %w", err)`.
- `web/` handlers receive errors and render a user-facing message (e.g.
  "Suche fehlgeschlagen") instead of silently showing empty results.
- "No results found" is **not** an error — it is an empty slice with `nil` error.

---

## 5. Testing Strategy

Three independent levels:

**Parser tests (`library-le/parser_test.go`)**
Feed fixture HTML from `testdata/` into parser functions, assert correct domain output.
No HTTP. This is the highest-value test level since the HTML structure is the part most
likely to break. Existing tests in `search_test.go` and `game-index_test.go` are
migrated here.

**`opac.go` tests (future, optional)**
Use `httptest.NewServer` to serve fixture HTML, assert correct orchestration — session
created, correct request built, errors surfaced. Lower priority; parser tests already
cover the fragile part.

**`web/` tests (`web/server_test.go`)**
Define a mock `catalog.Client` that implements the interface, inject it via a
constructor or parameter. Test handler logic (filtering, grouping by branch, template
rendering) without any HTTP to the real catalog. Existing `server_test.go` tests for
`arrangeByBranch` and `encodeBranch` are unaffected.

---

## 6. Out of Scope

- `domain.Movie` and `domain.Game` with their `IsAvailable string` fields appear to be
  unused legacy types. Cleanup is a separate task.
- Return date as `time.Time` instead of `string` — separate task.
- Parallel result fetching (existing TODO comment in `search.go`) — separate task.
