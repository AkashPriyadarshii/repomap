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

// TestGitignore verifies the 90% gitignore subset.
func TestGitignore(t *testing.T) {
	gi := &Gitignore{}
	for _, l := range []string{"node_modules/", "*.log", "!.keep.log", "/rooted", "build/"} {
		gi.add(l)
	}
	cases := []struct{ p, want string }{
		{"node_modules/a/b.js", "ignored"},
		{"src/app.js", "keep"},
		{"debug.log", "ignored"},
		{"logs/app.log", "ignored"},
		{".keep.log", "kept"}, // negation wins
		{"rooted", "ignored"},
		{"x/rooted", "keep"}, // anchored: only root
		{"build/x.js", "ignored"},
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
	write("main.go", "package main\nfunc main(){}\nfunc run(){}\n")
	write("lib/util.go", "package lib\nfunc Helper(){}\nfunc More(){}\n")
	write("lib/test_util_test.go", "package lib\nfunc TestX(){}\n")
	write("node_modules/x/index.js", "module.exports=1\n")
	write("vendor/y.go", "package y\n")
	write(".gitignore", "vendor/\n")

	res, err := Build(context.Background(), root, mustGI(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Fatalf("expected 2 files (main.go, lib/util.go), got %d: %v", len(res.Files), paths(res))
	}
	// main.go references lib/util.go -> util should rank higher.
	if res.Files[0].Path != "lib/util.go" {
		t.Fatalf("expected lib/util.go first by refs, got %v", paths(res))
	}
	// Budget truncates.
	orig := len(res.Files)
	res.Budget(1)
	if len(res.Files) >= orig {
		t.Fatalf("budget 1 should truncate from %d files", orig)
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
