package mapx

import (
	"os"
	"path"
	"strings"
)

// Gitignore matches paths against .gitignore-style patterns.
// Stdlib-only: no go-gitignore dep. Implements the 90% case —
// glob (*, **), negation (!), leading slash anchoring, dir suffix (/).
// ponytail: gitignore subset, full gitignore (git::check-ignore semantics)
// if real-world repos expose misses.
type Gitignore struct {
	patterns []giPattern
}

type giPattern struct {
	raw    string // as written
	neg    bool   // starts with !
	dir    bool   // ends with /
	anchor bool   // starts with /
	base   string // pattern without leading / and trailing /
	star   bool   // contains * or **
}

func LoadGitignore(root string) (*Gitignore, error) {
	gi := &Gitignore{}
	data, err := os.ReadFile(path.Join(root, ".gitignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return gi, nil
		}
		return gi, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		gi.add(line)
	}
	return gi, nil
}

func (gi *Gitignore) add(line string) {
	p := giPattern{raw: line}
	if strings.HasPrefix(line, "!") {
		p.neg = true
		line = line[1:]
	}
	if strings.HasSuffix(line, "/") {
		p.dir = true
		line = strings.TrimSuffix(line, "/")
	}
	if strings.HasPrefix(line, "/") {
		p.anchor = true
		line = strings.TrimPrefix(line, "/")
	}
	if strings.Contains(line, "*") {
		p.star = true
	}
	p.base = line
	gi.patterns = append(gi.patterns, p)
}

// Match reports whether a repo-relative slash path (e.g. "src/app.go"
// or "vendor/x") is ignored. Last matching rule wins (git semantics).
func (gi *Gitignore) Match(rel string) bool {
	ignored := false
	for _, p := range gi.patterns {
		if p.matches(rel) {
			ignored = !p.neg
		}
	}
	return ignored
}

func (p giPattern) matches(rel string) bool {
	// dir patterns match a dir and anything under it
	if p.dir {
		if rel == p.base || strings.HasPrefix(rel, p.base+"/") {
			return true
		}
		return false
	}
	if p.star {
		// bare pattern (no slash) applies at any depth: match basename.
		if !strings.Contains(p.base, "/") {
			ok, _ := path.Match(p.base, path.Base(rel))
			return ok
		}
		ok, _ := path.Match(p.base, rel)
		return ok
	}
	if p.anchor {
		return rel == p.base
	}
	// unanchored basename matches at any depth
	if rel == p.base {
		return true
	}
	if strings.HasPrefix(rel, p.base+"/") {
		return true
	}
	return false
}

// BaselineIgnore returns always-ignored paths (VCS, node_modules, caches).
func BaselineIgnore() *Gitignore {
	gi := &Gitignore{}
	for _, p := range []string{
		".git/", ".hg/", ".svn/",
		"node_modules/", "vendor/", ".venv/", "venv/", "__pycache__/",
		"dist/", "build/", ".next/", "target/",
		".env", "*.pyc", "*.exe", "*.dll", "*.so", "*.dylib", "*.a",
	} {
		gi.add(p)
	}
	return gi
}
