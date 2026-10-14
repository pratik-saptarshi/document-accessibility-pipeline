package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	v2ObservationInputLimit  = 64 << 10
	v2ObservationTextLimit   = 256 << 10
	v2ObservationPNGLimit    = 4 << 20
	v2ObservationDiagLimit   = 16 << 10
	v2ObservationVersionCap  = 4 << 10
	v2ObservationWaitDelay   = 250 * time.Millisecond
	v2ObservationDeadline    = 30 * time.Second
	v2ObservationManifestLoc = "docs/fixtures/remediation/financial-table-experiment.v2.json"
)

type v2ObservationFixture struct {
	name       string
	pdfPath    string
	pdfSHA     string
	textPath   string
	textSHA    string
	textBytes  []byte
	renderPath string
	renderSHA  string
	renderData []byte
	inputBytes []byte
}

// ObserveFinancialTableV2 reruns only the four declared synthetic text/render
// controls. It verifies every input and trusted baseline before invoking a
// parser, stages outputs privately, and succeeds only when they exactly match
// the retained evidence. It never publishes or overwrites observations.
func ObserveFinancialTableV2(ctx context.Context, repoRoot string) error {
	if ctx == nil {
		return errors.New("v2 observation requires a cancellable context")
	}
	if repoRoot == "" {
		return errors.New("repository root is required")
	}
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	fixtures, observations, err := preflightV2Observations(root)
	if err != nil {
		return err
	}

	stage, err := os.MkdirTemp("", "financial-table-v2-observations-")
	if err != nil {
		return fmt.Errorf("create private v2 observation staging: %w", err)
	}
	defer os.RemoveAll(stage)
	for _, fixture := range fixtures {
		if err := os.WriteFile(filepath.Join(stage, fixture.name+".pdf"), fixture.inputBytes, 0o600); err != nil {
			return fmt.Errorf("stage verified input %s privately: %w", fixture.name, err)
		}
	}
	textVersion, err := v2ObservationToolVersion(ctx, "pdftotext", []string{"-v"})
	if err != nil {
		return err
	}
	ghostscriptVersion, err := v2ObservationToolVersion(ctx, "gs", []string{"--version"})
	if err != nil {
		return err
	}
	if textVersion != observations.Observations[0].TextTool {
		return fmt.Errorf("pdftotext version %q differs from trusted baseline %q", textVersion, observations.Observations[0].TextTool)
	}
	wantGhostscript := strings.TrimPrefix(observations.Observations[0].RenderTool, "Ghostscript ")
	wantGhostscript = strings.TrimSuffix(wantGhostscript, "; png16m; 72 dpi")
	if ghostscriptVersion != wantGhostscript {
		return fmt.Errorf("Ghostscript version %q differs from trusted baseline %q", ghostscriptVersion, wantGhostscript)
	}

	for _, fixture := range fixtures {
		pdfPath := filepath.Join(stage, fixture.name+".pdf")
		textBytes, err := runV2ObservationTool(ctx, "pdftotext", []string{"-layout", pdfPath, "-"}, v2ObservationTextLimit)
		if err != nil {
			return fmt.Errorf("observe text for %s: %w", fixture.name, err)
		}
		if err := stageAndCompareV2Observation(stage, fixture.name+".txt", textBytes, fixture.textBytes, fixture.textPath, fixture.textSHA, v2ObservationTextLimit); err != nil {
			return err
		}

		gsArgs := []string{"-q", "-dSAFER", "-dBATCH", "-dNOPAUSE", "-sDEVICE=png16m", "-r72", "-sOutputFile=-", pdfPath}
		renderBytes, err := runV2ObservationTool(ctx, "gs", gsArgs, v2ObservationPNGLimit)
		if err != nil {
			return fmt.Errorf("render %s: %w", fixture.name, err)
		}
		if err := stageAndCompareV2Observation(stage, fixture.name+".png", renderBytes, fixture.renderData, fixture.renderPath, fixture.renderSHA, v2ObservationPNGLimit); err != nil {
			return err
		}
	}
	return nil
}

func preflightV2Observations(root string) ([]v2ObservationFixture, generatedV2ObservationManifest, error) {
	fixtureManifestPath := filepath.Join(root, filepath.FromSlash(v2ObservationManifestLoc))
	fixtureManifestBytes, err := readV2FileBounded(fixtureManifestPath, 256<<10)
	if err != nil {
		return nil, generatedV2ObservationManifest{}, fmt.Errorf("read v2 fixture manifest: %w", err)
	}
	var fixtureManifest v2AuthorManifest
	if err := decodeStrictJSON(fixtureManifestBytes, &fixtureManifest); err != nil {
		return nil, generatedV2ObservationManifest{}, fmt.Errorf("decode v2 fixture manifest: %w", err)
	}
	if fixtureManifest.ObservationManifest != "docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json" || len(fixtureManifest.Cases) != 2 {
		return nil, generatedV2ObservationManifest{}, errors.New("v2 manifest does not declare the exact two-family observation scope")
	}
	observationManifestPath := filepath.Join(root, filepath.FromSlash(fixtureManifest.ObservationManifest))
	observationBytes, err := readV2FileBounded(observationManifestPath, 64<<10)
	if err != nil {
		return nil, generatedV2ObservationManifest{}, fmt.Errorf("read trusted v2 observation manifest: %w", err)
	}
	var observations generatedV2ObservationManifest
	if err := decodeStrictJSON(observationBytes, &observations); err != nil {
		return nil, generatedV2ObservationManifest{}, fmt.Errorf("decode trusted v2 observation manifest: %w", err)
	}
	if observations.SchemaVersion != v2ObservationSchema || observations.Scope != "independent text extraction and rendering observations for synthetic metric-complete v2 fixtures" || len(observations.Observations) != 4 {
		return nil, generatedV2ObservationManifest{}, errors.New("trusted v2 observation manifest has unexpected scope or row count")
	}

	wantFamilies := []struct {
		family string
		stem   string
	}{
		{family: "statements", stem: "statement"},
		{family: "invoices", stem: "invoice"},
	}
	fixtures := make([]v2ObservationFixture, 0, 4)
	for familyIndex, family := range wantFamilies {
		item := fixtureManifest.Cases[familyIndex]
		if item.Family != family.family {
			return nil, generatedV2ObservationManifest{}, fmt.Errorf("v2 fixture family %d = %q, want %q", familyIndex, item.Family, family.family)
		}
		for stateIndex, state := range []string{"positive", "damaged"} {
			name := family.stem + "-" + state + "-v2"
			pdfPath := "docs/fixtures/remediation/financial-tables-v2/" + name + ".pdf"
			textPath := "docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".txt"
			renderPath := "docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".png"
			pdfSHA := item.PositiveSHA
			if stateIndex == 0 {
				if item.PositivePath != pdfPath {
					return nil, generatedV2ObservationManifest{}, fmt.Errorf("unexpected positive PDF path for %s", name)
				}
			} else {
				pdfSHA = item.DamagedSHA
				if item.DamagedPath != pdfPath {
					return nil, generatedV2ObservationManifest{}, fmt.Errorf("unexpected damaged PDF path for %s", name)
				}
			}
			observation := observations.Observations[len(fixtures)]
			if observation.Fixture != name+".pdf" || observation.TextOutputPath != textPath || observation.RenderOutputPath != renderPath || observation.TextExitStatus != 0 || observation.RenderExitStatus != 0 || observation.PDFSHA256 != pdfSHA || !isSHA256(observation.TextOutputSHA) || !isSHA256(observation.RenderOutputSHA) {
				return nil, generatedV2ObservationManifest{}, fmt.Errorf("trusted observation metadata for %s does not match the fixed manifest", name)
			}
			fixture := v2ObservationFixture{name: name, pdfPath: pdfPath, pdfSHA: pdfSHA, textPath: textPath, textSHA: observation.TextOutputSHA, renderPath: renderPath, renderSHA: observation.RenderOutputSHA}
			inputPath := filepath.Join(root, filepath.FromSlash(fixture.pdfPath))
			input, err := readV2FileBounded(inputPath, v2ObservationInputLimit)
			if err != nil {
				return nil, generatedV2ObservationManifest{}, fmt.Errorf("preflight v2 input %s: %w", fixture.name, err)
			}
			if got := sha256Hex(input); got != fixture.pdfSHA {
				return nil, generatedV2ObservationManifest{}, fmt.Errorf("preflight v2 input SHA-256 for %s = %s, want %s", fixture.name, got, fixture.pdfSHA)
			}
			if !bytes.HasPrefix(input, []byte("%PDF-2.0\n")) {
				return nil, generatedV2ObservationManifest{}, fmt.Errorf("preflight v2 input %s is not PDF 2.0", fixture.name)
			}
			fixture.inputBytes = input
			fixture.textBytes, err = readV2ObservationBaseline(root, fixture.textPath, fixture.textSHA, v2ObservationTextLimit)
			if err != nil {
				return nil, generatedV2ObservationManifest{}, err
			}
			fixture.renderData, err = readV2ObservationBaseline(root, fixture.renderPath, fixture.renderSHA, v2ObservationPNGLimit)
			if err != nil {
				return nil, generatedV2ObservationManifest{}, err
			}
			fixtures = append(fixtures, fixture)
		}
	}
	for _, observation := range observations.Observations {
		if observation.TextTool != observations.Observations[0].TextTool || !strings.HasPrefix(observation.RenderTool, "Ghostscript ") || observation.RenderTool != observations.Observations[0].RenderTool {
			return nil, generatedV2ObservationManifest{}, errors.New("trusted v2 observation manifest lacks fixed tool identity")
		}
	}
	for familyStart := 0; familyStart < len(fixtures); familyStart += 2 {
		if !bytes.Equal(fixtures[familyStart].textBytes, fixtures[familyStart+1].textBytes) || !bytes.Equal(fixtures[familyStart].renderData, fixtures[familyStart+1].renderData) {
			return nil, generatedV2ObservationManifest{}, fmt.Errorf("trusted positive/damaged observations differ for %s", fixtures[familyStart].name)
		}
	}
	return fixtures, observations, nil
}

func readV2ObservationBaseline(root, relativePath, expectedSHA string, limit int64) ([]byte, error) {
	if !isSHA256(expectedSHA) {
		return nil, fmt.Errorf("trusted observation %s has invalid SHA-256", relativePath)
	}
	data, err := readV2FileBounded(filepath.Join(root, filepath.FromSlash(relativePath)), limit)
	if err != nil {
		return nil, fmt.Errorf("preflight trusted observation %s: %w", relativePath, err)
	}
	if got := sha256Hex(data); got != expectedSHA {
		return nil, fmt.Errorf("trusted observation SHA-256 for %s = %s, want %s", relativePath, got, expectedSHA)
	}
	return data, nil
}

func v2ObservationToolVersion(ctx context.Context, tool string, args []string) (string, error) {
	stdout, stderr, err := runV2ObservationCommand(ctx, tool, args, v2ObservationVersionCap)
	if err != nil {
		return "", fmt.Errorf("read %s version: %w", tool, err)
	}
	combined := append(append([]byte(nil), stdout...), stderr...)
	for _, line := range strings.Split(string(combined), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
	}
	return "", fmt.Errorf("%s version command returned no version", tool)
}

func runV2ObservationTool(ctx context.Context, tool string, args []string, outputLimit int) ([]byte, error) {
	stdout, _, err := runV2ObservationCommand(ctx, tool, args, outputLimit)
	return stdout, err
}

func runV2ObservationCommand(parent context.Context, tool string, args []string, outputLimit int) ([]byte, []byte, error) {
	return runV2ObservationCommandWithRunner(parent, tool, args, outputLimit, runV2BoundedProcess)
}

func runV2ObservationCommandWithRunner(parent context.Context, tool string, args []string, outputLimit int, runner v2BoundedProcessRunner) ([]byte, []byte, error) {
	if parent == nil || outputLimit <= 0 {
		return nil, nil, errors.New("v2 observation command requires context and positive output bound")
	}
	if runner == nil {
		return nil, nil, errors.New("v2 observation command requires a process runner")
	}
	ctx, cancel := context.WithTimeout(parent, v2ObservationDeadline)
	defer cancel()
	stdout := &limitedV2Output{limit: outputLimit}
	stderr := &limitedV2Output{limit: v2ObservationDiagLimit}
	runErr := runner(ctx, tool, args, nil, stdout, stderr, v2ObservationWaitDelay)
	if err := incompleteV2ReapError(tool, ctx, runErr); err != nil {
		return nil, nil, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		if errors.Is(parent.Err(), context.DeadlineExceeded) {
			return nil, nil, fmt.Errorf("tool caller deadline elapsed: %w", parent.Err())
		}
		return nil, nil, fmt.Errorf("tool timed out after %s", v2ObservationDeadline)
	}
	if ctx.Err() != nil {
		return nil, nil, fmt.Errorf("tool canceled: %w", ctx.Err())
	}
	if stdout.Overflow() {
		return nil, nil, fmt.Errorf("tool stdout exceeded %d bytes", outputLimit)
	}
	if stderr.Overflow() {
		return nil, nil, fmt.Errorf("tool stderr exceeded %d bytes", v2ObservationDiagLimit)
	}
	if runErr != nil {
		return nil, nil, fmt.Errorf("tool exited unsuccessfully: %w", runErr)
	}
	return []byte(stdout.String()), []byte(stderr.String()), nil
}

func stageAndCompareV2Observation(stage, name string, actual, expected []byte, relativePath, expectedSHA string, limit int64) error {
	if int64(len(actual)) > limit {
		return fmt.Errorf("staged observation %s exceeds %d bytes", name, limit)
	}
	path := filepath.Join(stage, name)
	if err := os.WriteFile(path, actual, 0o600); err != nil {
		return fmt.Errorf("stage observation %s privately: %w", name, err)
	}
	digest := sha256.Sum256(actual)
	if got := hex.EncodeToString(digest[:]); got != expectedSHA {
		return fmt.Errorf("staged observation SHA-256 for %s = %s, want trusted %s", relativePath, got, expectedSHA)
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("staged observation %s differs from trusted baseline %s", name, relativePath)
	}
	return nil
}
