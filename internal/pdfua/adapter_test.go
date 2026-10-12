package pdfua

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		{"wrong profile", 0, strings.Replace(report("true"), "PDF/UA-2 validation profile", "PDF/UA-1 validation profile", 1), "", true},
		{"malformed report", 0, `<report><jobs><job><validationReport profileName="PDF/UA-2 validation profile" isCompliant="true"></job></jobs></report>`, "", true},
		{"truncated report", 0, `<report><jobs><job><validationReport profileName="PDF/UA-2 validation profile" isCompliant="true"`, "", true},
		{"multiple jobs", 0, strings.Replace(report("true"), "</jobs>", `<job><validationReport profileName="PDF/UA-2 validation profile" isCompliant="true"></validationReport></job></jobs>`, 1), "", true},
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
	if _, err := Validate(ctx, options, pdf); err == nil || !strings.Contains(err.Error(), "cancellation") {
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
			data, err := os.ReadFile(filepath.Join("testdata", tc.name))
			if err != nil {
				t.Fatal(err)
			}
			profile, compliant, err := parseReport(data)
			if err != nil {
				t.Fatal(err)
			}
			if profile != "PDF/UA-2 validation profile" || compliant != tc.compliant {
				t.Fatalf("parser returned profile %q compliant=%v", profile, compliant)
			}
		})
	}
}

func init() {
	if os.Getenv("PDFUA_FAKE_CHILD") != "1" {
		return
	}
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--flavour" && (i+1 >= len(args) || args[i+1] != "ua2") {
			os.Exit(31)
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

func report(compliant string) string {
	countCompliant, countNonCompliant := "0", "1"
	if compliant == "true" {
		countCompliant, countNonCompliant = "1", "0"
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><jobs><job><validationReport profileName="PDF/UA-2 validation profile" isCompliant="` + compliant + `"></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + countCompliant + `" nonCompliant="` + countNonCompliant + `" failedJobs="0">1</validationReports></batchSummary></report>`
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
