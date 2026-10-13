package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/financialfixtures"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: financial-fixtures-v2 <author|observe> [flags]")
	}
	switch args[0] {
	case "author":
		flags := flag.NewFlagSet("author", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		repoRoot := flags.String("repo-root", ".", "repository root")
		moduleCache := flags.String("module-cache", "", "existing Go module cache containing pinned pdfcpu v0.16.2")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("author accepts no positional arguments")
		}
		afmpath, err := financialfixtures.ResolvePinnedHelveticaAFM(*moduleCache)
		if err != nil {
			return err
		}
		return financialfixtures.AuthorFinancialTableV2(*repoRoot, afmpath)
	case "observe":
		flags := flag.NewFlagSet("observe", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		repoRoot := flags.String("repo-root", ".", "repository root")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("observe accepts no positional arguments")
		}
		return financialfixtures.ObserveFinancialTableV2(context.Background(), *repoRoot)
	default:
		return fmt.Errorf("unknown financial-fixtures-v2 action %q", args[0])
	}
}
