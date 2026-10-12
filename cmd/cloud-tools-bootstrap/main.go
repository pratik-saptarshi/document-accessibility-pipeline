package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/cloudtools"
)

// buildtools.manifest.json is embedded so cloud environments can install the
// pinned tool set from this module without a repository checkout.
//
//go:embed buildtools.manifest.json
var embeddedManifest []byte

func main() {
	manifestPath := flag.String("manifest", "", "optional override for the embedded pinned tool manifest")
	root := flag.String("root", "/workspace/.codex-tools/cloud-build-tools", "isolated tool installation root")
	cache := flag.String("cache", "/workspace/.cache/cloud-build-tools", "isolated download and Go build cache")
	goBinary := flag.String("go", "", "verified Go toolchain executable")
	flag.Parse()
	if *goBinary == "" {
		fmt.Fprintln(os.Stderr, "--go must point to the checksum-verified Go toolchain installed by scripts/bootstrap-cloud-build-tools.sh")
		os.Exit(2)
	}
	data := embeddedManifest
	if *manifestPath != "" {
		var err error
		data, err = os.ReadFile(filepath.Clean(*manifestPath))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	m, err := cloudtools.DecodeManifest(bytes.NewReader(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	absCache, err := filepath.Abs(*cache)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cloudtools.Bootstrap(context.Background(), m, cloudtools.Options{Root: absRoot, Cache: absCache, GoBinary: *goBinary, Output: os.Stdout}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Tool root: %s\nActivate for this shell with: source %s\n", absRoot, filepath.Join(absRoot, "activate.sh"))
}
