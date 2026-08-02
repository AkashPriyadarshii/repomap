# repomap — PRD

## Product

Third tool in the "read → search → orient" trilogy for AI coding agents:

| Tool | Answer | Status |
|------|--------|--------|
| zcat | read a file | shipped |
| rustygrep | find a needle | shipped |
| **repomap** | orient: what's in this repo | **this repo** |

**Problem:** real repos have thousands of files. `tree` output dwarfs an
LLM context window before the agent reads one file, and a 5000-line listing
is useless — the agent wants "where is the auth code", not "here are every
vendor dir and node_modules".

**Job:** answer "what does this repo contain" in the fewest tokens.

## Goals (v0.1.0)

1. Walk a repo, honoring `.gitignore`.
2. Extract per-file symbols (functions, classes, types) via regex.
3. Rank files by importance for an agent.
4. Emit a token-budgeted tree, or machine-readable JSON.
5. One static binary, stdlib only, no runtime deps.

## Non-goals (v0.1.0)

- Not a code search engine (that's rustygrep).
- Not an RAG store / embeddings.
- No tree-sitter parsing (regex extractors first — see design.md).
- No remote/API mode.

## Interface

```
repomap                  # map, ranked, token-capped
repomap --budget 2000    # cap output (default 4000)
repomap --full           # no truncation
repomap --json           # machine-readable
repomap --no-gitignore   # include ignored files (warn: huge)
```

`--json` is the money feature — same pattern as zcat/rustygrep. An agent
gets a machine-readable map, not a human tree to re-parse.

## Success criteria (v0.1.0)

- `repomap` on this repo prints a ranked tree, tokens ≤ budget.
- `repomap --json` emits `{files:[...], total_tokens}` parseable as JSON.
- Ignores `.gitignore`-excluded paths; `.git`, `node_modules`, vendor gone.
- `go test ./...` passes.
- Binary runs on Windows (this dev box) via `go build`.
