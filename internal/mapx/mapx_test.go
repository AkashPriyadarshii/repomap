package mapx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestExtractSymbols covers Go, Python, TS extraction.
func TestExtractSymbols(t *testing.T) {
	got := ExtractSymbols(".go", "package main\n\nfunc main() {}\nfunc serveHTTP(w, r) {}\nfunc (s *Srv) handle() {}\n")
	want := []string{"main", "serveHTTP", "handle"}
	for _, w := range want {
		if !contains(got, w) {
			t.Fatalf("missing symbol %q in %v", w, got)
		}
	}
	py := ExtractSymbols(".py", "import os\nclass Foo:\n    def bar(self):\n        pass\ndef top(): pass\n")
	if !contains(py, "Foo") || !contains(py, "bar") || !contains(py, "top") {
		t.Fatalf("python symbols = %v", py)
	}
	ts := ExtractSymbols(".ts", "export interface User {}\nexport function fetch(): void {}")
	if !contains(ts, "User") || !contains(ts, "fetch") {
		t.Fatalf("ts symbols = %v", ts)
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// TestGitignore verifies the gitignore subset.
func TestGitignore(t *testing.T) {
	gi := &Gitignore{}
	for _, l := range []string{"node_modules/", "*.log", "!.keep.log", "/rooted", "build/", "dist/**/out"} {
		gi.add(l)
	}
	cases := []struct{ p, want string }{
		{"node_modules/a/b.js", "ignored"},
		{"x/node_modules/y.js", "ignored"}, // unanchored dir: any depth
		{"src/app.js", "keep"},
		{"debug.log", "ignored"},
		{"logs/app.log", "ignored"},
		{"logs/.keep.log", "kept"}, // negation wins at depth
		{"rooted", "ignored"},
		{"x/rooted", "keep"}, // anchored: only root
		{"build/x.js", "ignored"},
		{"a/build/x.js", "ignored"}, // baseline-anchored? no: build/ unanchored
		{"dist/out", "ignored"},     // ** matches zero dirs
		{"dist/a/b/out", "ignored"}, // ** crosses dirs
		{"dist/a/x.js", "keep"},
	}
	for _, c := range cases {
		if gi.Match(c.p) != (c.want == "ignored") {
			t.Errorf("Match(%q) wrong: got %v, want %s", c.p, gi.Match(c.p), c.want)
		}
	}
}

// TestBuildEndToEnd walks a temp fixture, checks ranking + ignore + budget.
func TestBuildEndToEnd(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.go", "package main\nimport \"example.com/x/lib\"\nfunc main(){}\n")
	write("lib/lib.go", "package lib\nfunc Helper(){}\n")
	write("lib/test_util_test.go", "package lib\nfunc TestX(){}\n")
	write("node_modules/x/index.js", "module.exports=1\n")
	write("vendor/y.go", "package y\n")
	write(".gitignore", "vendor/\n")

	res, err := Build(context.Background(), root, mustGI(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Fatalf("expected 2 files (main.go, lib/lib.go), got %d: %v", len(res.Files), paths(res))
	}
	// main.go imports example.com/x/lib -> lib/lib.go gets 1 ref, ranks first.
	if res.Files[0].Path != "lib/lib.go" {
		t.Fatalf("expected lib/lib.go first by refs, got %v (refs=%v)", paths(res), refs(res))
	}
	if res.Files[0].Refs != 1 {
		t.Fatalf("expected lib/lib.go refs=1, got %v", refs(res))
	}
	// Budget truncates.
	orig := len(res.Files)
	res.Budget(1)
	if len(res.Files) >= orig {
		t.Fatalf("budget 1 should truncate from %d files", orig)
	}
}

// TestBudgetFit keeps everything when all files fit.
func TestBudgetFit(t *testing.T) {
	r := &Result{Files: []File{{Path: "a.go", Tokens: 3}, {Path: "b.go", Tokens: 4}}}
	r.Budget(4000)
	if len(r.Files) != 2 || r.Truncated != 0 || r.TotalTokens != 7 {
		t.Fatalf("fit-case emptied: %+v", r)
	}
}

func mustGI(t *testing.T, root string) *Gitignore {
	t.Helper()
	gi, err := LoadGitignore(root)
	if err != nil {
		t.Fatal(err)
	}
	return gi
}

func paths(r *Result) []string {
	out := make([]string, len(r.Files))
	for i, f := range r.Files {
		out[i] = f.Path
	}
	return out
}

func refs(r *Result) []int {
	out := make([]int, len(r.Files))
	for i, f := range r.Files {
		out[i] = f.Refs
	}
	return out
}
