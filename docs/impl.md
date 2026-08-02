# repomap — Implementation Plan (v0.1.0)

## Files

| File | Purpose |
|------|---------|
| `cmd/repomap/main.go` | CLI, flags, render |
| `internal/mapx/walk.go` | Walk + .gitignore filter |
| `internal/mapx/gitignore.go` | .gitignore matching (stdlib-only) |
| `internal/mapx/symbols.go` | regex symbol extraction |
| `internal/mapx/rank.go` | scoring + name gating |
| `internal/mapx/tokens.go` | token estimate + budget |
| `internal/mapx/render.go` | tree / JSON output |
| `internal/mapx/mapx_test.go` | one self-check test file |

## Order (verify after each)

1. **go.mod** — `module github.com/AkashPriyadarshii/repomap`, `go 1.26`
2. **gitignore.go** — parse lines into patterns, match path. Test: exclude
   `node_modules`, `.git`, `vendor`, `.env`, keep `src/app.go`.
3. **symbols.go** — regex per language. Test: Go func/method, Python
   class/def, TS interface.
4. **walk.go** — `os.WalkDir` + gitignore + import-ref counting. Test on
   temp fixture tree.
5. **rank.go** — weighted score + name gating.
6. **tokens.go** — estimate + budget truncation.
7. **render.go** — tree + JSON.
8. **main.go** — flags: `--budget`, `--full`, `--json`, `--no-gitignore`.
9. **README.md** — usage.
10. **Verify** — `go vet ./...`, `go test ./...`, `go build`, run on repo
    root, confirm tokens ≤ budget and JSON parses.
11. **Push** — git init, commit, `gh repo create`, push.

## Verification checklist

- [ ] `go vet ./...` clean
- [ ] `go test ./...` passes
- [ ] `go build -o repomap.exe` succeeds
- [ ] `./repomap.exe` on repo ≤ 4000 tokens
- [ ] `./repomap.exe --json` parses as valid JSON
- [ ] `.git`, `node_modules`, vendor excluded
