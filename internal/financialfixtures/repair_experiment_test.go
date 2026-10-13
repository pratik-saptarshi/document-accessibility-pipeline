package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// TestFinancialTableV2DamagedAssociationRepair is an opt-in, test-only bounded
// experiment. It reads only the reviewed immutable damaged inputs and writes
// candidates to private temporary staging; it never publishes candidate bytes.
func TestFinancialTableV2DamagedAssociationRepair(t *testing.T) {
	if os.Getenv("DAP_RUN_V2_REPAIR") != "1" {
		t.Skip("set DAP_RUN_V2_REPAIR=1 only after the exact v2 no-op and clearance gates pass")
	}
	manifest, err := readV2FileBounded(evidenceV2Path, v2PDFInputLimit)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateV2NoopManifestIdentity(manifest); err != nil {
		t.Fatal(err)
	}
	var evidence v2FixtureEvidence
	decoder := json.NewDecoder(bytes.NewReader(manifest))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatalf("decode reviewed v2 manifest: %v", err)
	}
	if err := validateV2NoopCases(evidence.Cases); err != nil {
		t.Fatal(err)
	}
	observationBytes, err := readV2FileBounded(filepath.Join("..", "..", evidence.ObservationManifest), v2PDFInputLimit)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateV2NoopObservationIdentity(observationBytes); err != nil {
		t.Fatal(err)
	}
	var observationManifest v2ObservationManifest
	if err := json.Unmarshal(observationBytes, &observationManifest); err != nil {
		t.Fatal(err)
	}
	observations := make(map[string]externalObservation, len(observationManifest.Observations))
	for _, observation := range observationManifest.Observations {
		observations[observation.Fixture] = observation
	}

	for _, fixture := range evidence.Cases {
		fixture := fixture
		t.Run(fixture.Family, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			observation, ok := observations[filepath.Base(fixture.DamagedPath)]
			if !ok || observation.TextExitStatus != 0 || observation.RenderExit != 0 {
				t.Fatalf("missing successful independent damaged-fixture baselines for %s", fixture.DamagedPath)
			}
			if err := verifyV2NoopToolVersions(ctx, observation.TextTool, observation.RenderTool); err != nil {
				t.Fatalf("verify pinned tools before processing %s: %v", fixture.DamagedPath, err)
			}
			inputPath := filepath.Join("..", "..", fixture.DamagedPath)
			input, inputBytes := readV2PDFBoundedWithContext(t, ctx, inputPath, fixture.DamagedSHA)
			assertV2HelveticaMetrics(ctx, t, input, evidence.Metrics)
			if err := checkV2PreservedDocumentState(input); err != nil {
				t.Fatalf("damaged input document state: %v", err)
			}
			assertReadingOrder(t, input, fixture.ReadingOrder)
			if err := assertCellContentBindings(ctx, input); err != nil {
				t.Fatalf("damaged input content/ParentTree bindings: %v", err)
			}
			if err := assertIDTree(input, mustStructTreeRoot(t, input)); err != nil {
				t.Fatalf("damaged input IDTree: %v", err)
			}
			assertTableAssociations(t, input, fixture.TargetHeader, fixture.TargetCells, fixture.TargetCells[1], fixture.WrongHeader, true)
			beforeHeaders, beforeRows := tableObservations(t, input)
			if err := repairV2DeclaredAssociation(t, input, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader+"-unmapped"); err == nil {
				t.Fatal("repair accepted an undeclared wrong-header mapping")
			}
			afterRejectedHeaders, afterRejectedRows := tableObservations(t, input)
			if !reflect.DeepEqual(beforeHeaders, afterRejectedHeaders) || !reflect.DeepEqual(beforeRows, afterRejectedRows) {
				t.Fatal("rejected repair attempt mutated the input context")
			}
			if err := repairV2DeclaredAssociation(t, input, fixture.TargetHeader, fixture.TargetCells[0], fixture.WrongHeader); err == nil {
				t.Fatal("repair accepted the wrong amount cell for the declared row-two target")
			}
			wrongCellHeaders, wrongCellRows := tableObservations(t, input)
			if !reflect.DeepEqual(beforeHeaders, wrongCellHeaders) || !reflect.DeepEqual(beforeRows, wrongCellRows) {
				t.Fatal("rejected wrong-cell mapping mutated the input context")
			}
			ambiguous, err := api.ReadContext(ctx, bytes.NewReader(inputBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read ambiguous-mapping negative control: %v", err)
			}
			ambiguousCell, err := v2TargetTableCell(ambiguous, 2, targetHeaderColumn(t, ambiguous, fixture.TargetHeader))
			if err != nil {
				t.Fatalf("resolve ambiguous-mapping negative control cell: %v", err)
			}
			ambiguousAttributes, ok := ambiguousCell["A"].(types.Dict)
			if !ok {
				t.Fatalf("ambiguous target attributes = %T, want dictionary", ambiguousCell["A"])
			}
			var targetID, wrongID string
			ambiguousHeaderObservations, _ := tableObservations(t, ambiguous)
			for _, header := range ambiguousHeaderObservations {
				if header.title == fixture.TargetHeader {
					targetID = header.id
				}
				if header.title == fixture.WrongHeader {
					wrongID = header.id
				}
			}
			ambiguousAttributes["Headers"] = types.Array{types.StringLiteral(wrongID), types.StringLiteral(targetID)}
			ambiguousHeaders, ambiguousRows := tableObservations(t, ambiguous)
			if err := repairV2DeclaredAssociation(t, ambiguous, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader); err == nil {
				t.Fatal("repair accepted an ambiguous two-header mapping")
			}
			afterAmbiguousHeaders, afterAmbiguousRows := tableObservations(t, ambiguous)
			if !reflect.DeepEqual(ambiguousHeaders, afterAmbiguousHeaders) || !reflect.DeepEqual(ambiguousRows, afterAmbiguousRows) {
				t.Fatal("rejected ambiguous mapping mutated its target")
			}
			tampered, err := api.ReadContext(ctx, bytes.NewReader(inputBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read protected-value negative control: %v", err)
			}
			tamperedCell, err := v2TargetTableCell(tampered, 2, targetHeaderColumn(t, tampered, fixture.TargetHeader))
			if err != nil {
				t.Fatalf("resolve protected-value negative control cell: %v", err)
			}
			tamperedCell["T"] = types.StringLiteral(fixture.TargetCells[1] + " changed")
			tamperedHeaders, tamperedRows := tableObservations(t, tampered)
			if err := repairV2DeclaredAssociation(t, tampered, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader); err == nil {
				t.Fatal("repair accepted a protected amount value that differed from the declared exact string")
			}
			afterValueRejectedHeaders, afterValueRejectedRows := tableObservations(t, tampered)
			if !reflect.DeepEqual(tamperedHeaders, afterValueRejectedHeaders) || !reflect.DeepEqual(tamperedRows, afterValueRejectedRows) {
				t.Fatal("rejected protected-value repair mutated the target relation")
			}
			beforeFonts, err := snapshotPageFonts(ctx, input)
			if err != nil {
				t.Fatalf("snapshot damaged input fonts: %v", err)
			}
			beforeMetadata := pdfMetadataSnapshot(input)

			if err := repairV2DeclaredAssociation(t, input, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader); err != nil {
				t.Fatalf("repair only declared %s /Headers relation: %v", fixture.TargetCells[1], err)
			}
			afterHeaders, afterRows := tableObservations(t, input)
			assertV2OnlyTargetRelationChanged(t, beforeHeaders, beforeRows, afterHeaders, afterRows, fixture.TargetHeader, fixture.TargetCells[1])
			if err := repairV2DeclaredAssociation(t, input, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader); err == nil {
				t.Fatal("repair accepted an already repaired/no-op relation")
			}
			noOpHeaders, noOpRows := tableObservations(t, input)
			if !reflect.DeepEqual(afterHeaders, noOpHeaders) || !reflect.DeepEqual(afterRows, noOpRows) {
				t.Fatal("rejected no-op repair changed candidate relations")
			}
			if err := assertCellContentBindings(ctx, input); err != nil {
				t.Fatalf("repaired in-memory content/ParentTree bindings: %v", err)
			}

			failureDir := t.TempDir()
			if err := os.Chmod(failureDir, 0o700); err != nil {
				t.Fatal(err)
			}
			failureContext, err := api.ReadContext(ctx, bytes.NewReader(inputBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read bounded-write failure control: %v", err)
			}
			if err := repairV2DeclaredAssociation(t, failureContext, fixture.TargetHeader, fixture.TargetCells[1], fixture.WrongHeader); err != nil {
				t.Fatalf("prepare bounded-write failure control: %v", err)
			}
			if _, _, err := writeV2RepairCandidate(ctx, failureContext, failureDir, "injected-write-failure.pdf", 64); err == nil {
				t.Fatal("bounded repair write accepted the injected small output ceiling")
			}
			if entries, err := os.ReadDir(failureDir); err != nil || len(entries) != 0 {
				t.Fatalf("failed bounded repair write left staged bytes: entries=%v err=%v", entries, err)
			}

			stageDir := t.TempDir()
			if err := os.Chmod(stageDir, 0o700); err != nil {
				t.Fatal(err)
			}
			outputPath, candidateBytes, err := writeV2RepairCandidate(ctx, input, stageDir, fixture.Family+"-v2-repaired.pdf", 256<<10)
			if err != nil {
				t.Fatalf("bounded v2 full-write to private staging: %v", err)
			}
			candidateDigest := sha256.Sum256(candidateBytes)
			fresh, err := api.ReadContext(ctx, bytes.NewReader(candidateBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read private candidate: %v", err)
			}
			if err := api.ValidateContext(ctx, fresh); err != nil {
				t.Fatalf("fresh-read candidate validation: %v", err)
			}
			assertV2HelveticaMetrics(ctx, t, fresh, evidence.Metrics)
			if err := checkV2PreservedDocumentState(fresh); err != nil {
				t.Fatalf("candidate document state: %v", err)
			}
			if err := assertIDTree(fresh, mustStructTreeRoot(t, fresh)); err != nil {
				t.Fatalf("candidate IDTree: %v", err)
			}
			if err := assertCellContentBindings(ctx, fresh); err != nil {
				t.Fatalf("candidate content/ParentTree bindings: %v", err)
			}
			assertReadingOrder(t, fresh, fixture.ReadingOrder)
			assertTableAssociations(t, fresh, fixture.TargetHeader, fixture.TargetCells, "", fixture.WrongHeader, false)
			candidateHeaders, candidateRows := tableObservations(t, fresh)
			assertV2OnlyTargetRelationChanged(t, beforeHeaders, beforeRows, candidateHeaders, candidateRows, fixture.TargetHeader, fixture.TargetCells[1])
			afterFonts, err := snapshotPageFonts(ctx, fresh)
			if err != nil {
				t.Fatalf("snapshot candidate fonts: %v", err)
			}
			if !reflect.DeepEqual(beforeFonts, afterFonts) {
				t.Fatalf("candidate changed full font resources/dictionaries: before=%#v after=%#v", beforeFonts, afterFonts)
			}
			if afterMetadata := pdfMetadataSnapshot(fresh); beforeMetadata != afterMetadata {
				t.Fatalf("candidate changed Info metadata: before=%#v after=%#v", beforeMetadata, afterMetadata)
			}
			textBytes, err := runV2NoopCommand(ctx, 256<<10, "pdftotext", "-layout", outputPath, "-")
			if err != nil {
				t.Fatalf("extract candidate text: %v", err)
			}
			baselineText, err := readV2FileBounded(filepath.Join("..", "..", observation.TextOutputPath), 256<<10)
			textDigest := sha256.Sum256(textBytes)
			if err != nil || !bytes.Equal(textBytes, baselineText) || hex.EncodeToString(textDigest[:]) != observation.TextOutputSHA {
				t.Fatalf("candidate extracted text differs from immutable damaged baseline: err=%v got_sha256=%x want=%s", err, sha256.Sum256(textBytes), observation.TextOutputSHA)
			}
			renderBytes, err := runV2NoopGhostscript(ctx, outputPath)
			if err != nil {
				t.Fatalf("render candidate: %v", err)
			}
			renderDigest := sha256.Sum256(renderBytes)
			if got := hex.EncodeToString(renderDigest[:]); got != observation.RenderOutputSHA {
				t.Fatalf("candidate render differs from immutable damaged baseline: got=%s want=%s", got, observation.RenderOutputSHA)
			}
			if afterInput, err := readV2FileBounded(inputPath, v2PDFInputLimit); err != nil || !bytes.Equal(afterInput, inputBytes) {
				t.Fatalf("immutable damaged source changed during experiment: err=%v", err)
			}
			if outputDir := os.Getenv("DAP_V2_CANDIDATE_DIR"); outputDir != "" {
				if err := retainVerifiedV2Candidate(outputDir, fixture.Family, candidateBytes, candidateDigest); err != nil {
					t.Fatalf("retain locally verified candidate in private staging: %v", err)
				}
			}
			t.Logf("PASS: private repaired candidate preserved all checked non-target structures and exact text/render; input_sha256=%s candidate_sha256=%x text_sha256=%s render_sha256=%s input_bytes=%d candidate_bytes=%d", fixture.DamagedSHA, candidateDigest, observation.TextOutputSHA, hex.EncodeToString(renderDigest[:]), len(inputBytes), len(candidateBytes))
		})
	}
}

func TestFinancialTableV2RetainedCandidatesMatchPinnedStructure(t *testing.T) {
	manifest, err := readV2FileBounded(evidenceV2Path, v2PDFInputLimit)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateV2NoopManifestIdentity(manifest); err != nil {
		t.Fatal(err)
	}
	var evidence v2FixtureEvidence
	decoder := json.NewDecoder(bytes.NewReader(manifest))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatal(err)
	}
	if err := validateV2NoopCases(evidence.Cases); err != nil {
		t.Fatal(err)
	}
	retained := map[string]struct {
		sha256 string
		bytes  int
	}{
		"statements": {sha256: "8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52", bytes: 2507},
		"invoices":   {sha256: "31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448", bytes: 2464},
	}
	for _, fixture := range evidence.Cases {
		fixture := fixture
		t.Run(fixture.Family, func(t *testing.T) {
			want, ok := retained[fixture.Family]
			if !ok {
				t.Fatalf("no retained candidate pin for family %q", fixture.Family)
			}
			candidatePath := filepath.Join("..", "..", "docs", "fixtures", "remediation", "financial-tables-v2", "candidates", map[string]string{
				"statements": "statement-repaired-v2.pdf",
				"invoices":   "invoice-repaired-v2.pdf",
			}[fixture.Family])
			candidateBytes, err := readV2FileBounded(candidatePath, 256<<10)
			if err != nil {
				t.Fatal(err)
			}
			candidateDigest := sha256.Sum256(candidateBytes)
			if got := hex.EncodeToString(candidateDigest[:]); got != want.sha256 || len(candidateBytes) != want.bytes {
				t.Fatalf("retained candidate changed: sha256=%s bytes=%d, want %s/%d", got, len(candidateBytes), want.sha256, want.bytes)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			originalPath := filepath.Join("..", "..", fixture.DamagedPath)
			original, _ := readV2PDFBoundedWithContext(t, ctx, originalPath, fixture.DamagedSHA)
			candidate, err := api.ReadContext(ctx, bytes.NewReader(candidateBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read retained candidate: %v", err)
			}
			if err := api.ValidateContext(ctx, candidate); err != nil {
				t.Fatalf("validate retained candidate: %v", err)
			}
			assertV2HelveticaMetrics(ctx, t, candidate, evidence.Metrics)
			if err := checkV2PreservedDocumentState(candidate); err != nil {
				t.Fatalf("retained candidate document state: %v", err)
			}
			if err := assertIDTree(candidate, mustStructTreeRoot(t, candidate)); err != nil {
				t.Fatalf("retained candidate IDTree: %v", err)
			}
			if err := assertCellContentBindings(ctx, candidate); err != nil {
				t.Fatalf("retained candidate content/ParentTree bindings: %v", err)
			}
			assertReadingOrder(t, candidate, fixture.ReadingOrder)
			assertTableAssociations(t, candidate, fixture.TargetHeader, fixture.TargetCells, "", fixture.WrongHeader, false)
			beforeHeaders, beforeRows := tableObservations(t, original)
			afterHeaders, afterRows := tableObservations(t, candidate)
			assertV2OnlyTargetRelationChanged(t, beforeHeaders, beforeRows, afterHeaders, afterRows, fixture.TargetHeader, fixture.TargetCells[1])
			beforeFonts, err := snapshotPageFonts(ctx, original)
			if err != nil {
				t.Fatal(err)
			}
			afterFonts, err := snapshotPageFonts(ctx, candidate)
			if err != nil || !reflect.DeepEqual(beforeFonts, afterFonts) {
				t.Fatalf("retained candidate changed font resources/dictionaries: err=%v before=%#v after=%#v", err, beforeFonts, afterFonts)
			}
			if before, after := pdfMetadataSnapshot(original), pdfMetadataSnapshot(candidate); before != after {
				t.Fatalf("retained candidate changed Info metadata: before=%#v after=%#v", before, after)
			}
			t.Logf("PASS: hash-pinned retained candidate preserves protected structure, exact strings, order, and all non-target associations; sha256=%s bytes=%d", want.sha256, want.bytes)
		})
	}
}

func writeV2RepairCandidate(ctx context.Context, document *model.Context, directory, name string, maxOutput int64) (string, []byte, error) {
	path := filepath.Join(directory, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", nil, fmt.Errorf("create private candidate staging file: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(path)
		}
	}()
	writeErr := api.WriteContext(ctx, document, &maxBytesWriter{writer: file, remaining: maxOutput})
	if writeErr != nil {
		_ = file.Close()
		return "", nil, fmt.Errorf("bounded candidate write: %w", writeErr)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return "", nil, fmt.Errorf("sync private candidate staging file: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", nil, fmt.Errorf("close private candidate staging file: %w", err)
	}
	data, err := readV2FileBounded(path, maxOutput)
	if err != nil {
		return "", nil, fmt.Errorf("read bounded private candidate: %w", err)
	}
	complete = true
	return path, data, nil
}

func retainVerifiedV2Candidate(directory, family string, data []byte, expected [sha256.Size]byte) error {
	if !filepath.IsAbs(directory) {
		return fmt.Errorf("candidate output directory must be absolute")
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return fmt.Errorf("inspect private candidate directory: %w", err)
	}
	if !info.IsDir() || info.Mode().Perm() != 0o700 {
		return fmt.Errorf("candidate output directory must be private mode 0700")
	}
	if got := sha256.Sum256(data); got != expected {
		return fmt.Errorf("staged candidate changed before retention")
	}
	name := map[string]string{"statements": "statement-repaired-v2.pdf", "invoices": "invoice-repaired-v2.pdf"}[family]
	if name == "" {
		return fmt.Errorf("unsupported candidate family %q", family)
	}
	path := filepath.Join(directory, name)
	file, err := os.CreateTemp(directory, "."+name+"-*.tmp")
	if err != nil {
		return fmt.Errorf("create private candidate staging file: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("set private candidate staging mode: %w", err)
	}
	n, err := file.Write(data)
	if err != nil || n != len(data) {
		_ = file.Close()
		if err == nil {
			err = fmt.Errorf("short write: %d of %d bytes", n, len(data))
		}
		return fmt.Errorf("write private candidate staging file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync private candidate staging file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close private candidate staging file: %w", err)
	}
	retained, err := readV2FileBounded(temporaryPath, 256<<10)
	if err != nil || sha256.Sum256(retained) != expected || !bytes.Equal(retained, data) {
		return fmt.Errorf("staged candidate failed bounded read-back verification: %v", err)
	}
	if err := os.Link(temporaryPath, path); err != nil {
		return fmt.Errorf("atomically publish verified candidate without replacement: %w", err)
	}
	if err := os.Remove(temporaryPath); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("remove candidate staging link after publication: %w", err)
	}
	retained, err = readV2FileBounded(path, 256<<10)
	if err != nil || sha256.Sum256(retained) != expected || !bytes.Equal(retained, data) {
		_ = os.Remove(path)
		return fmt.Errorf("published candidate failed bounded read-back verification: %v", err)
	}
	return nil
}

func TestRetainVerifiedV2CandidatePublishesCompleteBytesWithoutClobber(t *testing.T) {
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	data := []byte("verified synthetic candidate bytes")
	digest := sha256.Sum256(data)
	if err := retainVerifiedV2Candidate(directory, "statements", data, digest); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "statement-repaired-v2.pdf")
	retained, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(retained, data) {
		t.Fatalf("published candidate = %q, %v", retained, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("published candidate mode = %v, %v", info, err)
	}
	if err := retainVerifiedV2Candidate(directory, "statements", []byte("replacement"), sha256.Sum256([]byte("replacement"))); err == nil {
		t.Fatal("candidate retention replaced an existing file")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, data) {
		t.Fatalf("existing candidate changed after rejected replacement: %q, %v", after, err)
	}
	badDigest := sha256.Sum256([]byte("different"))
	if err := retainVerifiedV2Candidate(directory, "invoices", data, badDigest); err == nil {
		t.Fatal("candidate retention accepted a mismatched expected digest")
	}
	if _, err := os.Lstat(filepath.Join(directory, "invoice-repaired-v2.pdf")); !os.IsNotExist(err) {
		t.Fatalf("failed candidate validation left a final path: %v", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("candidate failure left staging file %q", entry.Name())
		}
	}
}

func TestRetainVerifiedV2CandidateConcurrentNoClobber(t *testing.T) {
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	payloads := [][]byte{[]byte("complete candidate one"), []byte("complete candidate two")}
	start := make(chan struct{})
	results := make(chan error, len(payloads))
	for _, payload := range payloads {
		payload := payload
		go func() {
			<-start
			results <- retainVerifiedV2Candidate(directory, "statements", payload, sha256.Sum256(payload))
		}()
	}
	close(start)
	first, second := <-results, <-results
	if (first == nil) == (second == nil) {
		t.Fatalf("concurrent retention results = %v, %v; want exactly one publication", first, second)
	}
	published, err := readV2FileBounded(filepath.Join(directory, "statement-repaired-v2.pdf"), 256<<10)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(published, payloads[0]) && !bytes.Equal(published, payloads[1]) {
		t.Fatalf("published candidate is partial or mixed: %q", published)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "statement-repaired-v2.pdf" {
		t.Fatalf("concurrent publication left unexpected staging files: %v", entries)
	}
}

func repairV2DeclaredAssociation(t *testing.T, ctx *model.Context, targetHeader, targetCell, wrongHeader string) error {
	t.Helper()
	if ctx == nil {
		return fmt.Errorf("nil PDF context")
	}
	headers, _ := tableObservations(t, ctx)
	targetColumn, targetID, wrongID := -1, "", ""
	for _, header := range headers {
		if header.title == targetHeader {
			targetColumn, targetID = header.index, header.id
		}
		if header.title == wrongHeader {
			wrongID = header.id
		}
	}
	if targetColumn < 0 || targetID == "" || wrongID == "" {
		return fmt.Errorf("target/wrong header missing: %q/%q", targetHeader, wrongHeader)
	}
	cell, err := v2TargetTableCell(ctx, 2, targetColumn)
	if err != nil {
		return err
	}
	text, ok := cell["T"].(types.StringLiteral)
	if !ok || string(text) != targetCell || cell["S"] != types.Name("TD") {
		return fmt.Errorf("target cell identity/tag = %#v/%#v, want TD %q", cell["S"], cell["T"], targetCell)
	}
	attributes, ok := cell["A"].(types.Dict)
	if !ok || attributes["O"] != types.Name("Table") {
		return fmt.Errorf("target cell /A = %#v, want Table attribute dictionary", cell["A"])
	}
	prior, ok := attributes["Headers"].(types.Array)
	if !ok || len(prior) != 1 || prior[0] != types.StringLiteral(wrongID) {
		return fmt.Errorf("target cell prior /Headers = %#v, want exactly damaged ID %q", attributes["Headers"], wrongID)
	}
	attributes["Headers"] = types.Array{types.StringLiteral(targetID)}
	return nil
}

func targetHeaderColumn(t *testing.T, ctx *model.Context, title string) int {
	t.Helper()
	headers, _ := tableObservations(t, ctx)
	for _, header := range headers {
		if header.title == title {
			return header.index
		}
	}
	t.Fatalf("target header %q is missing", title)
	return -1
}

func v2TargetTableCell(ctx *model.Context, rowIndex, targetColumn int) (types.Dict, error) {
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return nil, fmt.Errorf("resolve StructTreeRoot: %w", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return nil, fmt.Errorf("resolve Document: %w", err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return nil, fmt.Errorf("resolve Table: %w", err)
	}
	rows, ok := table["K"].(types.Array)
	if !ok || len(rows) != 3 || rowIndex < 1 || rowIndex >= len(rows) {
		return nil, fmt.Errorf("table rows = %T/%d, requested data row %d", table["K"], len(rows), rowIndex)
	}
	row, err := ctx.DereferenceDict(rows[rowIndex])
	if err != nil {
		return nil, fmt.Errorf("resolve target data row: %w", err)
	}
	cells, ok := row["K"].(types.Array)
	if !ok || targetColumn >= len(cells) {
		return nil, fmt.Errorf("target row cells = %T/%d, no column %d", row["K"], len(cells), targetColumn)
	}
	cell, err := ctx.DereferenceDict(cells[targetColumn])
	if err != nil {
		return nil, fmt.Errorf("resolve target cell: %w", err)
	}
	return cell, nil
}

func assertV2OnlyTargetRelationChanged(t *testing.T, beforeHeaders []headerObservation, beforeRows [][]cellObservation, afterHeaders []headerObservation, afterRows [][]cellObservation, targetHeader, targetCell string) {
	t.Helper()
	if !reflect.DeepEqual(beforeHeaders, afterHeaders) || len(beforeRows) != len(afterRows) {
		t.Fatalf("repair changed header identity/order or row count: headers %#v -> %#v; rows %d -> %d", beforeHeaders, afterHeaders, len(beforeRows), len(afterRows))
	}
	targetColumn := -1
	for _, header := range beforeHeaders {
		if header.title == targetHeader {
			targetColumn = header.index
		}
	}
	if targetColumn < 0 || len(beforeRows) != 2 {
		t.Fatalf("missing target header %q or two-row topology", targetHeader)
	}
	targetFound := false
	for row := range beforeRows {
		if len(beforeRows[row]) != len(afterRows[row]) {
			t.Fatalf("repair changed row %d width", row)
		}
		for column := range beforeRows[row] {
			before, after := beforeRows[row][column], afterRows[row][column]
			if before.title != after.title || before.index != after.index || before.role != after.role {
				t.Fatalf("repair changed cell identity/order at row %d column %d: %#v -> %#v", row, column, before, after)
			}
			if row == 1 && column == targetColumn {
				if before.title != targetCell || reflect.DeepEqual(before.headerIDs, after.headerIDs) || len(after.headerIDs) != 1 {
					t.Fatalf("target relation not changed exactly once at row %d column %d: %#v -> %#v", row, column, before, after)
				}
				targetFound = true
				continue
			}
			if !reflect.DeepEqual(before.headerIDs, after.headerIDs) {
				t.Fatalf("repair changed non-target relation at row %d column %d: %#v -> %#v", row, column, before.headerIDs, after.headerIDs)
			}
		}
	}
	if !targetFound {
		t.Fatal("declared row-two target relation was not found")
	}
}
