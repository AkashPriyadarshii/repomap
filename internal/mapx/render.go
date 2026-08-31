package mapx

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
)

// RenderTree returns a token-budgeted tree. Files already ranked by Build.
func (r *Result) RenderTree() string {
	var b strings.Builder
	root := &treeNode{name: ".", dir: true, children: map[string]*treeNode{}}
	for _, f := range r.Files {
		parts := strings.Split(f.Path, "/")
		cur := root
		for i, part := range parts {
			if i == len(parts)-1 {
				n := &treeNode{name: part, file: f}
				cur.add(n)
				break
			}
			child, ok := cur.children[part]
			if !ok {
				child = &treeNode{name: part, dir: true, children: map[string]*treeNode{}}
				cur.add(child)
			}
			cur = child
		}
	}
	// Draw.
	drawTree(&b, root, "", true)
	if r.Truncated > 0 {
		fmt.Fprintf(&b, "\n… %d files truncated, --full to see all\n", r.Truncated)
	}
	fmt.Fprintf(&b, "\n%d files, %d tokens\n", len(r.Files), r.TotalTokens)
	return b.String()
}

type treeNode struct {
	name     string
	dir      bool
	file     File
	children map[string]*treeNode
}

func (n *treeNode) add(c *treeNode) {
	if n.children == nil {
		n.children = map[string]*treeNode{}
	}
	n.children[c.name] = c
}

// drawTree walks the node tree, printing dirs then files.
func drawTree(b *strings.Builder, n *treeNode, prefix string, isLast bool) {
	// root prints nothing but recurses children.
	if n.name == "." {
		for _, ch := range sortedNodes(n.children) {
			drawTree(b, ch, "", true)
		}
		return
	}
	conn := "├── "
	if isLast {
		conn = "└── "
	}
	b.WriteString(prefix + conn + n.name)
	if n.dir {
		b.WriteString("/\n")
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
		for _, ch := range sortedNodes(n.children) {
			drawTree(b, ch, childPrefix, ch == lastNode(n.children))
		}
		return
	}
	// File line: name + refs + symbols.
	fmt.Fprintf(b, "\t%4d refs\t%4d tokens", n.file.Refs, n.file.Tokens)
	if len(n.file.Symbols) > 0 {
		b.WriteString("\t[" + strings.Join(n.file.Symbols, ", ") + "]")
	}
	b.WriteString("\n")
}

func sortedNodes(m map[string]*treeNode) []*treeNode {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]*treeNode, 0, len(names))
	for _, k := range names {
		out = append(out, m[k])
	}
	return out
}

func lastNode(m map[string]*treeNode) *treeNode {
	ns := sortedNodes(m)
	if len(ns) == 0 {
		return nil
	}
	return ns[len(ns)-1]
}

// RenderJSON marshals the result.
func (r *Result) RenderJSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RenderIndex emits a self-contained HTML page with the map embedded as
// JSON plus SEO/meta/robots — indexable, searchable, no server needed.
func (r *Result) RenderIndex(title string) string {
	jsonData, _ := json.Marshal(r)
	// Build a search index: path → symbols.
	esc := func(s string) string { return html.EscapeString(s) }
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + esc(title) + ` — repo map for AI agents</title>
<meta name="description" content="` + esc(title) + ` repo map: files, symbols, import refs ranked for LLM coding agents. Token-budgeted, machine-readable, searchable. Part of the read→search→orient trilogy (zcat, rustygrep, repomap).">
<meta name="robots" content="index, follow, max-snippet:-1, max-image-preview:large">
<meta name="author" content="Akash Priyadarshi">
<meta name="keywords" content="repomap, repo map, code map, AI agent, LLM, coding agent, code search, symbols, import refs, zcat, rustygrep, golang">
<meta name="theme-color" content="#060D0A">
<link rel="canonical" href="https://github.com/AkashPriyadarshii/repomap">
<meta property="og:type" content="website">
<meta property="og:url" content="https://github.com/AkashPriyadarshii/repomap">
<meta property="og:site_name" content="repomap">
<meta property="og:title" content="` + esc(title) + ` — repo map for AI agents">
<meta property="og:description" content="What's in this repo in ≤budget tokens. Ranked files, symbols, import refs for LLM agents.">
<meta property="og:image" content="https://raw.githubusercontent.com/AkashPriyadarshii/repomap/master/docs/repomap.png">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="` + esc(title) + ` — repo map for AI agents">
<meta name="twitter:description" content="What's in this repo in ≤budget tokens. Ranked files, symbols, import refs for LLM agents.">
<meta name="twitter:image" content="https://raw.githubusercontent.com/AkashPriyadarshii/repomap/master/docs/repomap.png">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Fragment+Mono:ital@0;1&family=Syne:wght@400;600;700;800&display=swap" rel="stylesheet">
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "repomap",
  "alternateName": "` + esc(title) + `",
  "description": "Repo map for AI coding agents: what's in this repo in ≤budget tokens.",
  "applicationCategory": "DeveloperApplication",
  "operatingSystem": "macOS, Linux, Windows",
  "softwareVersion": "0.1.0",
  "license": "https://opensource.org/licenses/MIT",
  "url": "https://github.com/AkashPriyadarshii/repomap",
  "author": {"@type": "Person", "name": "Akash Priyadarshi"},
  "offers": {"@type": "Offer", "price": "0", "priceCurrency": "USD"}
}
</script>
<style>
  :root {
    --bg: #060D0A;
    --panel: #0b1712;
    --panel-border: #132b20;
    --panel-hover: #0f241a;
    --ink: #e2ede7;
    --dim: #7da593;
    --faint: #345747;
    --accent: #10B981;
    --accent-hover: #34D399;
    --accent-deep: #047857;
    --accent-glow: rgba(16, 185, 129, 0.15);
    --font-display: 'Syne', sans-serif;
    --font-mono: 'Fragment Mono', ui-monospace, Menlo, monospace;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    background: var(--bg);
    color: var(--ink);
    font-family: var(--font-mono);
    font-size: 15px;
    line-height: 1.6;
    max-width: 920px;
    margin: 0 auto;
    padding: 3rem 1.5rem;
  }
  ::selection {
    background: var(--accent);
    color: var(--bg);
  }
  header {
    margin-bottom: 2rem;
    border-bottom: 1px solid var(--panel-border);
    padding-bottom: 1.5rem;
  }
  h1 {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.6rem, 4vw, 2.2rem);
    letter-spacing: -0.02em;
    color: var(--ink);
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }
  .brand-badge {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    font-weight: 400;
    color: var(--accent);
    background: rgba(16, 185, 129, 0.1);
    border: 1px solid var(--accent-deep);
    padding: 0.2rem 0.55rem;
    border-radius: 4px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }
  .subtitle {
    color: var(--dim);
    font-size: 0.9rem;
    margin-top: 0.5rem;
  }
  .search-wrap {
    position: relative;
    margin-bottom: 1.5rem;
  }
  input#q {
    width: 100%;
    padding: 0.85rem 1.1rem;
    font-family: var(--font-mono);
    font-size: 0.95rem;
    background: var(--panel);
    color: var(--ink);
    border: 1px solid var(--panel-border);
    border-radius: 8px;
    outline: none;
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  input#q:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }
  input#q::placeholder {
    color: var(--faint);
  }
  .summary {
    color: var(--dim);
    font-size: 0.85rem;
    margin-bottom: 1rem;
    padding: 0.25rem 0;
  }
  ul.results {
    list-style: none;
    display: grid;
    gap: 0.5rem;
  }
  ul.results li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--panel);
    border: 1px solid var(--panel-border);
    border-radius: 6px;
    padding: 0.65rem 0.9rem;
    transition: border-color 0.15s, background 0.15s;
    gap: 1rem;
  }
  ul.results li:hover {
    background: var(--panel-hover);
    border-color: var(--accent-deep);
  }
  code {
    font-family: var(--font-mono);
    color: var(--accent-hover);
    background: rgba(16, 185, 129, 0.08);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-size: 0.88rem;
    word-break: break-all;
  }
  .meta {
    display: flex;
    gap: 0.45rem;
    flex-shrink: 0;
  }
  .badge {
    font-size: 0.72rem;
    padding: 0.15rem 0.45rem;
    border-radius: 3px;
    letter-spacing: 0.02em;
  }
  .badge-refs {
    background: rgba(4, 120, 87, 0.25);
    color: var(--accent);
    border: 1px solid var(--accent-deep);
  }
  .badge-syms {
    background: #08150f;
    color: var(--dim);
    border: 1px solid var(--panel-border);
  }
  .no-match {
    color: var(--dim);
    padding: 2rem;
    text-align: center;
    background: var(--panel);
    border: 1px dashed var(--panel-border);
    border-radius: 6px;
  }
  pre.tree-dump {
    background: #040806;
    border: 1px solid var(--panel-border);
    border-radius: 8px;
    padding: 1.25rem;
    overflow: auto;
    font-family: var(--font-mono);
    font-size: 0.85rem;
    color: var(--dim);
    line-height: 1.7;
    margin-top: 1.5rem;
  }
</style>
<script>
const MAP = ` + string(jsonData) + `;
function search(q){
  const out = document.getElementById('out');
  if(!q){
    out.innerHTML = '<p class="summary">' + MAP.files.length + ' ranked files &middot; ' + MAP.total_tokens + ' tokens</p>' +
      '<ul class="results">' + MAP.files.slice(0,50).map(x=>'<li><span class="path"><code>'+x.path+'</code></span> <span class="meta"><span class="badge badge-refs">'+x.refs+' refs</span><span class="badge badge-syms">'+(x.symbols?x.symbols.length:0)+' syms</span></span></li>').join('') + '</ul>';
    return;
  }
  const terms = q.toLowerCase().split(/[^a-z0-9_./]+/).filter(Boolean);
  const hits = MAP.files
    .map(f=>({f, s:(f.path+' '+(f.symbols||[]).join(' ')).toLowerCase()}))
    .filter(x=>terms.every(t=>x.s.includes(t)))
    .slice(0,50);
  out.innerHTML = hits.length
    ? '<p class="summary">' + hits.length + ' matching files</p><ul class="results">' + hits.map(x=>'<li><span class="path"><code>'+x.f.path+'</code></span> <span class="meta"><span class="badge badge-refs">'+x.f.refs+' refs</span><span class="badge badge-syms">'+(x.f.symbols?x.f.symbols.length:0)+' syms</span></span></li>').join('') + '</ul>'
    : '<p class="no-match">no matching files or symbols found</p>';
}
document.addEventListener('DOMContentLoaded', () => search(''));
</script>
</head>
<body>
<header>
  <h1>` + esc(title) + ` <span class="brand-badge">AST Tree</span></h1>
  <p class="subtitle">Ranked files, symbols, and import references for AI coding agents &middot; Token-budgeted</p>
</header>
<div class="search-wrap">
  <input type="text" id="q" placeholder="search files and symbols… (client-side, instant)" oninput="search(this.value)">
</div>
<div id="out"></div>
<noscript><pre class="tree-dump">` + esc(html.EscapeString(r.RenderTree())) + `</pre></noscript>
</body>
</html>`)
	return sb.String()
}
