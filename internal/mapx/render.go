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
<title>` + esc(title) + ` — repo map</title>
<meta name="description" content="Token-budgeted repo map for AI agents. ` + esc(title) + ` files, symbols, refs.">
<meta name="robots" content="index, follow">
<meta property="og:type" content="website">
<meta property="og:title" content="` + esc(title) + ` repo map">
<meta property="og:description" content="Machine-readable repo map: files, symbols, import refs, ranked for LLM agents.">
<meta name="keywords" content="repomap, AI, agent, code map, repo map, symbols, ` + esc(title) + `">
<script>
const MAP = ` + string(jsonData) + `;
function search(q){
  const out = document.getElementById('out');
  if(!q){ out.innerHTML = '<p>' + MAP.files.length + ' files, ' + MAP.total_tokens + ' tokens</p>'; return; }
  const terms = q.toLowerCase().split(/[^a-z0-9_./]+/).filter(Boolean);
  const hits = MAP.files
    .map(f=>({f, s:(f.path+' '+f.symbols.join(' ')).toLowerCase()}))
    .filter(x=>terms.every(t=>x.s.includes(t)))
    .slice(0,50);
  out.innerHTML = hits.length
    ? '<ul>' + hits.map(x=>'<li><code>'+x.f.path+'</code> <small>'+x.f.refs+' refs · '+x.f.symbols.length+' syms</small></li>').join('') + '</ul>'
    : '<p>no matches</p>';
}
</script>
<style>
  body{font:16px/1.5 system-ui,sans-serif;max-width:900px;margin:2rem auto;padding:0 1rem;color:#1a1a1a}
  input{width:100%;padding:.6rem;font-size:1rem;border:1px solid #ccc;border-radius:6px}
  code{background:#f4f4f4;padding:.1rem .3rem;border-radius:4px}
  ul{list-style:none;padding:0} li{padding:.25rem 0;border-bottom:1px solid #eee}
  small{color:#666}
  h1{font-size:1.4rem}
  pre{background:#f8f8f8;padding:1rem;overflow:auto;font-size:.85rem}
</style>
</head>
<body>
<h1>` + esc(title) + ` — repo map</h1>
<input type="text" id="q" placeholder="search files and symbols… (client-side, no server)" oninput="search(this.value)">
<div id="out"></div>
<noscript><pre>` + esc(html.EscapeString(r.RenderTree())) + `</pre></noscript>
</body>
</html>`)
	return sb.String()
}
