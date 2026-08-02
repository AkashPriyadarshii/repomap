package mapx

import (
	"regexp"
	"strings"
)

// symbolRe describes how to find named declarations in a language.
type symbolRe struct {
	re   *regexp.Regexp
	lang string
}

var symbolPatterns = []symbolRe{
	// Go: methods first (have receiver), then funcs, types, const/var.
	// Method regex: func (RECV) NAME( — capture NAME.
	{regexp.MustCompile(`(?m)^func\s*\([^)]*\)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "go"},
	{regexp.MustCompile(`(?m)^func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "go"},
	{regexp.MustCompile(`(?m)^type\s+([A-Za-z_][A-Za-z0-9_]*)\s`), "go"},
	{regexp.MustCompile(`(?m)^(?:const|var)\s+([A-Za-z_][A-Za-z0-9_]*)\s`), "go"},
	// Python
	{regexp.MustCompile(`(?m)^\s*def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "python"},
	{regexp.MustCompile(`(?m)^\s*class\s+([A-Za-z_][A-Za-z0-9_]*)\b`), "python"},
	{regexp.MustCompile(`(?m)^\s*async\s+def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`), "python"},
	// JS / TS / JSX / TSX
	{regexp.MustCompile(`(?m)^(?:export\s+)?(?:default\s+)?(?:function|class|interface|type)\s+([A-Za-z_$][A-Za-z0-9_$]*)`), "js"},
	{regexp.MustCompile(`(?m)^\s*(?:export\s+)?const\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:async\s*)?\(`), "js"},
	// Rust
	{regexp.MustCompile(`(?m)^fn\s+([a-z_][a-z0-9_]*)\s*[<(]`), "rust"},
	{regexp.MustCompile(`(?m)^(?:pub\s+)?(?:struct|enum|trait|impl)\s+([A-Za-z_][A-Za-z0-9_]*)`), "rust"},
	// Zig
	{regexp.MustCompile(`(?m)^pub?\s+fn\s+([a-zA-Z_][a-zA-Z0-9_]*)`), "zig"},
	// Kotlin
	{regexp.MustCompile(`(?m)^(?:fun|class|interface|object)\s+([A-Za-z_][A-Za-z0-9_]*)`), "kotlin"},
	// PowerShell
	{regexp.MustCompile(`(?m)^function\s+([A-Za-z_][A-Za-z0-9_-]*)`), "powershell"},
	// Shell / Bash
	{regexp.MustCompile(`(?m)^([a-zA-Z_][a-zA-Z0-9_]*)\s*\(\s*\)\s*\{`), "shell"},
	// Ruby
	{regexp.MustCompile(`(?m)^(?:def|class|module)\s+([A-Za-z_][A-Za-z0-9_]*[!?]?)`), "ruby"},
	// PHP
	{regexp.MustCompile(`(?m)^(?:function|class|interface|trait)\s+([A-Za-z_][A-Za-z0-9_]*)`), "php"},
}

// langByExt maps file extensions to a language family.
var langByExt = map[string]string{
	".go": "go", ".py": "python", ".js": "js", ".ts": "js", ".jsx": "js", ".tsx": "js",
	".rs": "rust", ".zig": "zig", ".kt": "kotlin", ".kts": "kotlin", ".ps1": "powershell",
	".sh": "shell", ".bash": "shell", ".rb": "ruby", ".php": "php",
}

// ExtractSymbols returns symbol names found in the first maxLines of content.
func ExtractSymbols(ext, content string) []string {
	lang, ok := langByExt[strings.ToLower(ext)]
	if !ok {
		return nil
	}
	var out []string
	for _, sr := range symbolPatterns {
		if sr.lang != lang {
			continue
		}
		for _, m := range sr.re.FindAllStringSubmatch(content, -1) {
			if len(m) > 1 && m[1] != "" {
				out = append(out, m[1])
			}
		}
	}
	return dedupe(out)
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0]
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

var importPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^\s*(?:import|from)\s+["']([^"']+)["']`), // python, js, ts
	regexp.MustCompile(`(?m)^\s*import\s+"([^"]+)"`),                 // go
	regexp.MustCompile(`(?m)^\s*use\s+([a-zA-Z_:][a-zA-Z0-9_:]*)\s*[;{]"`), // rust/php
	regexp.MustCompile(`(?m)require\(['"]([^'"]+)['"]\)`),            // node
	regexp.MustCompile(`(?m)^\s*using\s+([a-zA-Z_][a-zA-Z0-9_.]*);`), // c#/php
}

// importSegs splits an import token into candidate segments for ref
// matching: last path chunk, plus for Rust "crate::a::b" the module names.
// Keeps only word-ish segments (no "crate", "super", bare letters).
func importSegs(tok string) []string {
	tok = strings.Trim(tok, `"'`)
	tok = strings.TrimPrefix(tok, "./")
	tok = strings.TrimPrefix(tok, "/")
	tok = strings.TrimPrefix(tok, "file://")
	// strip version/scope noise
	if i := strings.IndexAny(tok, "@^~"); i >= 0 {
		tok = tok[:i]
	}
	if tok == "" {
		return nil
	}
	sep := "/"
	parts := strings.Split(tok, sep)
	last := strings.TrimSpace(parts[len(parts)-1])
	// handle crate::a::b
	scoped := strings.Split(last, "::")
	cands := []string{}
	for _, s := range scoped {
		s = strings.TrimSpace(s)
		if s == "" || s == "crate" || s == "super" || s == "self" || s == "std" {
			continue
		}
		if len(s) == 1 {
			continue // bare drive letters, single letters: noise
		}
		cands = append(cands, s)
	}
	// for dir imports, also allow the dir prefix to match a subdir
	if len(parts) > 1 {
		dir := strings.Join(parts[:len(parts)-1], "/")
		dir = strings.Trim(dir, "/")
		if dir != "" {
			cands = append(cands, dir)
		}
	}
	return cands
}

// trimExt strips the final extension.
func trimExt(base string) string {
	if i := strings.LastIndex(base, "."); i >= 0 {
		return base[:i]
	}
	return base
}
