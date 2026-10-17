package financialfixtures

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	v2AFMSHA256         = "db772f2830fb6d000907791d8d26a12524d96943a9a739e520ee855c6b25c96f"
	v2PDFCPUCommit      = "b41a41d5323f6119ff6a183ad3f287eb4a20a617"
	v2ObservationSchema = "2.0.0"
	v2GoEnvOutputLimit  = 4 << 10
	v2GoEnvWaitDelay    = 250 * time.Millisecond
)

const v2GoEnvTimeout = 5 * time.Second

var (
	v2PDFObjectPattern     = regexp.MustCompile(`(?ms)^([0-9]+) 0 obj\n(.*?)\nendobj\n`)
	v2PDFTrailerPattern    = regexp.MustCompile(`trailer\n<< /Size 23 /Root ([0-9]+) 0 R >>\nstartxref\n[0-9]+\n%%EOF\n?$`)
	errV2HelperReapTimeout = errors.New("financial fixture helper reap timed out")
)

type v2BoundedProcessRunner func(context.Context, string, []string, []string, io.Writer, io.Writer, time.Duration) error

func incompleteV2ReapError(name string, ctx context.Context, runErr error) error {
	if !errors.Is(runErr, errV2HelperReapTimeout) {
		return nil
	}
	contextDetail := ""
	if ctx != nil && ctx.Err() != nil {
		contextDetail = " after " + ctx.Err().Error()
	}
	return fmt.Errorf("%s cleanup incomplete%s; helper and reap/I/O goroutines may still be active, and captured output is incomplete: %w", name, contextDetail, runErr)
}

func freezeV2ProcessOutput(stdout, stderr io.Writer) {
	for _, writer := range []io.Writer{stdout, stderr} {
		if frozen, ok := writer.(interface{ Freeze() }); ok {
			frozen.Freeze()
		}
	}
}

func freezeV2ProcessOutputAfterReapTimeout(stdout, stderr io.Writer, runErr error) error {
	if errors.Is(runErr, errV2HelperReapTimeout) {
		freezeV2ProcessOutput(stdout, stderr)
	}
	return runErr
}

type v2BaseManifest struct {
	Cases []struct {
		Family             string   `json:"family"`
		PositivePath       string   `json:"positive_path"`
		PositiveSHA256     string   `json:"positive_sha256"`
		DamagedPath        string   `json:"damaged_path"`
		DamagedSHA256      string   `json:"damaged_sha256"`
		TargetHeader       string   `json:"target_header"`
		TargetCells        []string `json:"target_cells"`
		DamagedWrongHeader string   `json:"damaged_wrong_header"`
		ReadingOrder       []string `json:"reading_order"`
	} `json:"cases"`
}

type v2AuthorManifest struct {
	SchemaVersion       string          `json:"schema_version"`
	Scope               string          `json:"scope"`
	FontMetrics         v2AuthorMetrics `json:"font_metrics"`
	Cases               []v2AuthorCase  `json:"cases"`
	ObservationManifest string          `json:"external_observation_manifest"`
	ExecutionStatus     string          `json:"execution_status"`
}

type v2AuthorMetrics struct {
	FontName      string           `json:"font_name"`
	Encoding      string           `json:"encoding"`
	FirstChar     int              `json:"first_char"`
	LastChar      int              `json:"last_char"`
	Widths        []int            `json:"widths_1000_em"`
	Descriptor    v2FontDescriptor `json:"descriptor"`
	WidthsSHA256  string           `json:"widths_sha256"`
	DescriptorSHA string           `json:"descriptor_sha256"`
	SourceVersion string           `json:"source_version"`
	SourceCommit  string           `json:"source_commit"`
	AFMPath       string           `json:"afm_path"`
	AFMSHA256     string           `json:"afm_sha256"`
	License       string           `json:"license"`
	LicenseNotice string           `json:"license_notice"`
	Mapping       string           `json:"mapping_algorithm"`
}

type v2FontDescriptor struct {
	Type        string `json:"Type"`
	FontName    string `json:"FontName"`
	Flags       int    `json:"Flags"`
	FontBBox    []int  `json:"FontBBox"`
	ItalicAngle int    `json:"ItalicAngle"`
	Ascent      int    `json:"Ascent"`
	Descent     int    `json:"Descent"`
	CapHeight   int    `json:"CapHeight"`
	StemV       int    `json:"StemV"`
}

type v2AuthorCase struct {
	Family       string   `json:"family"`
	PositivePath string   `json:"positive_path"`
	PositiveSHA  string   `json:"positive_sha256"`
	DamagedPath  string   `json:"damaged_path"`
	DamagedSHA   string   `json:"damaged_sha256"`
	TargetHeader string   `json:"target_header"`
	TargetCells  []string `json:"target_cells"`
	WrongHeader  string   `json:"damaged_wrong_header"`
	ReadingOrder []string `json:"reading_order"`
}

type generatedV2ObservationManifest struct {
	SchemaVersion string             `json:"schema_version"`
	Scope         string             `json:"scope"`
	Observations  []v2ObservationRow `json:"observations"`
}

type v2ObservationRow struct {
	Fixture          string `json:"fixture"`
	PDFSHA256        string `json:"pdf_sha256"`
	TextCommand      string `json:"text_command"`
	TextTool         string `json:"text_tool"`
	TextExitStatus   int    `json:"text_exit_status"`
	TextOutputPath   string `json:"text_output_path"`
	TextOutputSHA    string `json:"text_output_sha256"`
	RenderCommand    string `json:"render_command"`
	RenderTool       string `json:"render_tool"`
	RenderExitStatus int    `json:"render_exit_status"`
	RenderOutputPath string `json:"render_output_path"`
	RenderOutputSHA  string `json:"render_output_sha256"`
}

// v2ObservationInput contains hashes collected by the bounded shell observer.
// The Go helper owns only validation and deterministic JSON serialization.
type FinancialTableV2ObservationInput struct {
	Name         string
	PDFSHA256    string
	TextSHA256   string
	RenderSHA256 string
}

// ResolvePinnedHelveticaAFM locates only the AFM already present in the local
// module cache. An unset GOMODCACHE is resolved with `go env`; no download or
// network command is issued, and the returned file must match the pinned hash.
func ResolvePinnedHelveticaAFM(moduleCache string) (string, error) {
	return ResolvePinnedHelveticaAFMWithContext(context.Background(), moduleCache)
}

// ResolvePinnedHelveticaAFMWithContext resolves the local AFM while honoring
// caller cancellation, including during the bounded `go env` subprocess.
func ResolvePinnedHelveticaAFMWithContext(ctx context.Context, moduleCache string) (string, error) {
	if ctx == nil {
		return "", errors.New("resolve pinned Helvetica AFM requires a cancellable context")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if moduleCache == "" {
		moduleCache = os.Getenv("GOMODCACHE")
	}
	if moduleCache == "" {
		boundedCtx, cancel := context.WithTimeout(ctx, v2GoEnvTimeout)
		defer cancel()
		output, err := readV2GoModuleCache(boundedCtx)
		if err != nil {
			return "", fmt.Errorf("resolve local Go module cache: %w", err)
		}
		moduleCache = output
	}
	if moduleCache == "" {
		return "", errors.New("Go module cache path is empty")
	}
	path := filepath.Join(moduleCache, "github.com/pdfcpu/pdfcpu@v0.16.2/internal/corefont/Core14_AFMs/Helvetica.afm")
	widths, got, err := parseHelveticaAFM(path)
	if err != nil {
		return "", err
	}
	if len(widths) != 256 || got != v2AFMSHA256 {
		return "", fmt.Errorf("cached Helvetica AFM SHA-256 = %s, want %s", got, v2AFMSHA256)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return path, nil
}

// readV2GoModuleCache runs only `go env GOMODCACHE`. Callers supply a deadline;
// stdout/stderr are each capped so a broken or unexpected go executable cannot
// stall or grow this fixture-authoring path without bound.
func readV2GoModuleCache(ctx context.Context) (string, error) {
	return readV2GoModuleCacheWithRunner(ctx, runV2BoundedProcess)
}

func readV2GoModuleCacheWithRunner(ctx context.Context, runner v2BoundedProcessRunner) (string, error) {
	if ctx == nil {
		return "", errors.New("go env GOMODCACHE requires a bounded context")
	}
	if runner == nil {
		return "", errors.New("go env GOMODCACHE requires a process runner")
	}
	stdout := &limitedV2Output{limit: v2GoEnvOutputLimit}
	stderr := &limitedV2Output{limit: v2GoEnvOutputLimit}
	runErr := runner(ctx, "go", []string{"env", "GOMODCACHE"}, v2GoEnvEnvironment(), stdout, stderr, v2GoEnvWaitDelay)
	if err := incompleteV2ReapError("go env GOMODCACHE", ctx, runErr); err != nil {
		return "", err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("go env GOMODCACHE timed out: %w", ctx.Err())
	}
	if ctx.Err() != nil {
		return "", fmt.Errorf("go env GOMODCACHE canceled: %w", ctx.Err())
	}
	if stdout.Overflow() {
		return "", fmt.Errorf("go env GOMODCACHE stdout exceeded %d bytes", v2GoEnvOutputLimit)
	}
	if stderr.Overflow() {
		return "", fmt.Errorf("go env GOMODCACHE stderr exceeded %d bytes", v2GoEnvOutputLimit)
	}
	if runErr != nil {
		return "", fmt.Errorf("go env GOMODCACHE failed: %w", runErr)
	}
	moduleCache := strings.TrimSpace(stdout.String())
	if moduleCache == "" {
		return "", errors.New("go env GOMODCACHE returned an empty path")
	}
	return moduleCache, nil
}

func v2GoEnvEnvironment() []string {
	environment := os.Environ()
	filtered := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "GOTOOLCHAIN=") {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered, "GOTOOLCHAIN=local")
}

type limitedV2Output struct {
	mu       sync.RWMutex
	buffer   bytes.Buffer
	limit    int
	overflow bool
	frozen   bool
}

func (output *limitedV2Output) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.frozen {
		return len(data), nil
	}
	remaining := output.limit - output.buffer.Len()
	if remaining < 0 {
		remaining = 0
	}
	if len(data) > remaining {
		if remaining > 0 {
			_, _ = output.buffer.Write(data[:remaining])
		}
		output.overflow = true
		// Report the whole write as consumed so os/exec continues draining the
		// pipe. The caller rejects the bounded prefix after the child exits.
		return len(data), nil
	}
	return output.buffer.Write(data)
}

func (output *limitedV2Output) String() string {
	output.mu.RLock()
	defer output.mu.RUnlock()
	return output.buffer.String()
}

func (output *limitedV2Output) Overflow() bool {
	output.mu.RLock()
	defer output.mu.RUnlock()
	return output.overflow
}

func (output *limitedV2Output) Freeze() {
	output.mu.Lock()
	defer output.mu.Unlock()
	output.frozen = true
}

// AuthorFinancialTableV2 produces only the four fixed synthetic financial
// fixtures and their metric manifest. It accepts the pinned local AFM path;
// it performs no network access and writes no candidate or repaired PDF.
func AuthorFinancialTableV2(repoRoot, afmPath string) error {
	return AuthorFinancialTableV2WithContext(context.Background(), repoRoot, afmPath)
}

// AuthorFinancialTableV2WithContext authors only the fixed synthetic fixtures
// and checks cancellation before resolving tools and publishing any outputs.
func AuthorFinancialTableV2WithContext(ctx context.Context, repoRoot, afmPath string) error {
	if ctx == nil {
		return errors.New("v2 authoring requires a cancellable context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if repoRoot == "" {
		return errors.New("repository root is required")
	}
	if afmPath == "" {
		var err error
		afmPath, err = ResolvePinnedHelveticaAFMWithContext(ctx, "")
		if err != nil {
			return err
		}
	}
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	widths, afmDigest, err := parseHelveticaAFM(afmPath)
	if err != nil {
		return err
	}
	if afmDigest != v2AFMSHA256 {
		return fmt.Errorf("pinned Helvetica AFM SHA-256 = %s, want %s", afmDigest, v2AFMSHA256)
	}
	basePath := filepath.Join(root, "docs/fixtures/remediation/financial-table-experiment.v1.json")
	baseBytes, err := readV2FileBounded(basePath, 256<<10)
	if err != nil {
		return fmt.Errorf("read v1 fixture manifest: %w", err)
	}
	var base v2BaseManifest
	if err := json.Unmarshal(baseBytes, &base); err != nil {
		return fmt.Errorf("decode v1 fixture manifest: %w", err)
	}
	if len(base.Cases) != 2 {
		return fmt.Errorf("v1 fixture family count = %d, want statements and invoices", len(base.Cases))
	}
	expectedBytes, err := readV2FileBounded(filepath.Join(root, "docs/fixtures/remediation/financial-table-experiment.v2.json"), 256<<10)
	if err != nil {
		return fmt.Errorf("read expected v2 manifest: %w", err)
	}
	var expected v2AuthorManifest
	if err := decodeStrictJSON(expectedBytes, &expected); err != nil {
		return fmt.Errorf("decode expected v2 manifest: %w", err)
	}

	manifest := v2AuthorManifest{
		SchemaVersion:       "2.0.0",
		Scope:               "synthetic-fixtures-only; metric-complete PDF 2.0 inputs; no repair or conformance claim",
		ObservationManifest: "docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json",
		ExecutionStatus:     "fixture authoring and independent observations only; no pdfcpu writer, no repair, and no validator run",
	}
	manifest.FontMetrics = makeV2AuthorMetrics(widths, afmDigest)
	outputs := make(map[string][]byte, 4)
	manifest.Cases = make([]v2AuthorCase, 0, len(base.Cases))
	for _, source := range base.Cases {
		stem := ""
		switch source.Family {
		case "statements":
			stem = "statement"
		case "invoices":
			stem = "invoice"
		default:
			return fmt.Errorf("unsupported v1 fixture family %q", source.Family)
		}
		item := v2AuthorCase{
			Family:       source.Family,
			PositivePath: "docs/fixtures/remediation/financial-tables-v2/" + stem + "-positive-v2.pdf",
			DamagedPath:  "docs/fixtures/remediation/financial-tables-v2/" + stem + "-damaged-v2.pdf",
			TargetHeader: source.TargetHeader,
			TargetCells:  source.TargetCells,
			WrongHeader:  source.DamagedWrongHeader,
			ReadingOrder: source.ReadingOrder,
		}
		for _, state := range []string{"positive", "damaged"} {
			basePath, expectedHash := source.PositivePath, source.PositiveSHA256
			if state == "damaged" {
				basePath, expectedHash = source.DamagedPath, source.DamagedSHA256
			}
			inputPath := filepath.Join(root, filepath.FromSlash(basePath))
			input, err := readV2FileBounded(inputPath, 64<<10)
			if err != nil {
				return fmt.Errorf("read immutable v1 input %s: %w", basePath, err)
			}
			if got := sha256Hex(input); got != expectedHash {
				return fmt.Errorf("immutable v1 input hash for %s = %s, want %s", basePath, got, expectedHash)
			}
			pdf, err := buildMetricCompleteV2PDF(input, widths)
			if err != nil {
				return fmt.Errorf("author %s %s fixture: %w", source.Family, state, err)
			}
			path := "docs/fixtures/remediation/financial-tables-v2/" + stem + "-" + state + "-v2.pdf"
			outputs[path] = pdf
			if state == "positive" {
				item.PositiveSHA = sha256Hex(pdf)
			} else {
				item.DamagedSHA = sha256Hex(pdf)
			}
		}
		manifest.Cases = append(manifest.Cases, item)
	}
	if !reflect.DeepEqual(expected, manifest) {
		return errors.New("v2 manifest pins differ from preflighted source inputs and metrics; no output was published")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := publishV2Artifacts(root, outputs); err != nil {
		return fmt.Errorf("publish preflighted v2 fixtures: %w", err)
	}
	return nil
}

// WriteFinancialTableV2ObservationManifest validates the four shell-collected
// observations and writes their deterministic versioned JSON representation.
func WriteFinancialTableV2ObservationManifest(records []FinancialTableV2ObservationInput, pdftotextVersion, ghostscriptVersion string, output io.Writer) error {
	if len(records) != 4 || pdftotextVersion == "" || ghostscriptVersion == "" || output == nil {
		return errors.New("four observations, tool versions, and output are required")
	}
	allowed := []string{"statement-positive-v2", "statement-damaged-v2", "invoice-positive-v2", "invoice-damaged-v2"}
	rows := make([]v2ObservationRow, len(records))
	for i, record := range records {
		if record.Name != allowed[i] || !isSHA256(record.PDFSHA256) || !isSHA256(record.TextSHA256) || !isSHA256(record.RenderSHA256) {
			return fmt.Errorf("invalid or out-of-order observation record %d: %#v", i, record)
		}
		name := record.Name
		rows[i] = v2ObservationRow{
			Fixture:          name + ".pdf",
			PDFSHA256:        record.PDFSHA256,
			TextCommand:      "pdftotext -layout docs/fixtures/remediation/financial-tables-v2/" + name + ".pdf docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".txt",
			TextTool:         pdftotextVersion,
			TextExitStatus:   0,
			TextOutputPath:   "docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".txt",
			TextOutputSHA:    record.TextSHA256,
			RenderCommand:    "gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r72 -sOutputFile=docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".png docs/fixtures/remediation/financial-tables-v2/" + name + ".pdf",
			RenderTool:       "Ghostscript " + ghostscriptVersion + "; png16m; 72 dpi",
			RenderExitStatus: 0,
			RenderOutputPath: "docs/fixtures/remediation/financial-tables-v2/observations/" + name + ".png",
			RenderOutputSHA:  record.RenderSHA256,
		}
	}
	manifest := generatedV2ObservationManifest{
		SchemaVersion: v2ObservationSchema,
		Scope:         "independent text extraction and rendering observations for synthetic metric-complete v2 fixtures",
		Observations:  rows,
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func parseHelveticaAFM(path string) ([]int, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open pinned Helvetica AFM: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 2<<20))
	if err != nil {
		return nil, "", fmt.Errorf("read pinned Helvetica AFM: %w", err)
	}
	if len(data) == 2<<20 {
		return nil, "", errors.New("pinned Helvetica AFM exceeds 2 MiB")
	}
	glyphByCode := map[int]string{}
	widthByGlyph := map[string]int{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 64<<10)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 || fields[0] != "C" || fields[2] != ";" || fields[3] != "WX" || fields[5] != ";" || fields[6] != "N" {
			continue
		}
		code, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, "", fmt.Errorf("parse AFM character code %q: %w", fields[1], err)
		}
		width, err := strconv.Atoi(fields[4])
		if err != nil || width < 0 {
			return nil, "", fmt.Errorf("invalid AFM width %q", fields[4])
		}
		name := fields[7]
		widthByGlyph[name] = width
		if code >= 0 {
			if code > 255 {
				return nil, "", fmt.Errorf("StandardEncoding AFM code %d outside 0..255", code)
			}
			glyphByCode[code] = name
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("scan pinned Helvetica AFM: %w", err)
	}
	widths := make([]int, 256)
	for code := range widths {
		widths[code] = widthByGlyph[glyphByCode[code]]
	}
	return widths, sha256Hex(data), nil
}

func makeV2AuthorMetrics(widths []int, afmDigest string) v2AuthorMetrics {
	descriptor := v2FontDescriptor{
		Type: "FontDescriptor", FontName: "Helvetica", Flags: 32,
		FontBBox: []int{-166, -225, 1000, 931}, ItalicAngle: 0,
		Ascent: 718, Descent: -207, CapHeight: 718, StemV: 88,
	}
	widthText := make([]string, len(widths))
	for i, width := range widths {
		widthText[i] = strconv.Itoa(width)
	}
	widthHash := sha256Hex([]byte(strings.Join(widthText, ",")))
	descriptorBytes, _ := json.Marshal(map[string]any{
		"Type": descriptor.Type, "FontName": descriptor.FontName, "Flags": descriptor.Flags,
		"FontBBox": descriptor.FontBBox, "ItalicAngle": descriptor.ItalicAngle, "Ascent": descriptor.Ascent,
		"Descent": descriptor.Descent, "CapHeight": descriptor.CapHeight, "StemV": descriptor.StemV,
	})
	return v2AuthorMetrics{
		FontName: "Helvetica", Encoding: "StandardEncoding", FirstChar: 0, LastChar: 255,
		Widths: widths, Descriptor: descriptor, WidthsSHA256: widthHash, DescriptorSHA: sha256Hex(descriptorBytes),
		SourceVersion: "github.com/pdfcpu/pdfcpu v0.16.2", SourceCommit: v2PDFCPUCommit,
		AFMPath: "internal/corefont/Core14_AFMs/Helvetica.afm", AFMSHA256: afmDigest,
		License:       "Adobe Core 14 AFM permission notice; see AFM-NOTICE.txt",
		LicenseNotice: "docs/fixtures/remediation/financial-tables-v2/AFM-NOTICE.txt",
		Mapping:       "Pinned pdfcpu v0.16.2 metrics/gen.go: map each Helvetica AFM C code >= 0 to its N glyph; use that glyph WX for the corresponding StandardEncoding code, set unmapped codes to 0; widths cover codes 0..255.",
	}
}

func buildMetricCompleteV2PDF(input []byte, widths []int) ([]byte, error) {
	if len(input) == 0 || len(input) > 64<<10 || !bytes.HasPrefix(input, []byte("%PDF-2.0\n")) {
		return nil, errors.New("input must be a bounded PDF 2.0 fixture")
	}
	if len(widths) != 256 {
		return nil, fmt.Errorf("Helvetica width count = %d, want 256", len(widths))
	}
	objects := make(map[int][]byte, 23)
	for _, match := range v2PDFObjectPattern.FindAllSubmatch(input, -1) {
		number, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, fmt.Errorf("parse fixture object number: %w", err)
		}
		if _, duplicate := objects[number]; duplicate {
			return nil, fmt.Errorf("duplicate fixture object %d", number)
		}
		objects[number] = append([]byte(nil), match[2]...)
	}
	if len(objects) != 22 {
		return nil, fmt.Errorf("fixture object count = %d, want exactly 22", len(objects))
	}
	for number := 1; number <= 22; number++ {
		if _, ok := objects[number]; !ok {
			return nil, fmt.Errorf("fixture object %d is missing", number)
		}
	}
	if !bytes.Equal(objects[5], []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")) {
		return nil, errors.New("fixture object 5 is not the reviewed minimal Helvetica dictionary")
	}
	trailer := v2PDFTrailerPattern.FindSubmatch(input)
	if len(trailer) != 2 {
		return nil, errors.New("fixture does not have the reviewed classic trailer")
	}
	widthText := make([]string, len(widths))
	for i, width := range widths {
		if width < 0 || width > 10000 {
			return nil, fmt.Errorf("Helvetica width[%d] = %d is outside bounds", i, width)
		}
		widthText[i] = strconv.Itoa(width)
	}
	objects[5] = []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /StandardEncoding /FirstChar 0 /LastChar 255 /Widths [" + strings.Join(widthText, " ") + "] /FontDescriptor 23 0 R >>")
	objects[23] = []byte("<< /Type /FontDescriptor /FontName /Helvetica /Flags 32 /FontBBox [-166 -225 1000 931] /ItalicAngle 0 /Ascent 718 /Descent -207 /CapHeight 718 /StemV 88 >>")

	var output bytes.Buffer
	output.WriteString("%PDF-2.0\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, 24)
	for number := 1; number <= 23; number++ {
		offsets[number] = output.Len()
		fmt.Fprintf(&output, "%d 0 obj\n", number)
		output.Write(objects[number])
		output.WriteString("\nendobj\n")
	}
	xrefOffset := output.Len()
	output.WriteString("xref\n0 24\n0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&output, "%010d 00000 n \n", offset)
	}
	rootNumber := string(trailer[1])
	fmt.Fprintf(&output, "trailer\n<< /Size 24 /Root %s 0 R >>\nstartxref\n%d\n%%%%EOF\n", rootNumber, xrefOffset)
	return output.Bytes(), nil
}

func marshalIndented(value any) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing JSON data: %v", err)
	}
	return nil
}

func readV2FileBounded(path string, limit int64) ([]byte, error) {
	return readV2FileBoundedWithContext(context.Background(), path, limit)
}

func readV2FileBoundedWithContext(ctx context.Context, path string, limit int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return nil, errors.New("positive file limit required")
	}
	file, err := openRegularInput(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file %s exceeds %d-byte limit", path, limit)
	}
	return data, nil
}

type stagedV2Artifact struct {
	path      string
	temporary string
	expected  []byte
	fileInfo  os.FileInfo
}

// v2ArtifactOps keeps failure injection local to the existing four-file
// publisher. It is deliberately not a general filesystem abstraction.
type v2ArtifactOps struct {
	createTemp func(dir, pattern string) (*os.File, error)
	link       func(oldPath, newPath string) error
	remove     func(path string) error
	readBack   func(path string) ([]byte, error)
}

func defaultV2ArtifactOps() v2ArtifactOps {
	return v2ArtifactOps{
		createTemp: os.CreateTemp,
		link:       os.Link,
		remove:     os.Remove,
		readBack: func(path string) ([]byte, error) {
			return readV2FileBounded(path, 64<<10)
		},
	}
}

func (ops v2ArtifactOps) withDefaults() v2ArtifactOps {
	defaults := defaultV2ArtifactOps()
	if ops.createTemp == nil {
		ops.createTemp = defaults.createTemp
	}
	if ops.link == nil {
		ops.link = defaults.link
	}
	if ops.remove == nil {
		ops.remove = defaults.remove
	}
	if ops.readBack == nil {
		ops.readBack = defaults.readBack
	}
	return ops
}

func publishV2Artifacts(root string, outputs map[string][]byte) error {
	return publishV2ArtifactsWithOps(root, outputs, defaultV2ArtifactOps())
}

func publishV2ArtifactsWithOps(root string, outputs map[string][]byte, ops v2ArtifactOps) error {
	ops = ops.withDefaults()
	allowed := []string{
		"docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
	}
	if len(outputs) != len(allowed) {
		return fmt.Errorf("output count %d, want only %d declared v2 paths", len(outputs), len(allowed))
	}
	staged := make([]stagedV2Artifact, 0, len(allowed))
	cleanup := func() error {
		var failures []string
		for i := range staged {
			item := &staged[i]
			if item.temporary != "" {
				if err := ops.remove(item.temporary); err != nil && !errors.Is(err, os.ErrNotExist) {
					failures = append(failures, fmt.Sprintf("%s: %v", item.temporary, err))
					continue
				}
				item.temporary = ""
			}
		}
		if len(failures) > 0 {
			return errors.New(strings.Join(failures, "; "))
		}
		return nil
	}
	missing := make([]string, 0, len(allowed))
	for _, path := range allowed {
		data, ok := outputs[path]
		if !ok {
			return fmt.Errorf("missing declared output %s", path)
		}
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		original, err := readV2FileBounded(fullPath, 64<<10)
		if errors.Is(err, os.ErrNotExist) {
			missing = append(missing, path)
			continue
		}
		if err != nil {
			return fmt.Errorf("read existing v2 baseline %s: %w", path, err)
		}
		if !bytes.Equal(original, data) {
			return fmt.Errorf("existing v2 baseline %s differs from preflighted bytes; refusing to overwrite it", path)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	for _, path := range missing {
		data := outputs[path]
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		file, err := ops.createTemp(filepath.Dir(fullPath), ".financial-table-v2-*")
		if err != nil {
			if cleanupErr := cleanup(); cleanupErr != nil {
				return fmt.Errorf("stage %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("stage %s: %w", path, err)
		}
		temporary := file.Name()
		staged = append(staged, stagedV2Artifact{
			path:      fullPath,
			temporary: temporary,
			expected:  append([]byte(nil), data...),
		})
		written, err := file.Write(data)
		if err == nil && written != len(data) {
			err = io.ErrShortWrite
		}
		if err != nil {
			_ = file.Close()
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("stage bytes for %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("stage bytes for %s: %w", path, err)
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("sync staged %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("sync staged %s: %w", path, err)
		}
		if err := file.Chmod(0o644); err != nil {
			_ = file.Close()
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("set staged permissions for %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("set staged permissions for %s: %w", path, err)
		}
		fileInfo, err := file.Stat()
		if err != nil {
			_ = file.Close()
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("inspect staged %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("inspect staged %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("close staged %s: %w; staged-file cleanup failed and temporary files may remain: %v", path, err, cleanupErr)
			}
			return fmt.Errorf("close staged %s: %w", path, err)
		}
		staged[len(staged)-1].fileInfo = fileInfo
	}
	committed := 0
	for i := range staged {
		if err := ops.link(staged[i].temporary, staged[i].path); err != nil {
			rollbackErr := rollbackV2Artifacts(staged[:committed], ops.remove)
			cleanupErr := cleanup()
			if rollbackErr != nil {
				if cleanupErr != nil {
					return fmt.Errorf("replace %s: %w; rollback failed and residual publication state may remain: %v; staged-file cleanup also failed: %v", staged[i].path, err, rollbackErr, cleanupErr)
				}
				return fmt.Errorf("replace %s: %w; rollback failed and residual publication state may remain: %v", staged[i].path, err, rollbackErr)
			}
			if cleanupErr != nil {
				return fmt.Errorf("replace %s: %w; staged-file cleanup failed and temporary files may remain: %v", staged[i].path, err, cleanupErr)
			}
			return fmt.Errorf("replace %s: %w", staged[i].path, err)
		}
		committed++
		readBack, err := ops.readBack(staged[i].path)
		if err == nil && !bytes.Equal(readBack, staged[i].expected) {
			err = fmt.Errorf("read-back bytes differ from expected output")
		}
		if err != nil {
			verifyErr := fmt.Errorf("verify published %s: %w", staged[i].path, err)
			rollbackErr := rollbackV2Artifacts(staged[:committed], ops.remove)
			if rollbackErr != nil {
				cleanupErr := cleanup()
				if cleanupErr != nil {
					return fmt.Errorf("%w; rollback failed and residual publication state may remain: %v; staged-file cleanup also failed: %v", verifyErr, rollbackErr, cleanupErr)
				}
				return fmt.Errorf("%w; rollback failed and residual publication state may remain: %v", verifyErr, rollbackErr)
			}
			cleanupErr := cleanup()
			if cleanupErr != nil {
				return fmt.Errorf("%w; staged-file cleanup failed and temporary files may remain: %v", verifyErr, cleanupErr)
			}
			return verifyErr
		}
	}
	if cleanupErr := cleanup(); cleanupErr != nil {
		return fmt.Errorf("all published v2 outputs passed read-back verification, but staged-file cleanup failed and temporary files may remain: %w", cleanupErr)
	}
	return nil
}

func rollbackV2Artifacts(items []stagedV2Artifact, remove func(string) error) error {
	var failures []string
	for i := len(items) - 1; i >= 0; i-- {
		current, err := os.Lstat(items[i].path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: inspect target: %v", items[i].path, err))
			continue
		}
		if items[i].fileInfo == nil || !os.SameFile(items[i].fileInfo, current) {
			failures = append(failures, fmt.Sprintf("%s: target changed after publication; not removing it", items[i].path))
			continue
		}
		if err := remove(items[i].path); err != nil && !errors.Is(err, os.ErrNotExist) {
			failures = append(failures, fmt.Sprintf("%s: %v", items[i].path, err))
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
