# repomap site — DESIGN.md

Single static page. `site/index.html`. System fonts only, one inline `<style>`,
one inline `<script>` (copy buttons + budget toggle). No framework, no external
fonts, no build step. Open `site/index.html` directly in a browser.

## Tokens

- Accent emerald: `--acc: #059669`, bright `#34d399` (matches `assets/logo.svg`).
- Canvas paper `#f4f6f4`, ink `#0c1512`, muted `#5a6b63`.
- Dark terminal block `#0a1610` bg, `#d7e5dd` text, emerald prompt.
- Display font: system stack. Mono: ui-monospace stack.
- No gradients on light surface. No shadows. Hairline borders `#d4ddd8`.

## Layout

Nav (sticky, hairline) → hero (one-liner + install cmd + logo) →
stats strip (4 numbers, real) → live map demo (budget toggle 2000/full) →
why (3 bullets, no cards grid) → how it ranks (weights table) →
flags table → JSON block → architecture tree → trilogy strip →
footer (Ecosystem / Author / Social).

## Real numbers (verified 2026-09-15, `repomap.exe --full .`)

- Binary 3.9MB, 14 files mapped, 13149 tokens full, top file walk.go 2022 tokens.
- Paste real output in demo. Never invent. Re-verify with
  `repomap.exe --full . | tail -2` before editing numbers.

## JS

Copy buttons (`data-copy` → navigator.clipboard, fallback select).
Budget toggle swaps two `<pre>` blocks (2000 vs full). Nothing else.

## Anti-slop

No centered hero + 3 cards. No purple gradient. No Inter. Stats are a
strip, not cards. `prefers-reduced-motion` respected (no animation anyway).
