// Package pdfua contains a CI/test-only adapter contract for independent
// veraPDF PDF/UA-2 validation. It does not establish semantic accessibility or
// conformance by itself.
package pdfua

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Outcome string

const (
	OutcomePass    Outcome = "pass"
	OutcomeFinding Outcome = "finding"
	ua2ProfileName         = "PDF/UA-2 validation profile"
)

type Options struct {
	Executable       string
	Artifact         string
	ArtifactSHA256   string
	ValidatorVersion string
	Timeout          time.Duration
	MaxInput         int64
	MaxOutput        int64
	TempDir          string
}

type Result struct {
	ValidatorVersion string
	Profile          string
	Outcome          Outcome
	ExitCode         int
	ArtifactSHA256   string
	CandidateSHA256  string
	ReportSHA256     string
}

type boundedBuffer struct {
	buffer bytes.Buffer
	over   bool
	budget *outputBudget
}

type outputBudget struct {
	mu        sync.Mutex
	remaining int64
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.budget.mu.Lock()
	defer b.budget.mu.Unlock()
	if b.budget.remaining <= 0 {
		b.over = true
		return 0, errors.New("output limit exceeded")
	}
	if int64(len(p)) > b.budget.remaining {
		remaining := b.budget.remaining
		_, _ = b.buffer.Write(p[:remaining])
		b.budget.remaining = 0
		b.over = true
		return int(remaining), errors.New("output limit exceeded")
	}
	n, err := b.buffer.Write(p)
	b.budget.remaining -= int64(n)
	return n, err
}

// Validate copies the input into a private temporary directory, then invokes
// the configured executable with fixed PDF/UA-2 XML arguments. Diagnostics
// intentionally omit child output, input paths, and document content.
func Validate(parent context.Context, options Options, inputPath string) (Result, error) {
	if options.Executable == "" || !filepath.IsAbs(options.Executable) {
		return Result{}, fmt.Errorf("validator executable must be an explicit absolute path")
	}
	if strings.TrimSpace(options.ValidatorVersion) == "" {
		return Result{}, fmt.Errorf("validator version must be explicit")
	}
	if options.Artifact == "" || !filepath.IsAbs(options.Artifact) {
		return Result{}, fmt.Errorf("validator artifact must be an explicit absolute path")
	}
	expectedArtifactHash, err := hex.DecodeString(options.ArtifactSHA256)
	if err != nil || len(expectedArtifactHash) != sha256.Size {
		return Result{}, fmt.Errorf("validator artifact SHA-256 must be explicit")
	}
	if options.Timeout <= 0 || options.MaxInput <= 0 || options.MaxOutput <= 0 {
		return Result{}, fmt.Errorf("timeout and input/output limits must be positive")
	}
	if inputPath == "" {
		return Result{}, fmt.Errorf("input PDF path is required")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return Result{}, fmt.Errorf("open validation input: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return Result{}, fmt.Errorf("inspect validation input: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Result{}, fmt.Errorf("validation input must be a regular file")
	}
	if info.Size() > options.MaxInput {
		return Result{}, fmt.Errorf("validation input limit exceeded")
	}

	tempDir, err := os.MkdirTemp(options.TempDir, "pdfua-validation-")
	if err != nil {
		return Result{}, fmt.Errorf("create isolated validation directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	artifact, err := os.Open(options.Artifact)
	if err != nil {
		return Result{}, fmt.Errorf("open validator artifact: %w", err)
	}
	artifactInfo, statErr := artifact.Stat()
	if statErr != nil {
		artifact.Close()
		return Result{}, fmt.Errorf("inspect validator artifact: %w", statErr)
	}
	if !artifactInfo.Mode().IsRegular() {
		artifact.Close()
		return Result{}, fmt.Errorf("validator artifact must be a regular file")
	}
	isolatedArtifact := filepath.Join(tempDir, "validator.jar")
	artifactCopy, err := os.OpenFile(isolatedArtifact, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		artifact.Close()
		return Result{}, fmt.Errorf("create isolated validator artifact: %w", err)
	}
	artifactHasher := sha256.New()
	_, artifactCopyErr := io.Copy(io.MultiWriter(artifactCopy, artifactHasher), artifact)
	artifactCopyCloseErr := artifactCopy.Close()
	artifactCloseErr := artifact.Close()
	if artifactCopyErr != nil {
		return Result{}, fmt.Errorf("copy isolated validator artifact: %w", artifactCopyErr)
	}
	if artifactCopyCloseErr != nil || artifactCloseErr != nil {
		return Result{}, fmt.Errorf("close isolated validator artifact")
	}
	actualArtifactHash := artifactHasher.Sum(nil)
	if !bytes.Equal(actualArtifactHash, expectedArtifactHash) {
		return Result{}, fmt.Errorf("validator artifact digest mismatch")
	}
	isolatedInput := filepath.Join(tempDir, "input.pdf")
	copy, err := os.OpenFile(isolatedInput, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return Result{}, fmt.Errorf("create isolated validation input: %w", err)
	}
	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(copy, hasher), io.LimitReader(input, options.MaxInput+1))
	closeErr := copy.Close()
	if copyErr != nil {
		return Result{}, fmt.Errorf("copy isolated validation input: %w", copyErr)
	}
	if closeErr != nil {
		return Result{}, fmt.Errorf("close isolated validation input: %w", closeErr)
	}
	if written > options.MaxInput {
		return Result{}, fmt.Errorf("validation input limit exceeded")
	}

	ctx, cancel := context.WithTimeout(parent, options.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, options.Executable, "-Dfile.encoding=UTF8", "-XX:+IgnoreUnrecognizedVMOptions", "--add-exports=java.base/sun.security.pkcs=ALL-UNNAMED", "-jar", isolatedArtifact, "--flavour", "ua2", "--format", "xml", isolatedInput)
	cmd.Dir = tempDir
	budget := &outputBudget{remaining: options.MaxOutput}
	stdout := &boundedBuffer{budget: budget}
	stderr := &boundedBuffer{budget: budget}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	runErr := cmd.Run()
	if stdout.over || stderr.over {
		return Result{}, fmt.Errorf("validator output limit exceeded")
	}
	if ctx.Err() != nil {
		return Result{}, fmt.Errorf("validator timeout or cancellation: %w", ctx.Err())
	}
	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			return Result{}, fmt.Errorf("validator execution failed")
		}
		exitCode = exitErr.ExitCode()
	}
	profile, compliant, err := parseReport(stdout.buffer.Bytes())
	if err != nil {
		return Result{}, fmt.Errorf("invalid validator XML report: %w", err)
	}
	if profile != ua2ProfileName {
		return Result{}, fmt.Errorf("validator report profile mismatch")
	}
	reportHash := sha256.Sum256(stdout.buffer.Bytes())
	reportSHA256 := hex.EncodeToString(reportHash[:])
	switch exitCode {
	case 0:
		if !compliant {
			return Result{}, fmt.Errorf("validator exit/report mismatch")
		}
		return Result{ValidatorVersion: options.ValidatorVersion, Profile: profile, Outcome: OutcomePass, ExitCode: exitCode, ArtifactSHA256: hex.EncodeToString(actualArtifactHash), CandidateSHA256: hex.EncodeToString(hasher.Sum(nil)), ReportSHA256: reportSHA256}, nil
	case 1:
		if compliant {
			return Result{}, fmt.Errorf("validator exit/report mismatch")
		}
		return Result{ValidatorVersion: options.ValidatorVersion, Profile: profile, Outcome: OutcomeFinding, ExitCode: exitCode, ArtifactSHA256: hex.EncodeToString(actualArtifactHash), CandidateSHA256: hex.EncodeToString(hasher.Sum(nil)), ReportSHA256: reportSHA256}, nil
	default:
		return Result{}, fmt.Errorf("validator returned processing exit code %d", exitCode)
	}
}

func parseReport(data []byte) (string, bool, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root struct {
		XMLName xml.Name `xml:"report"`
		Jobs    struct {
			Job []struct {
				ValidationReports []struct {
					ProfileName string `xml:"profileName,attr"`
					Compliant   string `xml:"isCompliant,attr"`
				} `xml:"validationReport"`
			} `xml:"job"`
		} `xml:"jobs"`
		BatchSummaries []struct {
			TotalJobs         int `xml:"totalJobs,attr"`
			FailedToParse     int `xml:"failedToParse,attr"`
			Encrypted         int `xml:"encrypted,attr"`
			OutOfMemory       int `xml:"outOfMemory,attr"`
			VeraExceptions    int `xml:"veraExceptions,attr"`
			ValidationReports []struct {
				Compliant    int `xml:"compliant,attr"`
				NonCompliant int `xml:"nonCompliant,attr"`
				FailedJobs   int `xml:"failedJobs,attr"`
				Count        int `xml:",chardata"`
			} `xml:"validationReports"`
		} `xml:"batchSummary"`
	}
	if err := decoder.Decode(&root); err != nil {
		return "", false, err
	}
	if root.XMLName.Local != "report" || len(root.Jobs.Job) != 1 || len(root.Jobs.Job[0].ValidationReports) != 1 || len(root.BatchSummaries) != 1 {
		return "", false, fmt.Errorf("report must contain exactly one job, one validationReport, and one batchSummary")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return "", false, fmt.Errorf("unexpected data after validation report")
	}
	report := root.Jobs.Job[0].ValidationReports[0]
	if strings.TrimSpace(report.ProfileName) == "" {
		return "", false, fmt.Errorf("validationReport profileName is missing")
	}
	var compliant bool
	switch report.Compliant {
	case "true":
		compliant = true
	case "false":
		compliant = false
	default:
		return "", false, fmt.Errorf("validationReport isCompliant must be true or false")
	}
	summary := root.BatchSummaries[0]
	if summary.TotalJobs != 1 || summary.FailedToParse != 0 || summary.Encrypted != 0 || summary.OutOfMemory != 0 || summary.VeraExceptions != 0 || len(summary.ValidationReports) != 1 {
		return "", false, fmt.Errorf("batchSummary reports a processing failure or unexpected job count")
	}
	validationSummary := summary.ValidationReports[0]
	wantCompliant, wantNonCompliant := 0, 1
	if compliant {
		wantCompliant, wantNonCompliant = 1, 0
	}
	if validationSummary.Compliant != wantCompliant || validationSummary.NonCompliant != wantNonCompliant || validationSummary.FailedJobs != 0 || validationSummary.Count != 1 {
		return "", false, fmt.Errorf("batchSummary and validationReport disagree")
	}
	return report.ProfileName, compliant, nil
}
