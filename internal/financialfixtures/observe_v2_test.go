package financialfixtures

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestObserveFinancialTableV2ReproducesTrustedEvidenceWithoutChangingIt(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	root := copyV2ObservationPackage(t)
	installV2ObservationTools(t, root, "success", "")
	before := snapshotV2ObservationPackage(t, root)
	if err := ObserveFinancialTableV2(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	after := snapshotV2ObservationPackage(t, root)
	if !equalByteSnapshots(before, after) {
		t.Fatal("successful observation rerun changed a trusted fixture or baseline")
	}
}

func TestObserveFinancialTableV2PreflightsEveryInputBeforeLaunchingTools(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	lastInput := "docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf"
	for _, testCase := range []struct {
		name      string
		mutate    func(string) error
		wantError string
	}{
		{
			name: "hash mismatch on last input",
			mutate: func(path string) error {
				return os.WriteFile(path, []byte("%PDF-2.0\ncorrupt"), 0o600)
			},
			wantError: "input SHA-256",
		},
		{
			name: "size bound on last input",
			mutate: func(path string) error {
				data := bytes.Repeat([]byte("x"), v2ObservationInputLimit+1)
				copy(data, []byte("%PDF-2.0\n"))
				return os.WriteFile(path, data, 0o600)
			},
			wantError: "exceeds 65536-byte limit",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := copyV2ObservationPackage(t)
			marker := filepath.Join(t.TempDir(), "parser-called")
			installV2ObservationTools(t, root, "success", marker)
			input := filepath.Join(root, lastInput)
			if err := testCase.mutate(input); err != nil {
				t.Fatal(err)
			}
			before := snapshotV2ObservationPackage(t, root)
			if err := ObserveFinancialTableV2(context.Background(), root); err == nil || !strings.Contains(err.Error(), testCase.wantError) {
				t.Fatalf("last input preflight result = %v, want %q", err, testCase.wantError)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("parser marker exists after input preflight rejection: stat error=%v", err)
			}
			if !sameExceptV2Input(t, before, snapshotV2ObservationPackage(t, root), lastInput) {
				t.Fatal("input preflight rejection changed a trusted observation baseline")
			}
		})
	}
}

func TestObserveFinancialTableV2MismatchLeavesTrustedBaselinesUntouched(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	root := copyV2ObservationPackage(t)
	installV2ObservationTools(t, root, "changed-text", "")
	before := snapshotV2ObservationPackage(t, root)
	if err := ObserveFinancialTableV2(context.Background(), root); err == nil || !strings.Contains(err.Error(), "staged observation SHA-256") {
		t.Fatalf("changed observation result = %v, want baseline mismatch", err)
	}
	if !equalByteSnapshots(before, snapshotV2ObservationPackage(t, root)) {
		t.Fatal("mismatched observation changed a trusted fixture or baseline")
	}
}

func TestObserveFinancialTableV2ToolFailureLeavesTrustedBaselinesUntouched(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	root := copyV2ObservationPackage(t)
	installV2ObservationTools(t, root, "fail", "")
	before := snapshotV2ObservationPackage(t, root)
	if err := ObserveFinancialTableV2(context.Background(), root); err == nil || !strings.Contains(err.Error(), "unsuccessfully") {
		t.Fatalf("failed parser result = %v, want nonzero-exit failure", err)
	}
	if !equalByteSnapshots(before, snapshotV2ObservationPackage(t, root)) {
		t.Fatal("tool failure changed a trusted fixture or baseline")
	}
}

func TestV2ObservationCommandBoundsTimeoutOutputAndFailure(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	t.Run("timeout", func(t *testing.T) {
		tool := writeV2TestTool(t, "#!/bin/sh\nexec sleep 30\n")
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()
		if _, err := runV2ObservationTool(ctx, tool, nil, 64); err == nil || !strings.Contains(err.Error(), "caller deadline elapsed") {
			t.Fatalf("stalled tool error = %v, want caller-deadline classification", err)
		}
	})
	t.Run("stdout cap plus one", func(t *testing.T) {
		tool := writeV2TestTool(t, "#!/bin/sh\nprintf 123456789\n")
		if _, err := runV2ObservationTool(context.Background(), tool, nil, 8); err == nil || !strings.Contains(err.Error(), "stdout exceeded 8 bytes") {
			t.Fatalf("oversized tool error = %v, want 8-byte stdout cap", err)
		}
	})
	t.Run("stderr cap plus one", func(t *testing.T) {
		script := fmt.Sprintf("#!/bin/sh\ni=0\nwhile [ $i -lt %d ]; do printf x >&2; i=$((i+1)); done\n", v2ObservationDiagLimit+1)
		tool := writeV2TestTool(t, script)
		if _, _, err := runV2ObservationCommand(context.Background(), tool, nil, 8); err == nil || !strings.Contains(err.Error(), "stderr exceeded") {
			t.Fatalf("oversized diagnostics error = %v, want bounded stderr failure", err)
		}
	})
	t.Run("nonzero exit", func(t *testing.T) {
		tool := writeV2TestTool(t, "#!/bin/sh\nexit 7\n")
		if _, err := runV2ObservationTool(context.Background(), tool, nil, 8); err == nil || !strings.Contains(err.Error(), "unsuccessfully") {
			t.Fatalf("failed tool error = %v, want nonzero-exit failure", err)
		}
	})
}

func copyV2ObservationPackage(t *testing.T) string {
	t.Helper()
	sourceRoot := filepath.Join("..", "..")
	destinationRoot := t.TempDir()
	paths := []string{
		"docs/fixtures/remediation/financial-table-experiment.v2.json",
		"docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json",
	}
	for _, name := range []string{"statement-positive-v2", "statement-damaged-v2", "invoice-positive-v2", "invoice-damaged-v2"} {
		paths = append(paths,
			"docs/fixtures/remediation/financial-tables-v2/"+name+".pdf",
			"docs/fixtures/remediation/financial-tables-v2/observations/"+name+".txt",
			"docs/fixtures/remediation/financial-tables-v2/observations/"+name+".png",
		)
	}
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(sourceRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(destinationRoot, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return destinationRoot
}

func installV2ObservationTools(t *testing.T, root, mode, marker string) {
	t.Helper()
	bin := t.TempDir()
	baseline := filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2/observations")
	pdftotext := `#!/bin/sh
if [ -n "$V2_OBSERVATION_MARKER" ]; then printf called >> "$V2_OBSERVATION_MARKER"; fi
if [ "$1" = "-v" ]; then printf 'pdftotext version 25.03.0\n' >&2; exit 0; fi
case "$2" in
  *statement-positive-v2.pdf) name=statement-positive-v2 ;;
  *statement-damaged-v2.pdf) name=statement-damaged-v2 ;;
  *invoice-positive-v2.pdf) name=invoice-positive-v2 ;;
  *invoice-damaged-v2.pdf) name=invoice-damaged-v2 ;;
  *) exit 8 ;;
esac
if [ "$V2_OBSERVATION_MODE" = "fail" ]; then exit 7; fi
if [ "$V2_OBSERVATION_MODE" = "changed-text" ]; then printf 'changed output\n'; exit 0; fi
cat "$V2_OBSERVATION_BASELINE/$name.txt"
`
	ghostscript := `#!/bin/sh
if [ -n "$V2_OBSERVATION_MARKER" ]; then printf called >> "$V2_OBSERVATION_MARKER"; fi
if [ "$1" = "--version" ]; then printf '10.05.1\n'; exit 0; fi
for arg do input=$arg; done
case "$input" in
  *statement-positive-v2.pdf) name=statement-positive-v2 ;;
  *statement-damaged-v2.pdf) name=statement-damaged-v2 ;;
  *invoice-positive-v2.pdf) name=invoice-positive-v2 ;;
  *invoice-damaged-v2.pdf) name=invoice-damaged-v2 ;;
  *) exit 8 ;;
esac
cat "$V2_OBSERVATION_BASELINE/$name.png"
`
	if err := os.WriteFile(filepath.Join(bin, "pdftotext"), []byte(pdftotext), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "gs"), []byte(ghostscript), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("V2_OBSERVATION_BASELINE", baseline)
	t.Setenv("V2_OBSERVATION_MODE", mode)
	t.Setenv("V2_OBSERVATION_MARKER", marker)
}

func writeV2TestTool(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func snapshotV2ObservationPackage(t *testing.T, root string) map[string][]byte {
	t.Helper()
	paths := []string{
		"docs/fixtures/remediation/financial-table-experiment.v2.json",
		"docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json",
	}
	for _, name := range []string{"statement-positive-v2", "statement-damaged-v2", "invoice-positive-v2", "invoice-damaged-v2"} {
		paths = append(paths,
			"docs/fixtures/remediation/financial-tables-v2/"+name+".pdf",
			"docs/fixtures/remediation/financial-tables-v2/observations/"+name+".txt",
			"docs/fixtures/remediation/financial-tables-v2/observations/"+name+".png",
		)
	}
	result := make(map[string][]byte, len(paths))
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		result[name] = data
	}
	return result
}

func sameExceptV2Input(t *testing.T, before, after map[string][]byte, mutatedPath string) bool {
	t.Helper()
	if len(before) != len(after) {
		return false
	}
	for name, data := range before {
		if name == mutatedPath {
			continue
		}
		if !bytes.Equal(data, after[name]) {
			return false
		}
	}
	return true
}

const v2FixtureManifestRelPath = "docs/fixtures/remediation/financial-table-experiment.v2.json"
const v2ObservationsManifestRelPath = "docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json"

func TestPreflightV2ObservationsRejectsAmbiguousManifestKeys(t *testing.T) {
	for _, manifestPath := range []string{v2FixtureManifestRelPath, v2ObservationsManifestRelPath} {
		for _, tc := range []struct {
			name string
			key  string
		}{
			{name: "exact duplicate", key: `"schema_version": "0.0.0",`},
			{name: "case-variant alias", key: `"Schema_Version": "0.0.0",`},
		} {
			t.Run(filepath.Base(filepath.Dir(manifestPath))+"/"+tc.name, func(t *testing.T) {
				root := copyV2ObservationPackage(t)
				path := filepath.Join(root, filepath.FromSlash(manifestPath))
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				mutated := bytes.Replace(data, []byte("{"), []byte("{\n  "+tc.key), 1)
				if bytes.Equal(mutated, data) {
					t.Fatalf("could not inject key into %s", manifestPath)
				}
				if err := os.WriteFile(path, mutated, 0o600); err != nil {
					t.Fatal(err)
				}

				_, _, err = preflightV2Observations(root)
				if err == nil || !strings.Contains(err.Error(), "duplicate or case-variant key") {
					t.Fatalf("preflight error = %v, want duplicate/case-variant rejection", err)
				}
			})
		}
	}
}
