package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		return errors.New("usage: financial-fixtures-v2 <author|observations> [flags]")
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
	case "observations":
		flags := flag.NewFlagSet("observations", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		recordsPath := flags.String("records", "", "bounded TSV of external observation hashes")
		outputPath := flags.String("output", "", "observation JSON manifest path")
		textTool := flags.String("pdftotext-version", "", "observed Poppler version")
		ghostscript := flags.String("ghostscript-version", "", "observed Ghostscript version")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *recordsPath == "" || *outputPath == "" {
			return errors.New("observations requires --records and --output, with no positional arguments")
		}
		records, err := readObservationRecords(*recordsPath)
		if err != nil {
			return err
		}
		var manifest bytes.Buffer
		if err := financialfixtures.WriteFinancialTableV2ObservationManifest(records, *textTool, *ghostscript, &manifest); err != nil {
			return err
		}
		return writeAtomic(*outputPath, manifest.Bytes())
	default:
		return fmt.Errorf("unknown financial-fixtures-v2 action %q", args[0])
	}
}

func readObservationRecords(path string) ([]financialfixtures.FinancialTableV2ObservationInput, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open observation records: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(io.LimitReader(file, 16<<10))
	scanner.Buffer(make([]byte, 1024), 4096)
	var records []financialfixtures.FinancialTableV2ObservationInput
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("observation row has %d fields, want 4", len(fields))
		}
		records = append(records, financialfixtures.FinancialTableV2ObservationInput{
			Name: fields[0], PDFSHA256: fields[1], TextSHA256: fields[2], RenderSHA256: fields[3],
		})
		if len(records) > 4 {
			return nil, errors.New("observation record count exceeds four")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read observation records: %w", err)
	}
	return records, nil
}

func writeAtomic(path string, content []byte) error {
	directory := filepath.Dir(path)
	file, err := os.CreateTemp(directory, ".financial-table-observations-*")
	if err != nil {
		return fmt.Errorf("stage observation manifest: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write staged observation manifest: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync staged observation manifest: %w", err)
	}
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("set observation manifest permissions: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged observation manifest: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish observation manifest: %w", err)
	}
	return nil
}
