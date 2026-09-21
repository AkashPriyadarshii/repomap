<!--
  =============================================================================
  SEO METADATA & KEYWORD INDEX
  =============================================================================
  Title: repomap | Token-Budgeted Repo Map for AI Coding Agents in Go
  Author: Akash Priyadarshi (@AkashPriyadarshii)
  Description: repomap walks any repo, ranks files by import refs and symbols,
  and emits a token-capped map so AI coding agents orient before they read or
  search. Stdlib-only Go, zero dependencies, tree/JSON/HTML output. Third of
  the read-search-orient trilogy with zcat and rustygrep.
  Keywords: repomap, repo map, code map, repository map, go cli, token budget,
  ai agents, llm tools, coding agents, code search, import refs, symbol
  extraction, ranked files, gitignore aware, json output, html index, zcat,
  rustygrep, golang cli, developer tools, codebase orientation.
  =============================================================================
-->

<div align="center">

<img src="assets/logo.svg" width="96" height="96" alt="repomap logo — emerald rm monogram over ranked-file bars and a token-budget cap">

# repomap

**Repo map for AI coding agents. What is in this repo, in under 4000 tokens.**

[![Go](https://img.shields.io/badge/go-1.26-00ADD8.svg?style=flat-square&logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/AkashPriyadarshii/repomap?style=flat-square)](https://github.com/AkashPriyadarshii/repomap/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/AkashPriyadarshii/repomap.svg?style=flat-square)](https://pkg.go.dev/github.com/AkashPriyadarshii/repomap)
[![Binary](https://img.shields.io/badge/binary-3.9MB-0f1114.svg?style=flat-square)](#)
[![Deps](https://img.shields.io/badge/deps-stdlib%20only-8A2BE2.svg?style=flat-square)](#)

**by [Akash Priyadarshi](https://github.com/AkashPriyadarshii)** · Patna, Bihar, India

Third of a trilogy:
[zcat](https://github.com/AkashPriyadarshii/zcat) (read a file) →
[rustygrep](https://github.com/AkashPriyadarshii/rustygrep) (find a needle) →
**repomap** (orient: where am I, what is here).

[Quickstart](#-quickstart) · [Why](#-why) · [How it ranks](#-how-it-ranks) · [Flags](#-flags) · [JSON](#-json-mode) · [Architecture](#-architecture)

</div>

[![crates.io](https://img.shields.io/crates/v/repomap?style=flat-square)](https://crates.io/crates/repomap) [![downloads](https://img.shields.io/crates/d/repomap?style=flat-square)](https://crates.io/crates/repomap) [![release](https://img.shields.io/github/v/release/AkashPriyadarshii/repomap?style=flat-square&label=release)](https://github.com/AkashPriyadarshii/repomap/releases)

---

## Why

`tree` dumps 5000 lines and eats your context window before you read one file.
An agent needs orientation first: which files matter, what they export, who
imports them. repomap walks the repo, ranks files by inbound import refs, and
cuts output at a token budget. You see the important files before you spend
tokens reading them.

- **Ranked, not listed.** Files imported by others float up. Entry points sink.
- **Token-capped.** Default 4000 tokens. Tail truncates with a count, never silently.
- **Symbols inline.** Top declarations per file, first 200 lines, regex-based.
- **Gitignore-aware.** Respects `.gitignore` plus baseline ignores (VCS, caches, vendor).
- **Zero dependencies.** Stdlib Go only. One static binary.

## Quickstart

```bash
go install github.com/AkashPriyadarshii/repomap/cmd/repomap@latest
# or build locally
go build -o repomap ./cmd/repomap
```

```bash
repomap                  # ranked tree, token-capped (default 4000)
repomap --budget 2000    # cap output
repomap --full           # no truncation, includes tests/vendor
repomap --json           # machine-readable
repomap --index -o index.html  # searchable HTML page, no server
repomap ./path/to/repo   # map a specific dir
```

Example on this repo:

```
$ repomap --budget 2000
└── internal/
    └── mapx/
        └── walk.go	   1 refs	2022 tokens	[supportedLang, scoreFile, clamp, firstLinesBytes, ...]

… 13 files truncated, --full to see all

1 files, 2022 tokens
```

## How it ranks

Score per file: `0.45 × refs + 0.25 × symbols + 0.15 × size + 0.05 × symbol density`.
A file is important when other files import it.

- **Refs:** inbound import edges. Go `import (...)` blocks, Python `import`/`from`,
  JS/TS `import`/`require`, Rust `use`, Ruby `require`.
- **Symbols:** declaration count from the first 200 lines. Go, Python, JS/TS/JSX/TSX,
  Rust, Zig, Kotlin, PowerShell, Shell, Ruby, PHP.
- **Budget:** keeps top-N by score until the token cap, always keeps rank 1.

Token estimate is bytes/4. Close enough for a budget.

## Flags

| Flag | Effect |
|------|--------|
| `--budget N` | Max output tokens (default 4000, `0` = unlimited) |
| `--full` | Show all files, skip budget truncation (includes tests, vendor, generated) |
| `--json` | Machine-readable output (same rank order) |
| `--index` | Self-contained searchable HTML page (SEO meta + client-side search) |
| `--index-title S` | Title for `--index` page (default `"repomap"`) |
| `-o FILE` | Write output to file instead of stdout |
| `--no-gitignore` | Include ignored files (warning: huge) |
| `--version` | Print version |

## JSON mode

`--json` gives agents a machine map, not a human tree to re-parse:

```json
{
  "files": [
    {"path": "internal/mapx/walk.go", "refs": 1,
     "symbols": ["Build", "parseGoImports"], "score": 0.402, "tokens": 2022}
  ],
  "total_tokens": 2022,
  "truncated": 13
}
```

## Architecture

```
cmd/repomap/main.go    — flags, budget wiring, output routing
internal/mapx/
  walk.go              — walk + parallel symbol/import extraction + ref index
  symbols.go           — regex symbol patterns per language
  rank.go              — Budget() truncation (always keeps rank 1)
  render.go            — tree / JSON / HTML index renderers
  gitignore.go         — stdlib .gitignore subset (*, **, !, anchoring, dir/)
```

- **Walk:** `filepath.WalkDir`, skips baseline + `.gitignore` dirs early.
- **Extract:** `runtime.NumCPU()` workers, `os.ReadFile` + first-200-lines symbol scan.
- **Refs:** basename + dir-segment index, one count per (importer, target) edge.
- **Render:** ranked tree with `refs`/`tokens`/`[symbols]` per line, JSON, or HTML.

## Non-goals

- Full gitignore parity (no nested `.gitignore` files yet).
- Tree-sitter parsing. Regex orients; swap only if real repos expose misses.
- Daemon mode, watch mode, incremental maps. Run it, read it, move on.

## Test

```bash
go test ./...   # unit + end-to-end fixture (ranking, ignore, budget)
go vet ./...    # static check
```

## Docs

- [docs/PRD.md](docs/PRD.md) — product, goals, interface
- [docs/design.md](docs/design.md) — architecture, ranking, budget
- [docs/impl.md](docs/impl.md) — implementation plan
- [docs/CLAUDE.md](docs/CLAUDE.md) — project conventions

## License

MIT. See [LICENSE](LICENSE).

---

## Ecosystem & Author

### Ecosystem
- [`design-genius`](https://github.com/AkashPriyadarshii/design-genius)
- [`akash-design-engineering`](https://github.com/AkashPriyadarshii/akash-design-engineering)
- [`tdlib-android`](https://github.com/AkashPriyadarshii/tdlib-android)
- [`kharcha`](https://github.com/AkashPriyadarshii/kharcha)

### Author
- **Akash Priyadarshi** (Patna, Bihar, India)
- [GitHub](https://github.com/AkashPriyadarshii) · [Portfolio](https://akashpriyadarshi.vercel.app) · [LinkedIn](https://linkedin.com/in/akash-priyadarshi-1aa51b37a) · [Resume](https://akashpriyadarshii.github.io/Resume/)

### Social
- [X / Twitter](https://x.com/Akash__ydv001) · [Threads](https://www.threads.com/@free_dev2026) · [Instagram](https://www.instagram.com/akash.priyadarshii/) · [Reddit](https://reddit.com/user/DragonfruitWeak2801)

---

<div align="center">
  <b>Read with zcat. Search with rustygrep. Orient with repomap.</b><br>
  <i>Keywords: repomap, repo map, code map, token budget, AI coding agents, import refs, golang cli.</i>
</div>
