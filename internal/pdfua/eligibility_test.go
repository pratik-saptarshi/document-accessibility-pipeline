package pdfua

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// detailedReport builds a report in the exact shape ParseReport accepts,
// with the given failed rules. Each rule entry is
// "clause|testNumber|object|test|context|errorMessage" and produces one
// failed check.
func detailedReport(compliant bool, failed ...string) string {
	compliantText, compliantCount, noncompliantCount := "false", "0", "1"
	if compliant {
		compliantText, compliantCount, noncompliantCount = "true", "1", "0"
	}
	var rules strings.Builder
	// Group consecutive entries sharing clause+testNumber into one rule.
	for i := 0; i < len(failed); {
		j := i
		parts := strings.SplitN(failed[i], "|", 6)
		var checks strings.Builder
		for j < len(failed) {
			p := strings.SplitN(failed[j], "|", 6)
			if p[0] != parts[0] || p[1] != parts[1] {
				break
			}
			fmt.Fprintf(&checks, `<check status="failed"><context>%s</context><errorMessage>%s</errorMessage></check>`, p[4], p[5])
			j++
		}
		fmt.Fprintf(&rules, `<rule specification="ISO 14289-2:2024" clause="%s" testNumber="%s" status="failed" failedChecks="%d"><description>desc</description><object>%s</object><test>%s</test>%s</rule>`,
			parts[0], parts[1], j-i, parts[2], parts[3], checks.String())
		i = j
	}
	failedRules, failedChecks := 0, len(failed)
	if len(failed) > 0 {
		// Count distinct clause|testNumber groups.
		seen := map[string]bool{}
		for _, f := range failed {
			p := strings.SplitN(f, "|", 6)
			if !seen[p[0]+"|"+p[1]] {
				seen[p[0]+"|"+p[1]] = true
				failedRules++
			}
		}
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><buildInformation><releaseDetails id="core" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/></buildInformation><jobs><job><item><name>/input/document.pdf</name></item><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="` + compliantText + `"><details passedRules="0" failedRules="` + fmt.Sprint(failedRules) + `" passedChecks="0" failedChecks="` + fmt.Sprint(failedChecks) + `">` + rules.String() + `</details></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + compliantCount + `" nonCompliant="` + noncompliantCount + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

const (
	findingA = "8.11.1|2|XMPDoc|hasMetadata == false|ctx-a|missing XMP"
	findingB = "8.4.5.8|1|Font|isMapped == false|ctx-b|no Unicode mapping"
	findingC = "8.2.2|1|Content|isTagged == false|ctx-c|untagged content"
)

func TestParseFailedChecksEnumeratesEveryFailedCheck(t *testing.T) {
	findings, err := ParseFailedChecks([]byte(detailedReport(false, findingA, findingB, findingB)))
	if err != nil {
		t.Fatalf("ParseFailedChecks() error = %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("ParseFailedChecks() returned %d findings, want 3", len(findings))
	}
	if findings[0].Clause != "8.11.1" || findings[0].Context != "ctx-a" || findings[1].ErrorMessage != "no Unicode mapping" {
		t.Fatalf("unexpected findings: %#v", findings)
	}
	pass, err := ParseFailedChecks([]byte(detailedReport(true)))
	if err != nil || len(pass) != 0 {
		t.Fatalf("compliant ParseFailedChecks() = %v, %v; want no findings", pass, err)
	}
}

func TestParseFailedChecksFailsClosed(t *testing.T) {
	// Noncompliant report without details cannot enumerate its findings.
	if _, err := ParseFailedChecks([]byte(bundleTestReport("false"))); err == nil {
		t.Fatal("accepted a noncompliant report with no failed-check details")
	}
	// Declared counts that disagree with the enumerated checks are rejected.
	broken := strings.Replace(detailedReport(false, findingA), `failedChecks="1"><rule`, `failedChecks="2"><rule`, 1)
	if _, err := ParseFailedChecks([]byte(broken)); err == nil {
		t.Fatal("accepted a report whose declared failedChecks count disagrees with its checks")
	}
	// Malformed XML is rejected.
	if _, err := ParseFailedChecks([]byte("<report>")); err == nil {
		t.Fatal("accepted malformed XML")
	}
}

func TestClassifyFindingsMultiset(t *testing.T) {
	fa, _ := ParseFailedChecks([]byte(detailedReport(false, findingA)))
	fb, _ := ParseFailedChecks([]byte(detailedReport(false, findingB)))
	fc, _ := ParseFailedChecks([]byte(detailedReport(false, findingC)))
	damaged := append(append([]Finding{}, fa...), fb...)
	damaged = append(damaged, fb...) // findingB twice in damaged
	candidate := append(append([]Finding{}, fa...), fb...)
	candidate = append(candidate, fc...) // findingC introduced
	got := ClassifyFindings(damaged, candidate)
	if len(got.Inherited) != 2 || len(got.Introduced) != 1 || len(got.Resolved) != 1 {
		t.Fatalf("ClassifyFindings() = inherited %d, introduced %d, resolved %d; want 2/1/1",
			len(got.Inherited), len(got.Introduced), len(got.Resolved))
	}
	if got.Introduced[0].Clause != "8.2.2" || got.Resolved[0].Clause != "8.4.5.8" {
		t.Fatalf("unexpected introduced/resolved: %#v / %#v", got.Introduced, got.Resolved)
	}
}

func TestAssessEligibilityZeroIntroducedRequired(t *testing.T) {
	damaged := detailedReport(false, findingA, findingB)
	eligible, err := AssessEligibility([]byte(damaged), []byte(detailedReport(false, findingA, findingB)))
	if err != nil || !eligible.Eligible {
		t.Fatalf("identical findings: AssessEligibility() = %#v, %v; want eligible", eligible, err)
	}
	// A candidate that resolves a finding stays eligible; compliance is not required.
	resolved, err := AssessEligibility([]byte(damaged), []byte(detailedReport(false, findingA)))
	if err != nil || !resolved.Eligible || len(resolved.Classification.Resolved) != 1 {
		t.Fatalf("resolved finding: AssessEligibility() = %#v, %v; want eligible with one resolved", resolved, err)
	}
	// One introduced finding makes the candidate ineligible.
	introduced, err := AssessEligibility([]byte(damaged), []byte(detailedReport(false, findingA, findingB, findingC)))
	if err != nil || introduced.Eligible || len(introduced.Classification.Introduced) != 1 {
		t.Fatalf("introduced finding: AssessEligibility() = %#v, %v; want ineligible with one introduced", introduced, err)
	}
	// A compliant candidate introduces nothing.
	compliant, err := AssessEligibility([]byte(damaged), []byte(detailedReport(true)))
	if err != nil || !compliant.Eligible {
		t.Fatalf("compliant candidate: AssessEligibility() = %#v, %v; want eligible", compliant, err)
	}
}

func TestVerifyDamagedBundleBindsExactDamagedInput(t *testing.T) {
	report := bundleTestReport("false")
	metadata, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-statement-damaged", statementV2Damaged))
	result, err := VerifyDamagedBundle(metadata, []byte(report), "financial-v2-statement-damaged")
	if err != nil || result.CandidateSHA256 != statementV2Damaged || result.Outcome != OutcomeFinding {
		t.Fatalf("VerifyDamagedBundle() = %#v, %v", result, err)
	}
	// Candidate bytes in the damaged slot are rejected.
	wrongHash, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-statement-damaged", statementV2Candidate))
	if _, err := VerifyDamagedBundle(wrongHash, []byte(report), "financial-v2-statement-damaged"); err == nil {
		t.Fatal("accepted repaired candidate bytes in the damaged slot")
	}
	// The other family's damaged control name is rejected.
	wrongName, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-invoice-damaged", invoiceV2Damaged))
	if _, err := VerifyDamagedBundle(wrongName, []byte(report), "financial-v2-statement-damaged"); err == nil {
		t.Fatal("accepted invoice damaged metadata in the statement slot")
	}
	unknown, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-unknown-damaged", strings.Repeat("b", 64)))
	if _, err := VerifyDamagedBundle(unknown, []byte(report), "financial-v2-unknown-damaged"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unknown damaged error = %v, want unsupported", err)
	}
}

// TestEligibilityFromDownloadedCIPairedBundles runs only in the pdfua2 CI
// job, where PDFUA_BUNDLE_DIR holds the fresh isolated-container bundles
// for both damaged inputs and both candidates. It is the paired-run
// evidence Decision 0026 requires: zero introduced findings per family.
func TestEligibilityFromDownloadedCIPairedBundles(t *testing.T) {
	root := os.Getenv("PDFUA_BUNDLE_DIR")
	if root == "" {
		t.Skip("set PDFUA_BUNDLE_DIR to verify paired reports from the isolated CI container")
	}
	for _, family := range []struct{ candidate, damaged string }{
		{"financial-v2-statement", "financial-v2-statement-damaged"},
		{"financial-v2-invoice", "financial-v2-invoice-damaged"},
	} {
		t.Run(family.candidate, func(t *testing.T) {
			read := func(control string) ([]byte, []byte) {
				metadata, err := os.ReadFile(filepath.Join(root, control, "contract.json"))
				if err != nil {
					t.Fatalf("CI validator did not provide %s contract metadata", control)
				}
				reportXML, err := os.ReadFile(filepath.Join(root, control, "report.xml"))
				if err != nil {
					t.Fatalf("CI validator did not provide %s report", control)
				}
				return metadata, reportXML
			}
			damagedMetadata, damagedReport := read(family.damaged)
			if _, err := VerifyDamagedBundle(damagedMetadata, damagedReport, family.damaged); err != nil {
				t.Fatalf("CI damaged report failed verification: %v", err)
			}
			candidateMetadata, candidateReport := read(family.candidate)
			if _, err := VerifyCandidateBundle(candidateMetadata, candidateReport, family.candidate); err != nil {
				t.Fatalf("CI candidate report failed verification: %v", err)
			}
			eligibility, err := AssessEligibility(damagedReport, candidateReport)
			if err != nil {
				t.Fatalf("paired eligibility assessment failed: %v", err)
			}
			t.Logf("%s: inherited=%d introduced=%d resolved=%d eligible=%v",
				family.candidate, len(eligibility.Classification.Inherited),
				len(eligibility.Classification.Introduced), len(eligibility.Classification.Resolved),
				eligibility.Eligible)
			if !eligibility.Eligible {
				t.Fatalf("candidate introduced %d finding(s) not present in its damaged input; first: %#v",
					len(eligibility.Classification.Introduced), eligibility.Classification.Introduced[0])
			}
		})
	}
}
