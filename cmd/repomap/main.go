package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkashPriyadarshii/repomap/internal/mapx"
)

func main() {
	var (
		budget      = flag.Int("budget", 4000, "max output tokens (0 = unlimited)")
		full        = flag.Bool("full", false, "include test/vendor/generated files")
		jsonOut      = flag.Bool("json", false, "output JSON")
		noGitignore = flag.Bool("no-gitignore", false, "ignore .gitignore rules")
		index       = flag.Bool("index", false, "write self-contained searchable HTML index")
		indexTitle  = flag.String("index-title", "repomap", "title for --index page")
		out         = flag.String("o", "", "write output to file instead of stdout")
		version     = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *version {
		fmt.Println("repomap v0.1.0")
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

	var output string
	switch {
	case *index:
		output = res.RenderIndex(*indexTitle)
	case *jsonOut:
		j, err := res.RenderJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "repomap: json: %v\n", err)
			os.Exit(1)
		}
		output = j
	default:
		output = res.RenderTree()
	}

	if *out != "" {
		if err := os.WriteFile(*out, []byte(output), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "repomap: write %s: %v\n", *out, err)
			os.Exit(1)
		}
		abs, _ := filepath.Abs(*out)
		fmt.Printf("wrote %s\n", abs)
		return
	}
	fmt.Println(output)
}
