package mapx

import (
	"context"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// File is one mapped file.
type File struct {
	Path    string   `json:"path"`
	Size    int64    `json:"-"`
	Refs    int      `json:"refs"`
	Symbols []string `json:"symbols,omitempty"`
	Score   float64  `json:"score"`
	Tokens  int      `json:"tokens"`
}

// Result is the full map.
type Result struct {
	Files       []File `json:"files"`
	TotalTokens int    `json:"total_tokens"`
	Truncated   int    `json:"truncated"`
}

// nameGate: files whose path matches are hidden unless --full.
var nameGate = regexp.MustCompile(`(?i)(^|/)(test|tests|spec|vendor|node_modules|dist|build|generated|\.next|coverage)(/|$)|(_test\.go$|\.test\.|\.min\.js$)`)

var supportedExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
	".rs": true, ".zig": true, ".kt": true, ".kts": true, ".ps1": true, ".sh": true,
	".bash": true, ".rb": true, ".php": true, ".c": true, ".h": true, ".cpp": true,
	".hpp": true, ".java": true, ".md": true, ".html": true, ".css": true, ".json": true,
	".yaml": true, ".yml": true, ".toml": true, ".sql": true, ".dockerfile": true,
	".ex": true, ".exs": true, ".swift": true, ".cs": true, ".vue": true, ".svelte": true,
}

func supportedLang(ext string) bool { return supportedExts[ext] }

// scoreFile ranks importance for an LLM: refs > symbols > size.
func scoreFile(size int64, refs, syms int) float64 {
	score := 0.45*clamp(float64(refs)/30) +
		0.25*clamp(math.Log(float64(syms)+1)/3) +
		0.15*clamp(math.Log(float64(size)+1)/10) +
		0.05*clamp(float64(syms)/20)
	return math.Round(score*1000) / 1000
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func firstLines(s string, n int) string {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
			if count >= n {
				return s[:i]
			}
		}
	}
	return s
}

// parseImports extracts module-ish tokens from import lines.
func parseImports(ext, content string) []string {
	var out []string
	impRe := importReFor(ext)
	if impRe == nil {
		return nil
	}
	for _, m := range impRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && m[1] != "" {
			out = append(out, m[1])
		}
	}
	return out
}

func importReFor(ext string) *regexp.Regexp {
	switch strings.ToLower(ext) {
	case ".py":
		return regexp.MustCompile(`(?m)^\s*(?:import|from)\s+([A-Za-z_][A-Za-z0-9_.]*)`)
	case ".js", ".ts", ".jsx", ".tsx":
		return regexp.MustCompile(`(?m)(?:import|require)\s*\(?\s*['"]([^'"]+)['"]`)
	case ".go":
		return regexp.MustCompile(`(?m)^\s*import\s+["]?([^"\s)]+)["]?`)
	case ".rs":
		return regexp.MustCompile(`(?m)^\s*use\s+([A-Za-z_][A-Za-z0-9_:]*)`)
	case ".rb":
		return regexp.MustCompile(`(?m)^\s*require\s+['"]([^'"]+)['"]`)
	default:
		return nil
	}
}

// Build walks root, extracts symbols, counts refs, ranks, applies name gate.
func Build(ctx context.Context, root string, gitignore *Gitignore, full bool) (*Result, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	baseline := BaselineIgnore()

	var mu sync.Mutex
	byPath := map[string]*File{}
	imports := map[string][]string{}
	refsByPath := map[string]int{}

	// Pass 1: walk, collect supported files.
	var paths []string
	walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return nil
		}
		relS := filepath.ToSlash(rel)
		if d.IsDir() {
			if relS == "." {
				return nil
			}
			if baseline.Match(relS) || gitignore.Match(relS) {
				return filepath.SkipDir
			}
			return nil
		}
		if baseline.Match(relS) || gitignore.Match(relS) {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !supportedLang(ext) {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return nil
		}
		mu.Lock()
		byPath[relS] = &File{Path: relS, Size: info.Size()}
		paths = append(paths, relS)
		mu.Unlock()
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Pass 2: extract symbols + imports, parallel.
	fileCh := make(chan string)
	var wg sync.WaitGroup
	ctxWork, cancel := context.WithCancel(ctx)
	defer cancel()
	workers := 8
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for relS := range fileCh {
				select {
				case <-ctxWork.Done():
					return
				default:
				}
				abs := filepath.Join(root, filepath.FromSlash(relS))
				data, rerr := os.ReadFile(abs)
				if rerr != nil {
					continue
				}
				ext := strings.ToLower(filepath.Ext(abs))
				syms := ExtractSymbols(ext, firstLines(string(data), 200))
				mods := parseImports(ext, string(data))
				mu.Lock()
				if f, ok := byPath[relS]; ok {
					f.Symbols = syms
				}
				imports[relS] = mods
				mu.Unlock()
			}
		}()
	}
	go func() {
		defer close(fileCh)
		for _, p := range paths {
			select {
			case <-ctxWork.Done():
				return
			case fileCh <- p:
			}
		}
	}()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}

	// Ref counting: a file is important when OTHER files import it.
	// Import token ("internal/mapx", "crate::walker", "walker", "fs") →
	// candidate segments (walk a-z, path-ish chunks). A candidate matches
	// a target file whose basename equals it, or a dir prefix equal to it.
	// Heuristic; a map orients, it doesn't reason.
	for src, mods := range imports {
		for _, m := range mods {
			for _, seg := range importSegs(m) {
				for _, f := range byPath {
					if f.Path == src {
						continue
					}
					base := trimExt(filepath.Base(f.Path))
					if base == seg || strings.HasPrefix(f.Path, seg+"/") {
						refsByPath[f.Path]++
						break
					}
				}
			}
		}
	}

	// Finalize.
	res := &Result{}
	for _, f := range byPath {
		f.Refs = refsByPath[f.Path]
		f.Tokens = EstimateTokens(f.Size)
		f.Score = scoreFile(f.Size, f.Refs, len(f.Symbols))
		if !full && nameGate.MatchString(f.Path) {
			continue
		}
		res.Files = append(res.Files, *f)
	}
	sort.Slice(res.Files, func(i, j int) bool {
		if res.Files[i].Score != res.Files[j].Score {
			return res.Files[i].Score > res.Files[j].Score
		}
		return res.Files[i].Path < res.Files[j].Path
	})
	res.TotalTokens = sumTokens(res.Files)
	return res, nil
}

func sumTokens(files []File) int {
	n := 0
	for _, f := range files {
		n += f.Tokens
	}
	return n
}

// EstimateTokens: ~4 chars/token.
func EstimateTokens(size int64) int {
	return int(size/4) + 1
}
