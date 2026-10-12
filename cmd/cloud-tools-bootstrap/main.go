package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap/internal/cloudtools"
)

// buildtools.manifest.json is embedded so cloud environments can install the
// pinned tool set from this module without a repository checkout.
//
//go:embed buildtools.manifest.json
var embeddedManifest []byte

func main() {
	os.Exit(run())
}

func run() int {
	flags := flag.NewFlagSet("cloud-tools-bootstrap", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	manifestPath := flags.String("manifest", "", "optional override for the embedded pinned tool manifest")
	root := flags.String("root", "/workspace/.codex-tools/cloud-build-tools", "isolated tool installation root")
	cache := flags.String("cache", "/workspace/.cache/cloud-build-tools", "isolated download and Go build cache")
	goBinary := flags.String("go", "", "verified Go toolchain executable")
	verifyOnly := flags.Bool("verify", false, "read-only verification of installed tool versions and digests")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return 2
	}
	if *goBinary == "" {
		fmt.Fprintln(os.Stderr, "--go must point to the checksum-verified Go toolchain installed by scripts/bootstrap-cloud-build-tools.sh")
		return 2
	}
	data := embeddedManifest
	if *manifestPath != "" {
		var err error
		data, err = os.ReadFile(filepath.Clean(*manifestPath))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	m, err := cloudtools.DecodeManifest(bytes.NewReader(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	absCache, err := filepath.Abs(*cache)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	ctx, stop := cloudtools.ContextWithSignals(context.Background())
	defer stop()
	if *verifyOnly {
		results, err := cloudtools.VerifyInstalled(ctx, m, cloudtools.Options{Root: absRoot, Cache: absCache, GoBinary: *goBinary})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		failed := false
		for _, result := range results {
			if result.Path == "" {
				fmt.Printf("CI-ONLY %s %s: %s\n", result.Name, result.Version, result.Status)
				continue
			}
			if result.OK {
				fmt.Printf("VERIFIED %s %s at %s (%s)\n", result.Name, result.Version, result.Path, result.Status)
				continue
			}
			failed = true
			fmt.Printf("NOT VERIFIED %s %s at %s (%s)\n", result.Name, result.Version, result.Path, result.Status)
		}
		if failed {
			return 1
		}
		return 0
	}
	if err := cloudtools.Bootstrap(ctx, m, cloudtools.Options{Root: absRoot, Cache: absCache, GoBinary: *goBinary, Output: os.Stdout}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("Tool root: %s\nActivate for this shell with: source %s\n", absRoot, filepath.Join(absRoot, "activate.sh"))
	return 0
}
