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

func TestVerifyBundleReturnsDigestsForExactEvidence(t *testing.T) {
	passReport := bundleTestReport("true")
	findingReport := bundleTestReport("false")
	candidateDigest := strings.Repeat("a", 64)
	passMetadata, err := json.Marshal(bundleForControl(passReport, 0, "review-candidate", candidateDigest))
	if err != nil {
		t.Fatal(err)
	}
	findingMetadata, err := json.Marshal(bundleForControl(findingReport, 1, "review-candidate", candidateDigest))
	if err != nil {
		t.Fatal(err)
	}

	passResult, err := VerifyBundle(passMetadata, []byte(passReport))
	if err != nil {
		t.Fatal(err)
	}
	findingResult, err := VerifyBundle(findingMetadata, []byte(findingReport))
	if err != nil {
		t.Fatal(err)
	}

	if passResult.BundleMetadataSHA256 != reportDigest(passMetadata) {
		t.Fatalf("pass bundle metadata digest = %q; want digest of exact verified metadata %q", passResult.BundleMetadataSHA256, reportDigest(passMetadata))
	}
	if findingResult.BundleMetadataSHA256 != reportDigest(findingMetadata) {
		t.Fatalf("finding bundle metadata digest = %q; want digest of exact verified metadata %q", findingResult.BundleMetadataSHA256, reportDigest(findingMetadata))
	}
	if passResult.ValidatorVersion != findingResult.ValidatorVersion {
		t.Fatal("test reports must retain the same validator version")
	}
	if passResult.ReportSHA256 == findingResult.ReportSHA256 || passResult.BundleMetadataSHA256 == findingResult.BundleMetadataSHA256 || passResult.Outcome == findingResult.Outcome {
		t.Fatal("materially different same-version machine evidence must have distinct report identity, bundle identity, and outcome")
	}
	if passResult.CandidateSHA256 != findingResult.CandidateSHA256 || passResult.CandidateSHA256 != candidateDigest {
		t.Fatal("machine evidence identity comparison must retain the exact same candidate linkage")
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

func TestVerifyBundleRejectsWrongApplicationVersionWithRecomputedReportDigest(t *testing.T) {
	wrongVersionReport := strings.Replace(bundleTestReport("true"), `id="apps" version="1.30.3"`, `id="apps" version="1.30.2"`, 1)
	if wrongVersionReport == bundleTestReport("true") {
		t.Fatal("test fixture did not contain the apps release entry")
	}
	if _, err := VerifyBundle(validBundle(wrongVersionReport, 0), []byte(wrongVersionReport)); err == nil || !strings.Contains(err.Error(), "application version mismatch") {
		t.Fatalf("VerifyBundle accepted report app version 1.30.2 with a recomputed digest: %v", err)
	}
}

func TestVerifyBundleRejectsPaddedApplicationIdentityWithRecomputedReportDigest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		marker string
		padded string
	}{
		{name: "ASCII", marker: `id="apps"`, padded: `id=" apps "`},
		{name: "Unicode", marker: `id="apps"`, padded: "id=\"\u00a0apps\u2003\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := strings.Replace(bundleTestReport("true"), tc.marker, tc.padded, 1)
			if report == bundleTestReport("true") {
				t.Fatal("test fixture did not contain the apps release entry")
			}
			if _, err := VerifyBundle(validBundle(report, 0), []byte(report)); err == nil {
				t.Fatal("VerifyBundle accepted a padded apps component ID with a recomputed report digest")
			}
		})
	}
}

func TestVerifyBundleRejectsPaddedApplicationVersionWithRecomputedReportDigest(t *testing.T) {
	report := strings.Replace(bundleTestReport("true"), `id="apps" version="1.30.3"`, `id="apps" version=" 1.30.3 "`, 1)
	if report == bundleTestReport("true") {
		t.Fatal("test fixture did not contain the apps release entry")
	}
	if _, err := VerifyBundle(validBundle(report, 0), []byte(report)); err == nil {
		t.Fatal("VerifyBundle accepted a padded app version with a recomputed report digest")
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

func TestVerifyBundleRejectsCaseInsensitiveFieldAliases(t *testing.T) {
	reportXML := bundleTestReport("true")
	base := string(validBundle(reportXML, 0))
	aliases := []struct {
		name, key, alias, good, bad string
	}{
		{"schema version", "schema_version", "SCHEMA_VERSION", `1`, `99`},
		{"validator", "validator", "Validator", `"veraPDF"`, `"other"`},
		{"validator version", "validator_version", "Validator_Version", `"1.30.3"`, `"1.30.2"`},
		{"validator digest", "validator_sha256", "Validator_SHA256", `"` + expectedCLIHash + `"`, `"` + strings.Repeat("0", 64) + `"`},
		{"base image", "base_image", "Base_Image", `"` + expectedBaseImage + `"`, `"other"`},
		{"container image", "container_image_id", "Container_Image_Id", `"sha256:` + strings.Repeat("f", 64) + `"`, `"latest"`},
		{"control", "control", "Control", `"known-pass"`, `"unknown"`},
		{"profile", "profile", "Profile", `"` + UA2ProfileName + `"`, `"PDF/UA-1"`},
		{"exit code", "exit_code", "EXIT_CODE", `0`, `137`},
		{"candidate digest", "candidate_sha256", "Candidate_SHA256", `"` + passControlHash + `"`, `"` + strings.Repeat("0", 64) + `"`},
		{"report digest", "report_sha256", "Report_SHA256", "\"" + reportDigest([]byte(reportXML)) + "\"", `"` + strings.Repeat("0", 64) + `"`},
	}
	for _, tc := range aliases {
		t.Run(tc.name, func(t *testing.T) {
			needle := `"` + tc.key + `":` + tc.good
			replacement := needle + `,"` + tc.alias + `":` + tc.good
			// Make the canonical field fail and put its case-insensitive alias last.
			if tc.key == "exit_code" {
				replacement = `"exit_code":137,"EXIT_CODE":0`
			} else {
				replacement = `"` + tc.key + `":` + tc.bad + `,"` + tc.alias + `":` + tc.good
			}
			attacked := strings.Replace(base, needle, replacement, 1)
			if attacked == base {
				t.Fatal("test did not mutate the intended canonical field")
			}
			if _, err := VerifyBundle([]byte(attacked), []byte(reportXML)); err == nil {
				t.Fatal("VerifyBundle accepted a noncanonical case-insensitive field alias")
			}
		})
	}
}

func TestVerifyBundleRequiresPresentNonNullFields(t *testing.T) {
	reportXML := bundleTestReport("true")
	base := validBundle(reportXML, 0)
	for _, key := range []string{
		"schema_version", "validator", "validator_version", "validator_sha256", "base_image",
		"container_image_id", "control", "profile", "exit_code", "candidate_sha256", "report_sha256",
	} {
		t.Run(key, func(t *testing.T) {
			missing := mutateRawBundle(t, base, key, nil)
			if _, err := VerifyBundle(missing, []byte(reportXML)); err == nil {
				t.Fatal("VerifyBundle accepted missing required field")
			}
			null := mutateRawBundle(t, base, key, []byte("null"))
			if _, err := VerifyBundle(null, []byte(reportXML)); err == nil {
				t.Fatal("VerifyBundle accepted null required field")
			}
		})
	}
}

func TestVerifyControlBundleBindsOracleNameDigestAndOutcome(t *testing.T) {
	passReport := bundleTestReport("true")
	findingReport := bundleTestReport("false")
	passAsFinding, _ := json.Marshal(bundleForControl(passReport, 0, "known-finding", findingControlHash))
	findingAsPass, _ := json.Marshal(bundleForControl(findingReport, 1, "known-pass", passControlHash))
	if _, err := VerifyBundle(passAsFinding, []byte(passReport)); err != nil {
		t.Fatalf("generic verifier rejected an otherwise valid candidate report: %v", err)
	}
	if _, err := VerifyControlBundle(passAsFinding, []byte(passReport), "known-finding"); err == nil {
		t.Fatal("accepted compliant report in the known-finding control slot")
	}
	if _, err := VerifyBundle(findingAsPass, []byte(findingReport)); err != nil {
		t.Fatalf("generic verifier rejected an otherwise valid candidate report: %v", err)
	}
	if _, err := VerifyControlBundle(findingAsPass, []byte(findingReport), "known-finding"); err == nil {
		t.Fatal("accepted known-pass metadata for the known-finding control slot")
	}

	other, _ := json.Marshal(bundleForControl(passReport, 0, "candidate-17", strings.Repeat("c", 64)))
	if _, err := VerifyBundle(other, []byte(passReport)); err != nil {
		t.Fatalf("generic verifier rejected a valid non-oracle candidate: %v", err)
	}
}

func TestVerifyCandidateBundleBindsExactRepairAndPreservesCompletedFindings(t *testing.T) {
	for _, tc := range []struct {
		name, compliant, candidate, digest string
		exitCode                           int
		want                               Outcome
	}{
		{"statement finding", "false", "financial-v2-statement", statementV2Candidate, 1, OutcomeFinding},
		{"invoice pass", "true", "financial-v2-invoice", invoiceV2Candidate, 0, OutcomePass},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := bundleTestReport(tc.compliant)
			metadata, _ := json.Marshal(bundleForControl(report, tc.exitCode, tc.candidate, tc.digest))
			result, err := VerifyCandidateBundle(metadata, []byte(report), tc.candidate)
			if err != nil || result.Outcome != tc.want || result.CandidateSHA256 != tc.digest {
				t.Fatalf("VerifyCandidateBundle() = %#v, %v", result, err)
			}
		})
	}
	wrongHash, _ := json.Marshal(bundleForControl(bundleTestReport("true"), 0, "financial-v2-statement", invoiceV2Candidate))
	if _, err := VerifyCandidateBundle(wrongHash, []byte(bundleTestReport("true")), "financial-v2-statement"); err == nil {
		t.Fatal("accepted invoice candidate bytes in the statement slot")
	}
	wrongName, _ := json.Marshal(bundleForControl(bundleTestReport("true"), 0, "financial-v2-invoice", statementV2Candidate))
	if _, err := VerifyCandidateBundle(wrongName, []byte(bundleTestReport("true")), "financial-v2-statement"); err == nil {
		t.Fatal("accepted invoice metadata in the statement slot")
	}
	unknownCandidate, _ := json.Marshal(bundleForControl(bundleTestReport("true"), 0, "financial-v2-unknown", strings.Repeat("a", 64)))
	if _, err := VerifyCandidateBundle(unknownCandidate, []byte(bundleTestReport("true")), "financial-v2-unknown"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unknown candidate error = %v, want unsupported candidate", err)
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
			result, err := VerifyControlBundle(metadata, reportXML, tc.name)
			if err != nil || result.Outcome != tc.want {
				t.Fatalf("CI report outcome = %q, %v; want %q", result.Outcome, err, tc.want)
			}
		})
	}
}

func TestVerifyDownloadedCIFinancialRepairBundles(t *testing.T) {
	root := os.Getenv("PDFUA_BUNDLE_DIR")
	if root == "" {
		t.Skip("set PDFUA_BUNDLE_DIR to verify reports from the isolated CI container")
	}
	for _, name := range []string{"financial-v2-statement", "financial-v2-invoice"} {
		t.Run(name, func(t *testing.T) {
			metadata, err := os.ReadFile(filepath.Join(root, name, "contract.json"))
			if err != nil {
				t.Fatal("CI validator did not provide candidate contract metadata")
			}
			reportXML, err := os.ReadFile(filepath.Join(root, name, "report.xml"))
			if err != nil {
				t.Fatal("CI validator did not provide candidate report")
			}
			result, err := VerifyCandidateBundle(metadata, reportXML, name)
			if err != nil {
				t.Fatalf("CI candidate report failed verification: %v", err)
			}
			if result.Outcome != OutcomePass && result.Outcome != OutcomeFinding {
				t.Fatalf("unexpected candidate result outcome %q", result.Outcome)
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

func bundleForControl(reportXML string, exitCode int, control, candidateHash string) ValidationBundle {
	reportHash := sha256.Sum256([]byte(reportXML))
	return ValidationBundle{
		SchemaVersion: 1, Validator: "veraPDF", ValidatorVersion: "1.30.3",
		ValidatorSHA256: expectedCLIHash, BaseImage: expectedBaseImage, ContainerImageID: "sha256:" + strings.Repeat("f", 64),
		Control: control, Profile: UA2ProfileName, ExitCode: exitCode, CandidateSHA256: candidateHash,
		ReportSHA256: hex.EncodeToString(reportHash[:]),
	}
}

func bundleTestReport(compliant string) string {
	compliantCount, noncompliantCount := "0", "1"
	if compliant == "true" {
		compliantCount, noncompliantCount = "1", "0"
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><buildInformation><releaseDetails id="core" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/></buildInformation><jobs><job><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="` + compliant + `"></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + compliantCount + `" nonCompliant="` + noncompliantCount + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

func reportDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func mutateRawBundle(t *testing.T, data []byte, key string, value []byte) []byte {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if value == nil {
		delete(fields, key)
	} else {
		fields[key] = value
	}
	updated, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return updated
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
