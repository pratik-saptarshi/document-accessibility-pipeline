package pdfua

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyBundleAcceptsExplicitUA2PassAndFinding(t *testing.T) {
	for _, tc := range []struct {
		name, compliant string
		exitCode        int
		want            Outcome
	}{
		{"pass", "true", 0, OutcomePass},
		{"finding", "false", 1, OutcomeFinding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := bundleTestReport(tc.compliant)
			bundle := validBundle(report, tc.exitCode)
			result, err := VerifyBundle(bundle, []byte(report))
			if err != nil || result.Outcome != tc.want || result.Profile != UA2ProfileName {
				t.Fatalf("VerifyBundle() = %#v, %v; want %q", result, err, tc.want)
			}
		})
	}
}

func TestVerifyBundleFailsClosedForTamperingAndMismatch(t *testing.T) {
	reportXML := bundleTestReport("true")
	base := validBundle(reportXML, 0)
	tests := []struct {
		name   string
		bundle []byte
		report []byte
	}{
		{"report tamper", base, []byte(strings.Replace(reportXML, `isCompliant="true"`, `isCompliant="false"`, 1))},
		{"wrong CLI digest", mutateBundle(t, base, func(b *ValidationBundle) { b.ValidatorSHA256 = strings.Repeat("0", 64) }), []byte(reportXML)},
		{"wrong candidate digest", mutateBundle(t, base, func(b *ValidationBundle) { b.CandidateSHA256 = "not-a-digest" }), []byte(reportXML)},
		{"wrong image identity", mutateBundle(t, base, func(b *ValidationBundle) { b.ContainerImageID = "latest" }), []byte(reportXML)},
		{"wrong base image", mutateBundle(t, base, func(b *ValidationBundle) { b.BaseImage = "eclipse-temurin:latest" }), []byte(reportXML)},
		{"wrong profile", mutateBundle(t, base, func(b *ValidationBundle) { b.Profile = "PDF/UA-1" }), []byte(reportXML)},
		{"exit/report mismatch", mutateBundle(t, base, func(b *ValidationBundle) { b.ExitCode = 1 }), []byte(reportXML)},
		{"unexpected validator", mutateBundle(t, base, func(b *ValidationBundle) { b.Validator = "other" }), []byte(reportXML)},
		{"unsupported schema", mutateBundle(t, base, func(b *ValidationBundle) { b.SchemaVersion = 99 }), []byte(reportXML)},
		{"unknown field", append(append([]byte(nil), base[:len(base)-1]...), []byte(`,"extra":true}`)...), []byte(reportXML)},
		{"duplicate field", append(append([]byte(nil), base[:len(base)-1]...), []byte(`,"exit_code":0}`)...), []byte(reportXML)},
		{"trailing JSON", append(append([]byte(nil), base...), []byte(` {}`)...), []byte(reportXML)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := VerifyBundle(tc.bundle, tc.report); err == nil {
				t.Fatal("VerifyBundle accepted invalid or tampered bundle")
			}
		})
	}
}

func TestVerifyBundleRejectsOperationalAndReportErrors(t *testing.T) {
	reportXML := bundleTestReport("true")
	bundle := validBundle(reportXML, 0)
	var metadata ValidationBundle
	if err := json.Unmarshal(bundle, &metadata); err != nil {
		t.Fatal(err)
	}
	metadata.ExitCode = 137
	bundle, _ = json.Marshal(metadata)
	if _, err := VerifyBundle(bundle, []byte(reportXML)); err == nil {
		t.Fatal("accepted validator operational failure exit code")
	}
	metadata.ExitCode = 0
	metadata.ReportSHA256 = strings.Repeat("0", 64)
	bundle, _ = json.Marshal(metadata)
	if _, err := VerifyBundle(bundle, []byte(reportXML)); err == nil {
		t.Fatal("accepted report digest mismatch")
	}
	if _, err := VerifyBundle(validBundle(reportXML, 0), []byte(strings.Repeat("x", maxBundleReport+1))); err == nil {
		t.Fatal("accepted oversized report")
	}
}

func TestVerifyDownloadedCIControlBundles(t *testing.T) {
	root := os.Getenv("PDFUA_BUNDLE_DIR")
	if root == "" {
		t.Skip("set PDFUA_BUNDLE_DIR to verify reports from the isolated CI container")
	}
	for _, tc := range []struct {
		name string
		want Outcome
	}{
		{"known-pass", OutcomePass},
		{"known-finding", OutcomeFinding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadata, err := os.ReadFile(filepath.Join(root, tc.name, "contract.json"))
			if err != nil {
				t.Fatal("CI validator did not provide its contract metadata")
			}
			reportXML, err := os.ReadFile(filepath.Join(root, tc.name, "report.xml"))
			if err != nil {
				t.Fatal("CI validator did not provide its report")
			}
			result, err := VerifyBundle(metadata, reportXML)
			if err != nil || result.Outcome != tc.want {
				t.Fatalf("CI report outcome = %q, %v; want %q", result.Outcome, err, tc.want)
			}
		})
	}
}

func validBundle(reportXML string, exitCode int) []byte {
	reportHash := sha256.Sum256([]byte(reportXML))
	bundle := ValidationBundle{
		SchemaVersion: 1, Validator: "veraPDF", ValidatorVersion: "1.30.3",
		ValidatorSHA256: expectedCLIHash, BaseImage: expectedBaseImage, ContainerImageID: "sha256:" + strings.Repeat("f", 64),
		Profile: UA2ProfileName, ExitCode: exitCode,
		ReportSHA256: hex.EncodeToString(reportHash[:]),
	}
	if exitCode == 0 {
		bundle.Control = "known-pass"
		bundle.CandidateSHA256 = passControlHash
	} else {
		bundle.Control = "known-finding"
		bundle.CandidateSHA256 = findingControlHash
	}
	data, _ := json.Marshal(bundle)
	return data
}

func bundleTestReport(compliant string) string {
	compliantCount, noncompliantCount := "0", "1"
	if compliant == "true" {
		compliantCount, noncompliantCount = "1", "0"
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><jobs><job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="` + compliant + `"></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + compliantCount + `" nonCompliant="` + noncompliantCount + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

func mutateBundle(t *testing.T, data []byte, mutate func(*ValidationBundle)) []byte {
	t.Helper()
	var b ValidationBundle
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}
	mutate(&b)
	updated, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return updated
}
