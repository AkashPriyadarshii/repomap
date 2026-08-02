# CLAUDE.md — repomap

## Project

Repo-map tool for AI coding agents. Third of the trilogy:
[zcat](https://github.com/AkashPriyadarshii/zcat) (read) →
[rustygrep](https://github.com/AkashPriyadarshii/rustygrep) (search) →
**repomap** (orient). Answers "what's in this repo" in ≤budget tokens.

## Commands

```bash
go build -o repomap.exe .      # build (Windows)
go test ./...                  # run tests
go vet ./...                   # static check
./repomap.exe [dir]            # map current/repo dir
./repomap.exe --json           # machine-readable
```

## Layout

```
cmd/repomap/main.go     CLI + render
internal/mapx/          all logic, stdlib only
docs/                   PRD, design, impl, this file
```

## Conventions

- **Stdlib only.** No external deps — deliberate. A dependency needs a
  written justification in the PR.
- Small focused functions, early returns, no deep nesting.
- Rank math in one place (`rank.go`); budget math in one place
  (`tokens.go`). Don't spread either.
- `ponytail:` comments mark deliberate simplifications and their ceiling
  (see symbol regexes).
- Verify every change: `go test ./...` before commit.
