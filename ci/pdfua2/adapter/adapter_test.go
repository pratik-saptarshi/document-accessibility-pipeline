package pdfua

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	contract "github.com/pratik-saptarshi/document-accessibility-pipeline/internal/pdfua"
)

func TestValidateClassifiesUA2Reports(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		report   string
		want     Outcome
		wantErr  bool
	}{
		{"compliant report", 0, report("true"), OutcomePass, false},
		{"noncompliant report", 1, report("false"), OutcomeFinding, false},
		{"exit and report disagree", 0, report("false"), "", true},
		{"wrong profile", 0, strings.Replace(report("true"), contract.UA2ProfileName, "PDF/UA-1 validation profile", 1), "", true},
		{"malformed report", 0, `<report><jobs><job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="true"></job></jobs></report>`, "", true},
		{"truncated report", 0, `<report><jobs><job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="true"`, "", true},
		{"multiple jobs", 0, strings.Replace(report("true"), "</jobs>", `<job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="true"></validationReport></job></jobs>`, 1), "", true},
		{"batch summary failure", 0, strings.Replace(report("true"), `failedJobs="0"`, `failedJobs="1"`, 1), "", true},
		{"missing report", 0, "", "", true},
		{"unsupported exit code", 2, report("true"), "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pdf := fakePDF(t)
			if tc.name == "missing report" {
				t.Setenv("PDFUA_FAKE_NO_REPORT", "1")
			} else {
				t.Setenv("PDFUA_FAKE_NO_REPORT", "")
				t.Setenv("PDFUA_FAKE_REPORT", tc.report)
			}
			t.Setenv("PDFUA_FAKE_EXIT", string(rune('0'+tc.exitCode)))
			options := fakeOptions(t, fakeArtifact(t))
			options.Timeout = 10 * time.Second
			options.TempDir = t.TempDir()
			result, err := Validate(context.Background(), options, pdf)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && result.Outcome != tc.want {
				t.Fatalf("Outcome = %q, want %q", result.Outcome, tc.want)
			}
			inputHash := sha256.Sum256([]byte("%PDF synthetic test input only"))
			reportHash := sha256.Sum256([]byte(tc.report))
			artifactHash, _ := hex.DecodeString(options.ArtifactSHA256)
			if err == nil && (result.ValidatorVersion != "1.30.3" || result.ArtifactSHA256 != hex.EncodeToString(artifactHash) || result.CandidateSHA256 != hex.EncodeToString(inputHash[:]) || result.ReportSHA256 != hex.EncodeToString(reportHash[:])) {
				t.Fatalf("missing validator or candidate provenance: %#v", result)
			}
		})
	}
}

func TestValidateWithAuthenticatedVeraPDFAndRealUA2Controls(t *testing.T) {
	jar := os.Getenv("PDFUA_REAL_CLI_JAR")
	if jar == "" {
		t.Skip("set PDFUA_REAL_CLI_JAR to run the authenticated veraPDF integration")
	}
	jar, err := filepath.Abs(jar)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := exec.LookPath("java")
	if err != nil {
		t.Fatal(err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		t.Fatal(err)
	}
	if digest := os.Getenv("PDFUA_REAL_CLI_SHA256"); digest == "" {
		t.Fatal("PDFUA_REAL_CLI_SHA256 must pin the authenticated CLI artifact")
	}
	t.Run("tampered authenticated CLI artifact is rejected", func(t *testing.T) {
		original, err := os.ReadFile(jar)
		if err != nil {
			t.Fatal(err)
		}
		tampered := filepath.Join(t.TempDir(), "tampered-cli.jar")
		if err := os.WriteFile(tampered, append(original, []byte("tamper")...), 0o600); err != nil {
			t.Fatal(err)
		}
		input, err := filepath.Abs("../../../internal/pdfua/testdata/oracle-ua2-pass.pdf")
		if err != nil {
			t.Fatal(err)
		}
		options := Options{
			Executable: executable, Artifact: tampered,
			ArtifactSHA256:   os.Getenv("PDFUA_REAL_CLI_SHA256"),
			ValidatorVersion: "1.30.3", Timeout: 30 * time.Second,
			MaxInput: 1 << 20, MaxOutput: 1 << 20, TempDir: t.TempDir(),
		}
		if _, err := Validate(context.Background(), options, input); err == nil || !strings.Contains(err.Error(), "artifact digest mismatch") {
			t.Fatalf("tampered authenticated artifact was not rejected: %v", err)
		}
	})
	for _, tc := range []struct {
		name            string
		file            string
		candidateSHA256 string
		want            Outcome
	}{
		{name: "known PDF/UA-2 pass", file: "../../../internal/pdfua/testdata/oracle-ua2-pass.pdf", candidateSHA256: "c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b", want: OutcomePass},
		{name: "known PDF/UA-2 finding", file: "../../../internal/pdfua/testdata/oracle-ua2-fail.pdf", candidateSHA256: "2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0", want: OutcomeFinding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input, err := filepath.Abs(tc.file)
			if err != nil {
				t.Fatal(err)
			}
			options := Options{
				Executable: executable, Artifact: jar,
				ArtifactSHA256:   os.Getenv("PDFUA_REAL_CLI_SHA256"),
				ValidatorVersion: "1.30.3", Timeout: 30 * time.Second,
				MaxInput: 1 << 20, MaxOutput: 1 << 20, TempDir: t.TempDir(),
			}
			result, err := Validate(context.Background(), options, input)
			if err != nil {
				t.Fatalf("authenticated veraPDF validation failed: %v", err)
			}
			if result.Profile != contract.UA2ProfileName {
				t.Fatalf("veraPDF selected unexpected profile %q", result.Profile)
			}
			if result.Outcome != tc.want {
				t.Fatalf("veraPDF outcome = %q, want %q", result.Outcome, tc.want)
			}
			if result.ValidatorVersion != "1.30.3" || result.ArtifactSHA256 != options.ArtifactSHA256 || result.CandidateSHA256 != tc.candidateSHA256 || result.ReportSHA256 == "" {
				t.Fatalf("missing validator provenance: %#v", result)
			}
		})
	}
}

func TestValidateUsesExplicitUA2AndIsolatedInput(t *testing.T) {
	pdf := fakePDF(t)
	t.Setenv("PDFUA_FAKE_REPORT", report("true"))
	t.Setenv("PDFUA_FAKE_EXIT", "0")
	t.Setenv("PDFUA_FAKE_CAPTURE", filepath.Join(t.TempDir(), "args"))
	artifactPath := fakeArtifact(t)
	artifactData, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	artifactHash := sha256.Sum256(artifactData)
	tempDir := t.TempDir()
	_, err = Validate(context.Background(), Options{Executable: fakeExecutable(t), Artifact: artifactPath, ArtifactSHA256: hex.EncodeToString(artifactHash[:]), ValidatorVersion: "1.30.3", Timeout: 10 * time.Second, MaxInput: 1024, MaxOutput: 4096, TempDir: tempDir}, pdf)
	if err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(os.Getenv("PDFUA_FAKE_CAPTURE"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(args)
	for _, want := range []string{"-Dfile.encoding=UTF8", "--add-exports=java.base/sun.security.pkcs=ALL-UNNAMED", "-jar", "--flavour\tua2", "--format\txml"} {
		if !strings.Contains(got, want) {
			t.Fatalf("arguments %q do not contain %q", got, want)
		}
	}
	if strings.Contains(got, pdf) {
		t.Fatalf("validator received original input path instead of isolated copy: %q", got)
	}
	argv := strings.Split(got, "\t")
	jarIndex := -1
	for i, arg := range argv {
		if arg == "-jar" {
			jarIndex = i
			break
		}
	}
	if jarIndex < 0 || jarIndex+1 >= len(argv) || argv[jarIndex+1] == artifactPath {
		t.Fatalf("expected isolated validator artifact argument, got %q", got)
	}
	if len(argv) == 0 {
		t.Fatalf("expected isolated input path in arguments: %q", got)
	}
	if _, err := os.Stat(argv[len(argv)-1]); !os.IsNotExist(err) {
		t.Fatalf("temporary validation input should be removed after completion, stat err=%v", err)
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary directory should be cleaned; entries=%v err=%v", entries, err)
	}
}

func TestValidateRejectsArtifactDigestMismatch(t *testing.T) {
	pdf := fakePDF(t)
	artifact := fakeArtifact(t)
	options := fakeOptions(t, artifact)
	if err := os.WriteFile(artifact, []byte("different synthetic jar bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(context.Background(), options, pdf); err == nil || !strings.Contains(err.Error(), "artifact digest mismatch") {
		t.Fatalf("modified validator artifact should fail closed, got %v", err)
	}
}

func TestValidateDoesNotExposeInputOrArtifactPaths(t *testing.T) {
	t.Run("input path", func(t *testing.T) {
		marker := "sensitive-input-path-4b2e"
		options := fakeOptions(t, fakeArtifact(t))
		inputPath := filepath.Join(t.TempDir(), marker, "missing.pdf")
		_, err := Validate(context.Background(), options, inputPath)
		if err == nil || strings.Contains(err.Error(), marker) || strings.Contains(err.Error(), inputPath) {
			t.Fatalf("input error leaked sensitive path: %v", err)
		}
	})
	t.Run("artifact path", func(t *testing.T) {
		marker := "sensitive-artifact-path-9ca1"
		options := fakeOptions(t, fakeArtifact(t))
		options.Artifact = filepath.Join(t.TempDir(), marker, "missing.jar")
		_, err := Validate(context.Background(), options, fakePDF(t))
		if err == nil || strings.Contains(err.Error(), marker) || strings.Contains(err.Error(), options.Artifact) {
			t.Fatalf("artifact error leaked sensitive path: %v", err)
		}
	})
}

func TestParseReportRejectsMissingSummaryAttributes(t *testing.T) {
	valid := report("false")
	for _, attr := range []string{
		`totalJobs="1"`, `failedToParse="0"`, `encrypted="0"`, `outOfMemory="0"`, `veraExceptions="0"`,
		`compliant="0"`, `nonCompliant="1"`, `failedJobs="0"`,
	} {
		t.Run(strings.Split(attr, "=")[0], func(t *testing.T) {
			missing := strings.Replace(valid, " "+attr, "", 1)
			if missing == valid {
				t.Fatalf("test fixture does not contain %s", attr)
			}
			if _, _, err := contract.ParseReport([]byte(missing)); err == nil {
				t.Fatalf("report missing %s was accepted", attr)
			}
		})
	}
}

func TestParseReportRejectsDuplicateClassificationAttributes(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"compliance safe then unsafe", strings.Replace(report("true"), `isCompliant="true"`, `isCompliant="true" isCompliant="false"`, 1)},
		{"compliance unsafe then safe", strings.Replace(report("true"), `isCompliant="true"`, `isCompliant="false" isCompliant="true"`, 1)},
		{"counter safe then unsafe", strings.Replace(report("true"), `failedToParse="0"`, `failedToParse="0" failedToParse="1"`, 1)},
		{"counter unsafe then safe", strings.Replace(report("true"), `failedToParse="0"`, `failedToParse="1" failedToParse="0"`, 1)},
		{"summary safe then unsafe", strings.Replace(report("true"), `compliant="1"`, `compliant="1" compliant="0"`, 1)},
		{"summary unsafe then safe", strings.Replace(report("true"), `compliant="1"`, `compliant="0" compliant="1"`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := contract.ParseReport([]byte(tc.input))
			if err == nil {
				t.Fatal("accepted duplicate classification attribute")
			}
			if got := err.Error(); got != "malformed validator XML report" {
				t.Fatalf("duplicate attribute error = %q, want safe generic classification", got)
			}
		})
	}
}

func TestParseReportRejectsNamespaceShadowing(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"default namespace on report", strings.Replace(report("true"), "<report>", `<report xmlns="urn:shadow">`, 1)},
		{"default namespace on classification", strings.Replace(report("true"), "<validationReport", `<validationReport xmlns="urn:shadow"`, 1)},
		{"prefixed shadow element", strings.Replace(strings.Replace(report("true"), "<report>", `<x:report xmlns:x="urn:shadow">`, 1), "</report>", "</x:report>", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := contract.ParseReport([]byte(tc.input)); err == nil {
				t.Fatal("accepted namespaced or shadowed report")
			}
		})
	}
}

func TestParseReportRejectsEmptyRequiredNumericCounters(t *testing.T) {
	for _, attr := range []string{
		`failedToParse="0"`, `encrypted="0"`, `outOfMemory="0"`, `veraExceptions="0"`, `compliant="0"`, `nonCompliant="1"`, `failedJobs="0"`,
	} {
		t.Run(strings.Split(attr, "=")[0], func(t *testing.T) {
			input := strings.Replace(report("false"), attr, strings.Split(attr, "=")[0]+`=""`, 1)
			if _, _, err := contract.ParseReport([]byte(input)); err == nil {
				t.Fatalf("accepted empty required counter %s", attr)
			}
		})
	}
}

func TestParseReportErrorsRedactMalformedTokens(t *testing.T) {
	for _, tc := range []struct{ name, input, marker string }{
		{"malformed numeric value", strings.Replace(report("true"), `failedToParse="0"`, `failedToParse="NUMERIC_SECRET_8472"`, 1), "NUMERIC_SECRET_8472"},
		{"malformed tag", strings.Replace(report("true"), "</jobs>", "</TAG_SECRET_8472>", 1), "TAG_SECRET_8472"},
		{"unknown entity", strings.Replace(report("true"), `isCompliant="true"`, `isCompliant="&ENTITY_SECRET_8472;"`, 1), "ENTITY_SECRET_8472"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := contract.ParseReport([]byte(tc.input))
			if err == nil || strings.Contains(err.Error(), tc.marker) {
				t.Fatalf("parse error = %v, expected safe rejection without %q", err, tc.marker)
			}
		})
	}
}

func TestValidateNormalizesRelativeTempDir(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "relative-temp"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	pdf := fakePDF(t)
	t.Setenv("PDFUA_FAKE_REPORT", report("true"))
	t.Setenv("PDFUA_FAKE_EXIT", "0")
	capture := filepath.Join(root, "args")
	t.Setenv("PDFUA_FAKE_CAPTURE", capture)
	options := fakeOptions(t, fakeArtifact(t))
	options.TempDir = "relative-temp"
	if _, err := Validate(context.Background(), options, pdf); err != nil {
		t.Fatalf("Validate with relative TempDir failed: %v", err)
	}
	args, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	argv := strings.Split(string(args), "\t")
	for i, arg := range argv {
		if arg == "-jar" && i+1 < len(argv) {
			if !filepath.IsAbs(argv[i+1]) {
				t.Fatalf("validator got a relative isolated artifact path: %q", argv[i+1])
			}
		}
	}
	if len(argv) == 0 || !filepath.IsAbs(argv[len(argv)-1]) {
		t.Fatalf("validator got a relative isolated input path: %q", argv)
	}
	entries, err := os.ReadDir(filepath.Join(root, "relative-temp"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary validation directory should be cleaned; entries=%v err=%v", entries, err)
	}
}

func TestValidateChildOpensResolvedArtifactAndInput(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	for _, tc := range []struct {
		name    string
		tempDir string
	}{
		{name: "relative temp directory", tempDir: "relative-temp"},
		{name: "absolute temp directory", tempDir: filepath.Join(root, "absolute-temp")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := filepath.Join(root, filepath.Base(tc.tempDir))
			if err := os.Mkdir(tempDir, 0o700); err != nil {
				t.Fatal(err)
			}
			pdf := fakePDF(t)
			artifact := fakeArtifact(t)
			inputBytes, err := os.ReadFile(pdf)
			if err != nil {
				t.Fatal(err)
			}
			artifactBytes, err := os.ReadFile(artifact)
			if err != nil {
				t.Fatal(err)
			}
			inputHash := sha256.Sum256(inputBytes)
			artifactHash := sha256.Sum256(artifactBytes)
			resolvedCapture := filepath.Join(root, "resolved-"+filepath.Base(tc.tempDir))
			t.Setenv("PDFUA_FAKE_RESOLVE_PATHS", "1")
			t.Setenv("PDFUA_FAKE_EXPECT_INPUT_SHA256", hex.EncodeToString(inputHash[:]))
			t.Setenv("PDFUA_FAKE_EXPECT_ARTIFACT_SHA256", hex.EncodeToString(artifactHash[:]))
			t.Setenv("PDFUA_FAKE_RESOLVED_CAPTURE", resolvedCapture)
			t.Setenv("PDFUA_FAKE_REPORT", report("true"))
			t.Setenv("PDFUA_FAKE_EXIT", "0")

			options := fakeOptions(t, artifact)
			options.Timeout = 3 * time.Second
			options.TempDir = tc.tempDir
			if _, err := Validate(context.Background(), options, pdf); err != nil {
				t.Fatalf("Validate with %s failed: %v", tc.name, err)
			}
			resolved, err := os.ReadFile(resolvedCapture)
			if err != nil {
				t.Fatalf("child did not report opening both paths: %v", err)
			}
			wantResolved := hex.EncodeToString(artifactHash[:]) + "\t" + hex.EncodeToString(inputHash[:])
			if string(resolved) != wantResolved {
				t.Fatalf("child-resolved hashes = %q, want both artifact and input hashes", resolved)
			}
			for path, want := range map[string][]byte{pdf: inputBytes, artifact: artifactBytes} {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatalf("source file %q changed or became unreadable: err=%v", filepath.Base(path), err)
				}
			}
			entries, err := os.ReadDir(tempDir)
			if err != nil || len(entries) != 0 {
				t.Fatalf("temporary validation directory should be cleaned; entries=%v err=%v", entries, err)
			}
		})
	}
}

func TestParseReportRejectsNonWhitespaceOutsideRoot(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"leading", "garbage" + report("true")},
		{"trailing", report("true") + "garbage"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := contract.ParseReport([]byte(tc.input)); err == nil {
				t.Fatalf("accepted non-whitespace %s XML root", tc.name)
			}
		})
	}
}

func TestCopyWithContextStopsPreparationWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := cancelAfterRead{cancel: cancel}
	var destination bytes.Buffer
	if _, err := copyWithContext(ctx, &destination, &source, 1024); !errors.Is(err, context.Canceled) {
		t.Fatalf("copyWithContext error = %v, want context cancellation", err)
	}
	if got := destination.String(); got != "partial" {
		t.Fatalf("copy continued after cancellation: %q", got)
	}
}

type cancelAfterRead struct {
	cancel context.CancelFunc
	done   bool
}

func (r *cancelAfterRead) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	r.done = true
	r.cancel()
	return copy(p, "partial"), nil
}

func TestValidateBoundsTimeoutAndOutput(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		pdf := fakePDF(t)
		t.Setenv("PDFUA_FAKE_SLEEP", "1")
		options := fakeOptions(t, fakeArtifact(t))
		options.Timeout = 20 * time.Millisecond
		options.TempDir = t.TempDir()
		_, err := Validate(context.Background(), options, pdf)
		if err == nil || !strings.Contains(err.Error(), "timeout") {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})
	t.Run("output limit", func(t *testing.T) {
		pdf := fakePDF(t)
		t.Setenv("PDFUA_FAKE_REPORT", strings.Repeat("x", 24))
		t.Setenv("PDFUA_FAKE_STDERR", strings.Repeat("e", 24))
		options := fakeOptions(t, fakeArtifact(t))
		options.MaxOutput = 32
		options.TempDir = t.TempDir()
		_, err := Validate(context.Background(), options, pdf)
		if err == nil || !strings.Contains(err.Error(), "output limit") {
			t.Fatalf("expected output limit error, got %v", err)
		}
	})
}

func TestValidateUnavailableBinaryAndCanceledContextFailClosed(t *testing.T) {
	pdf := fakePDF(t)
	options := fakeOptions(t, fakeArtifact(t))
	options.Executable = filepath.Join(t.TempDir(), "missing-verapdf")
	options.Timeout = time.Second
	if _, err := Validate(context.Background(), options, pdf); err == nil || !strings.Contains(err.Error(), "execution failed") {
		t.Fatalf("unavailable binary should fail closed, got %v", err)
	}
	options.Executable = fakeExecutable(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Validate(ctx, options, pdf); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("canceled context should fail closed, got %v", err)
	}
}

func TestValidateRejectsOversizedInput(t *testing.T) {
	pdf := fakePDF(t)
	if err := os.WriteFile(pdf, []byte(strings.Repeat("x", 64)), 0o600); err != nil {
		t.Fatal(err)
	}
	options := fakeOptions(t, fakeArtifact(t))
	options.Timeout = time.Second
	options.MaxInput = 16
	options.TempDir = t.TempDir()
	_, err := Validate(context.Background(), options, pdf)
	if err == nil || !strings.Contains(err.Error(), "input limit") {
		t.Fatalf("expected input limit error, got %v", err)
	}
}

func TestSyntheticReportSamplesMatchOnlyTheParserContract(t *testing.T) {
	for _, tc := range []struct {
		name      string
		compliant bool
	}{
		{"contract-compliant.xml", true},
		{"contract-noncompliant.xml", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../../internal/pdfua/testdata", tc.name))
			if err != nil {
				t.Fatal(err)
			}
			profile, compliant, err := contract.ParseReport(data)
			if err != nil {
				t.Fatal(err)
			}
			if profile != contract.UA2ProfileName || compliant != tc.compliant {
				t.Fatalf("parser returned profile %q compliant=%v", profile, compliant)
			}
		})
	}
}

func init() {
	if os.Getenv("PDFUA_FAKE_PIPE_HOLDER") == "1" || os.Getenv("PDFUA_FAKE_SPAWN_PIPE_HOLDER") == "1" || os.Getenv("PDFUA_FAKE_SPAWN_CLOSED_PIPE_HOLDER") == "1" {
		return
	}
	if os.Getenv("PDFUA_FAKE_CHILD") != "1" {
		return
	}
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--flavour" && (i+1 >= len(args) || args[i+1] != "ua2") {
			os.Exit(31)
		}
	}
	if os.Getenv("PDFUA_FAKE_RESOLVE_PATHS") == "1" {
		jarPath := ""
		for i, arg := range args {
			if arg == "-jar" && i+1 < len(args) {
				jarPath = args[i+1]
				break
			}
		}
		artifactHash, artifactErr := fakeResolvedFileSHA256(jarPath)
		inputHash := ""
		var inputErr error
		if len(args) > 0 {
			inputHash, inputErr = fakeResolvedFileSHA256(args[len(args)-1])
		} else {
			inputErr = errors.New("missing input argument")
		}
		if artifactErr != nil || inputErr != nil ||
			artifactHash != os.Getenv("PDFUA_FAKE_EXPECT_ARTIFACT_SHA256") ||
			inputHash != os.Getenv("PDFUA_FAKE_EXPECT_INPUT_SHA256") {
			os.Exit(37)
		}
		resolved := os.Getenv("PDFUA_FAKE_RESOLVED_CAPTURE")
		if resolved == "" || os.WriteFile(resolved, []byte(artifactHash+"\t"+inputHash), 0o600) != nil {
			os.Exit(37)
		}
	}
	if capture := os.Getenv("PDFUA_FAKE_CAPTURE"); capture != "" {
		_ = os.WriteFile(capture, []byte(strings.Join(args, "\t")), 0o600)
	}
	if os.Getenv("PDFUA_FAKE_NO_REPORT") != "1" {
		_, _ = os.Stdout.WriteString(os.Getenv("PDFUA_FAKE_REPORT"))
	}
	_, _ = os.Stderr.WriteString(os.Getenv("PDFUA_FAKE_STDERR"))
	if os.Getenv("PDFUA_FAKE_SLEEP") != "" {
		time.Sleep(time.Second)
	}
	switch os.Getenv("PDFUA_FAKE_EXIT") {
	case "1":
		os.Exit(1)
	case "2":
		os.Exit(2)
	default:
		os.Exit(0)
	}
}

func fakeResolvedFileSHA256(path string) (string, error) {
	if path == "" {
		return "", errors.New("missing file argument")
	}
	file, err := openRegularNoFollow(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return "", errors.New("resolved file is not a bounded regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return "", errors.New("resolved file read failed or exceeded limit")
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func report(compliant string) string {
	countCompliant, countNonCompliant := "0", "1"
	if compliant == "true" {
		countCompliant, countNonCompliant = "1", "0"
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><jobs><job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="` + compliant + `"></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + countCompliant + `" nonCompliant="` + countNonCompliant + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

func fakePDF(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synthetic-not-a-conformance-oracle.pdf")
	if err := os.WriteFile(path, []byte("%PDF synthetic test input only"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func fakeExecutable(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("validator process-tree cancellation is supported only on Linux")
	}
	t.Setenv("PDFUA_FAKE_CHILD", "1")
	if os.Getenv("PDFUA_FAKE_REPORT") == "" && os.Getenv("PDFUA_FAKE_NO_REPORT") != "1" {
		t.Setenv("PDFUA_FAKE_REPORT", report("true"))
	}
	if os.Getenv("PDFUA_FAKE_EXIT") == "" {
		t.Setenv("PDFUA_FAKE_EXIT", "0")
	}
	return os.Args[0]
}

func fakeArtifact(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synthetic-verapdf-cli.jar")
	if err := os.WriteFile(path, []byte("synthetic non-executable artifact fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func fakeOptions(t *testing.T, artifact string) Options {
	t.Helper()
	data, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	return Options{Executable: fakeExecutable(t), Artifact: artifact, ArtifactSHA256: hex.EncodeToString(digest[:]), ValidatorVersion: "1.30.3", Timeout: 10 * time.Second, MaxInput: 1024, MaxOutput: 4096, TempDir: t.TempDir()}
}
