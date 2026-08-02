# Changelog

All notable changes to repomap.

## [v0.1.0] - 2026-08-02

### Added

- **Repo walk** honoring `.gitignore` + baseline ignores (VCS, node_modules,
  vendor, caches) — skip, not prune.
- **Symbol extraction** — regex-based for Go, Python, JS/TS/JSX/TSX, Rust,
  Zig, Kotlin, PowerShell, Shell, Ruby, PHP. Reads first 200 lines.
- **Import-ref ranking** — files imported by others score higher
  (0.45 refs / 0.25 symbols / 0.15 size / 0.05 name).
- **Token budget** — `--budget N` (default 4000), truncates tail, reports
  `… N files truncated`.
- **Output modes** — tree, `--json`, `--index` (self-contained searchable
  HTML with SEO meta + client-side search, no server), `--full`, `-o`.
- **Concurrency** — parallel symbol extraction across 8 workers.
- **FOSS** — MIT license, credited to Akash Priyadarshi.

### Fixed (pre-release, same commit)

- Symbol regexes anchored per-line (`(?m)`), method receivers excluded.
- Gitignore star patterns match basename at any depth (`*.log` → `logs/a.log`).
- Ref counting counts *inbound* imports, not self-imports.
- `--full` skips budget truncation (means *everything*).
- Flag parsing: flags precede positional path (Go `flag` convention).

### Notes

- Ranker is heuristic (`ponytail:` ceilings documented in design.md). Good
  enough to orient; swap to tree-sitter only if real repos expose misses.
- Token estimate = bytes/4, close enough for a budget.
