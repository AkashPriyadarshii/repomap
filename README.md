# repomap

[![Go](https://img.shields.io/badge/go-1.26-blue)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Release](https://img.shields.io/github/v/release/AkashPriyadarshii/repomap)](https://github.com/AkashPriyadarshii/repomap/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/AkashPriyadarshii/repomap)](https://pkg.go.dev/github.com/AkashPriyadarshii/repomap)
[![built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8)](https://go.dev/)

Repo map for AI coding agents. Answers "what's in this repo" in the fewest
tokens — orient before you read or search.

Third of a trilogy:
[zcat](https://github.com/AkashPriyadarshii/zcat) (read a file) →
[rustygrep](https://github.com/AkashPriyadarshii/rustygrep) (find a needle) →
**repomap** (orient: where am I, what's here).

## Why

`tree` drowns an LLM context window before the agent reads one file — and a
5000-line listing is useless. repomap walks the repo, ranks files by
importance for an agent, and emits a token-budgeted map.

## Install

```bash
go install github.com/AkashPriyadarshii/repomap/cmd/repomap@latest
# or build locally
go build -o repomap ./cmd/repomap
```

## Usage

```bash
repomap                  # ranked tree, token-capped (default 4000)
repomap --budget 2000    # cap output
repomap --full           # no truncation
repomap --json           # machine-readable
repomap --no-gitignore   # include ignored files (warning: huge)
repomap ./path/to/repo   # map a specific dir
```

## Example

```
$ repomap
├── cmd/server/main.go      59 refs  12 symbols  [main, setupRouter, serveHTTP]
├── internal/auth/
│   ├── jwt.go              41 refs   8 symbols  [Issue, Verify, ParseToken]
│   └── middleware.go       22 refs   4 symbols  [RequireAuth, OptionalAuth]
│
… 412 files truncated, 3,214 symbols hidden (--full to see all)
```

## JSON

`repomap --json` — same pattern as zcat/rustygrep. Agent gets a machine
map, not a human tree to re-parse:

```json
{
  "files": [
    {"path": "cmd/main.go", "refs": 12, "symbols": ["main", "setup"], "score": 0.9, "tokens": 180}
  ],
  "total_tokens": 3900,
  "truncated": 412
}
```

## Docs

- [docs/PRD.md](docs/PRD.md) — product, goals, interface
- [docs/design.md](docs/design.md) — architecture, ranking, budget
- [docs/impl.md](docs/impl.md) — implementation plan
- [docs/CLAUDE.md](docs/CLAUDE.md) — project conventions

## License

MIT
