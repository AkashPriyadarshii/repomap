package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkashPriyadarshii/repomap/internal/mapx"
)

var version = "v0.1.0" // overridable via -ldflags "-X main.version=..."

func main() {
	var (
		budget      = flag.Int("budget", 4000, "max output tokens (0 = unlimited)")
		full        = flag.Bool("full", false, "show all files, no budget truncation (includes test/vendor/generated)")
		jsonOut     = flag.Bool("json", false, "output JSON")
		noGitignore = flag.Bool("no-gitignore", false, "ignore .gitignore rules")
		index       = flag.Bool("index", false, "write self-contained searchable HTML index")
		indexTitle  = flag.String("index-title", "repomap", "title for --index page")
		out         = flag.String("o", "", "write output to file instead of stdout")
		showVersion = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("repomap " + version)
		return
	}

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	gi := &mapx.Gitignore{}
	if !*noGitignore {
		g, err := mapx.LoadGitignore(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repomap: gitignore: %v\n", err)
			os.Exit(1)
		}
		gi = g
	}

	res, err := mapx.Build(context.Background(), root, gi, *full)
	if err != nil {
		fmt.Fprintf(os.Stderr, "repomap: %v\n", err)
		os.Exit(1)
	}

	if *budget > 0 && !*full {
		res.Budget(*budget)
	}

	writeOut := func(s string) {
		if *out != "" {
			if err := os.WriteFile(*out, []byte(s), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "repomap: write %s: %v\n", *out, err)
				os.Exit(1)
			}
			abs, _ := filepath.Abs(*out)
			fmt.Printf("wrote %s\n", abs)
			return
		}
		fmt.Print(s)
		if len(s) == 0 || s[len(s)-1] != '\n' {
			fmt.Println()
		}
	}

	switch {
	case *index:
		writeOut(res.RenderIndex(*indexTitle))
	case *jsonOut:
		j, err := res.RenderJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "repomap: json: %v\n", err)
			os.Exit(1)
		}
		writeOut(j)
	default:
		writeOut(res.RenderTree())
	}
}
