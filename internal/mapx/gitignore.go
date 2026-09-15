package mapx

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Gitignore matches paths against .gitignore-style patterns.
// Stdlib-only: no go-gitignore dep. Implements glob (*, **, ?),
// negation (!), leading slash anchoring, dir suffix (/).
// Unanchored names match basenames at any depth (git semantics).
// ponytail: nested .gitignore files + escaped chars (!) if real-world
// repos expose misses.
type Gitignore struct {
	patterns []giPattern
}

type giPattern struct {
	raw    string // as written
	neg    bool   // starts with !
	dir    bool   // ends with /
	anchor bool   // starts with / or contains / mid-pattern
	base   string // pattern without leading / and trailing /
	re     *regexp.Regexp
}

func LoadGitignore(root string) (*Gitignore, error) {
	gi := &Gitignore{}
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
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
	} else if strings.Contains(line, "/") {
		// git: a slash anywhere but trailing anchors the pattern to root
		p.anchor = true
	}
	p.base = line
	p.re = globToRe(line, p.anchor)
	gi.patterns = append(gi.patterns, p)
}

// globToRe compiles a gitignore glob to a regexp over slash paths:
// **/ matches zero+ dirs, ** matches anything, * matches non-/ run.
func globToRe(pat string, anchored bool) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	if !anchored {
		b.WriteString("(?:.*/)?")
	}
	i := 0
	for i < len(pat) {
		switch pat[i] {
		case '*':
			if i+1 < len(pat) && pat[i+1] == '*' {
				// ** or **/
				if i+2 < len(pat) && pat[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 3
				} else {
					b.WriteString(".*")
					i += 2
				}
			} else {
				b.WriteString("[^/]*")
				i++
			}
		case '?':
			b.WriteString("[^/]")
			i++
		default:
			b.WriteString(regexp.QuoteMeta(string(pat[i])))
			i++
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return regexp.MustCompile("^$")
	}
	return re
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
	if p.dir {
		// dir itself or anything under it; unanchored matches at any depth
		if p.anchor {
			if rel == p.base || strings.HasPrefix(rel, p.base+"/") {
				return true
			}
			return false
		}
		for i := range rel {
			if i > 0 && rel[i-1] != '/' {
				continue
			}
			rest := rel[i:]
			if rest == p.base || strings.HasPrefix(rest, p.base+"/") {
				return true
			}
			if p.re.MatchString(rest) || p.re.MatchString(path.Base(rest)) {
				return true
			}
		}
		return p.re.MatchString(rel) || p.re.MatchString(path.Base(rel))
	}
	if p.anchor {
		return p.re.MatchString(rel)
	}
	return p.re.MatchString(rel) || p.re.MatchString(path.Base(rel))
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
