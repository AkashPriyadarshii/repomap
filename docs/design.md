# repomap — Design

## Architecture

```
cmd/repomap/main.go   CLI: flags, output rendering
internal/mapx/        all logic (stdlib only)
```

Single package. No interfaces with one implementation, no config layer.

## Pipeline

```
Walk (os.WalkDir + .gitignore filter)
  → per-file: Symbol extraction (regex per language)
  → global: Import/ref graph from require/import lines
  → Rank (weighted score)
  → Budget-aware emit (tree or JSON)
```

## Symbol extraction

Regex per language, on first 200 lines of file (symbols are head-heavy).
Match: `func NAME(`, `func (r *T) NAME(`, `def NAME(`, `class NAME`,
`type NAME`, `interface NAME`, `const/var NAME`, plus `impl`, `fn`,
`sub`, `subroutine` for long-tail languages.

Ladder note: regex ≈80% accurate, zero deps. `ponytail: regex extractors,
swap to tree-sitter if a real codebase exposes misses.`

## Ranking

Weighted score per file:

| Signal | Weight | Source |
|--------|--------|--------|
| import in-degree | 0.45 | import/require/use lines |
| symbol count (log scale) | 0.25 | own extraction |
| size (log scale) | 0.15 | bytes |
| last commit recency | 0.10 | git log |
| name boost/penalty | 0.05 | main/handler/service↑, test/vendor/gen↓ |

Name penalty also gates **emission**, not just score: files matching
`test|spec|vendor|node_modules|generated|dist|build` get hidden at
`--full` off. Cheap, deterministic, fine to be wrong — a map orients.

## Token budget

Token estimate: `bytes/3.5` (≈4 chars/token, close enough for a budget).
Walk files in rank order, add until budget exhausted, then truncate:
```
… 412 files truncated, 3,214 symbols hidden (--full to see all)
```
`--budget` default 4000.

## JSON shape

```json
{
  "files": [
    {"path": "cmd/main.go", "refs": 12, "symbols": ["main", "setup"], "score": 0.9, "tokens": 180}
  ],
  "total_tokens": 3900,
  "truncated": 412
}
```

## Concurrency

One place it pays: walking thousands of files. Fan out symbol extraction
over `GOMAXPROCS` workers on a channel of files. Context for early exit.
Not premature — a real repo walks 10k+ files; this is the one genuine use
of goroutines here.

## Build & ship

Single static binary. `GOOS` cross-compile for macOS/Linux/Windows — same
trick as zcat. CI via GitHub Actions `goreleaser` later; v0.1.0 ships from
`go build`.
