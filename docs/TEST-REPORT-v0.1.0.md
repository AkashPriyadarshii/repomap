# repomap v0.1.0 — Atomic Test Report

Date: 2026-08-02 · Go 1.26.4 windows/amd64

## Result: PASS (3/3)

`go vet ./...` — clean
`go build ./...` — success
`go test ./...` — **3 passed in 2 packages**

## Tests

| Test | Covers | Status |
|------|--------|--------|
| `TestExtractSymbols` | Go func/method/type, Python class/def, TS interface/function regex extraction | PASS |
| `TestGitignore` | basename `*.log` at any depth, negation `!`, anchored `/rooted`, dir `build/` | PASS |
| `TestBuildEndToEnd` | temp fixture: 2 kept + test/vendor/node_modules gated, ref-based ranking (`lib/util.go` ranks over `main.go`), budget truncation | PASS |

## Manual smoke tests (all passed)

| Command | Expected | Result |
|---------|----------|--------|
| `repomap .` | ranked tree ≤ 4000 tokens, truncation footer | ✓ 3 files, 3880 tokens |
| `repomap --json .` | valid JSON `{files, total_tokens, truncated}` | ✓ parsed |
| `repomap --index .` | self-contained HTML, SEO meta, client search | ✓ 2.9KB |
| `repomap --full .` | all files, no truncation | ✓ 27 files (rustygrep) |
| `repomap --budget 2000 /tmp/rustygrep` | ranked, refs resolved (`walker.rs` 6 refs) | ✓ |

## Defects found & fixed (atomic, same release)

1. Symbol regexes had no `(?m)` → matched string-start only, missed
   multiline bodies. **Fixed.**
2. Go method receiver `(s *Srv)` captured as symbol. Split method pattern.
   **Fixed.**
3. Gitignore `*.log` didn't match `logs/a.log` (basename depth). **Fixed.**
4. Ref counting used self-imports; now counts inbound. `--full` applied
   budget; now skips. Flags after positional arg broke parsing; documented
   Go flag convention. **All fixed.**

## Known limitations (v0.1.0, ponytail-acknowledged)

- Regex symbol extraction ≈80% accurate; tree-sitter is the upgrade path.
- Token estimate = bytes/4, heuristic.
- Gitignore supports the 90% subset; full git semantics deferred.
- No remote/API mode; no cross-repo graph.
