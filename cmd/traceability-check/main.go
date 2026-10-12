package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/traceability"
)

func main() {
	write := flag.Bool("write", false, "regenerate the Markdown views")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	catalog, err := traceability.LoadCatalog(filepath.Join(*root, "docs", "traceability", "catalog.json"))
	if err != nil {
		fail(err)
	}
	issues, err := traceability.LoadIssues(filepath.Join(*root, ".beads", "issues.jsonl"))
	if err != nil {
		fail(err)
	}
	if problems := traceability.Validate(catalog, issues); len(problems) > 0 {
		for _, problem := range problems {
			fmt.Fprintln(os.Stderr, problem)
		}
		os.Exit(1)
	}

	for relativePath, expected := range traceability.GeneratedMarkdown(catalog) {
		path := filepath.Join(*root, relativePath)
		if *write {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				fail(err)
			}
			if err := os.WriteFile(path, expected, 0o644); err != nil {
				fail(err)
			}
			continue
		}
		actual, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read generated view %s: %v\n", relativePath, err)
			os.Exit(1)
		}
		if !bytes.Equal(actual, expected) {
			fmt.Fprintf(os.Stderr, "generated file differs: %s (run with --write)\n", relativePath)
			os.Exit(1)
		}
	}
	fmt.Printf("traceability valid: %d requirements, %d stories, %d tasks, %d specification-only cases\n", len(catalog.Requirements), len(catalog.Stories), len(catalog.Tasks), len(catalog.Tests))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
