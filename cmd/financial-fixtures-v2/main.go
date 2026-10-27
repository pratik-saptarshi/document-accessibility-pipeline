package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/financialfixtures"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if ctx == nil {
		return errors.New("command requires a cancellable context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("usage: financial-fixtures-v2 <author|observe> [flags]")
	}
	switch args[0] {
	case "author":
		flags := flag.NewFlagSet("author", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		repoRoot := flags.String("repo-root", ".", "repository root")
		moduleCache := flags.String("module-cache", "", "existing Go module cache containing the build's resolved pdfcpu version")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("author accepts no positional arguments")
		}
		afmpath, err := financialfixtures.ResolvePinnedHelveticaAFMWithContext(ctx, *moduleCache)
		if err != nil {
			return err
		}
		return financialfixtures.AuthorFinancialTableV2WithContext(ctx, *repoRoot, afmpath)
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
		return financialfixtures.ObserveFinancialTableV2(ctx, *repoRoot)
	default:
		return fmt.Errorf("unknown financial-fixtures-v2 action %q", args[0])
	}
}
