package pdfua

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// checkXML is one failed check: context and error message.
type checkXML struct{ context, message string }

// ruleXML is one explicitly constructed failed rule. Tests build rules
// directly so duplicate checks, split rules, and identity collisions are
// exercised exactly as written — nothing is grouped or inferred.
type ruleXML struct {
	clause, testNumber, object, test string
	checks                           []checkXML
}

func buildReport(compliant bool, rules []ruleXML) string {
	compliantText, compliantCount, noncompliantCount := "false", "0", "1"
	if compliant {
		compliantText, compliantCount, noncompliantCount = "true", "1", "0"
	}
	failedChecks := 0
	var body strings.Builder
	for _, rule := range rules {
		failedChecks += len(rule.checks)
		var checks strings.Builder
		for _, check := range rule.checks {
			fmt.Fprintf(&checks, `<check status="failed"><context>%s</context><errorMessage>%s</errorMessage></check>`, check.context, check.message)
		}
		fmt.Fprintf(&body, `<rule specification="ISO 14289-2:2024" clause="%s" testNumber="%s" status="failed" failedChecks="%d"><description>rule description</description><object>%s</object><test>%s</test>%s</rule>`,
			rule.clause, rule.testNumber, len(rule.checks), rule.object, rule.test, checks.String())
	}
	return `<?xml version="1.0" encoding="UTF-8"?><report><buildInformation><releaseDetails id="core" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/></buildInformation><jobs><job><item><name>/input/document.pdf</name></item><validationReport profileName="PDF/UA-2 + Tagged PDF validation profile" isCompliant="` + compliantText + `"><details passedRules="0" failedRules="` + fmt.Sprint(len(rules)) + `" passedChecks="0" failedChecks="` + fmt.Sprint(failedChecks) + `">` + body.String() + `</details></validationReport></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + compliantCount + `" nonCompliant="` + noncompliantCount + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

var (
	ruleA = ruleXML{"8.11.1", "2", "XMPDoc", "hasMetadata == false", []checkXML{{"ctx-a", "missing XMP"}}}
	ruleB = ruleXML{"8.4.5.8", "1", "Font", "isMapped == false", []checkXML{{"ctx-b", "no Unicode mapping"}}}
	ruleC = ruleXML{"8.2.2", "1", "Content", "isTagged == false", []checkXML{{"ctx-c", "untagged content"}}}
)

func TestParseFailedChecksEnumeratesEveryFailedCheck(t *testing.T) {
	// One rule carrying two byte-identical failed checks: both count.
	dup := ruleXML{"8.4.5.8", "1", "Font", "isMapped == false", []checkXML{{"ctx-b", "no Unicode mapping"}, {"ctx-b", "no Unicode mapping"}}}
	findings, err := ParseFailedChecks([]byte(buildReport(false, []ruleXML{ruleA, dup})))
	if err != nil {
		t.Fatalf("ParseFailedChecks() error = %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("ParseFailedChecks() returned %d findings, want 3 (duplicate checks counted individually)", len(findings))
	}
	if findings[0].Clause != "8.11.1" || findings[0].Context != "ctx-a" || findings[1].ErrorMessage != "no Unicode mapping" || findings[1] != findings[2] {
		t.Fatalf("unexpected findings: %#v", findings)
	}
	pass, err := ParseFailedChecks([]byte(buildReport(true, nil)))
	if err != nil || len(pass) != 0 {
		t.Fatalf("compliant ParseFailedChecks() = %v, %v; want no findings", pass, err)
	}
}

func TestParseFailedChecksFailsClosed(t *testing.T) {
	valid := buildReport(false, []ruleXML{ruleA})
	cases := map[string]string{
		"noncompliant without details":       bundleTestReport("false"),
		"declared failedChecks disagree":     strings.Replace(valid, `failedChecks="1"><rule`, `failedChecks="2"><rule`, 1),
		"declared failedRules disagree":      strings.Replace(valid, `failedRules="1"`, `failedRules="2"`, 1),
		"rule failedChecks disagree":         strings.Replace(valid, `status="failed" failedChecks="1"`, `status="failed" failedChecks="2"`, 1),
		"duplicate details element":          strings.Replace(valid, `</details>`, `</details><details passedRules="0" failedRules="0" passedChecks="0" failedChecks="0"></details>`, 1),
		"duplicate object element":           strings.Replace(valid, `<object>XMPDoc</object>`, `<object>XMPDoc</object><object>Other</object>`, 1),
		"check missing errorMessage":         strings.Replace(valid, `<errorMessage>missing XMP</errorMessage>`, ``, 1),
		"unknown element in details":         strings.Replace(valid, `</details>`, `<extra/></details>`, 1),
		"unknown attribute on rule":          strings.Replace(valid, `<rule specification=`, `<rule bogus="1" specification=`, 1),
		"passed rule enumerated in details":  strings.Replace(valid, `status="failed" failedChecks="1"`, `status="passed" failedChecks="1"`, 1),
		"compliant declaring failed checks":  buildReport(true, []ruleXML{ruleA}),
		"noncompliant declaring zero failed": strings.Replace(buildReport(true, nil), `isCompliant="true"`, `isCompliant="false"`, 1),
		"malformed XML":                      "<report>",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseFailedChecks([]byte(input)); err == nil {
				t.Fatalf("ParseFailedChecks() accepted %s", name)
			}
		})
	}
}

func TestParseFailedChecksAcceptsRealCorpusReport(t *testing.T) {
	// The retained veraPDF-corpus report is a real captured report: it
	// ends with a trailing newline, carries a tags attribute on its rule,
	// and declares large passed counts that are not enumerated.
	data, err := os.ReadFile(filepath.Join("testdata", "known-finding", "report.xml"))
	if err != nil {
		t.Fatal(err)
	}
	findings, err := ParseFailedChecks(data)
	if err != nil {
		t.Fatalf("ParseFailedChecks(real corpus report) error = %v", err)
	}
	if len(findings) != 1 || findings[0].Clause != "8.2.4" || findings[0].TestNumber != "1" || findings[0].Object != "SENonStandard" {
		t.Fatalf("ParseFailedChecks(real corpus report) = %#v", findings)
	}
}

func TestClassifyFindingsMultiset(t *testing.T) {
	parse := func(rules []ruleXML) []Finding {
		t.Helper()
		findings, err := ParseFailedChecks([]byte(buildReport(false, rules)))
		if err != nil {
			t.Fatal(err)
		}
		return findings
	}
	// Damaged: A once, B twice (two identical checks in one rule).
	// Candidate: A once, B once, C once (new rule).
	dupB := ruleXML{"8.4.5.8", "1", "Font", "isMapped == false", []checkXML{{"ctx-b", "no Unicode mapping"}, {"ctx-b", "no Unicode mapping"}}}
	damaged := parse([]ruleXML{ruleA, dupB})
	candidate := parse([]ruleXML{ruleA, ruleB, ruleC})
	got := ClassifyFindings(damaged, candidate)
	if len(got.Inherited) != 2 || len(got.Introduced) != 1 || len(got.Resolved) != 1 {
		t.Fatalf("ClassifyFindings() = inherited %d, introduced %d, resolved %d; want 2/1/1",
			len(got.Inherited), len(got.Introduced), len(got.Resolved))
	}
	if got.Introduced[0].Clause != "8.2.2" || got.Resolved[0].Clause != "8.4.5.8" {
		t.Fatalf("unexpected introduced/resolved: %#v / %#v", got.Introduced, got.Resolved)
	}
}

func TestClassifyFindingsIdentityCollisionIsIntroduced(t *testing.T) {
	// Same rule, object, test, and context, but a different error message
	// is a different finding: the candidate occurrence is introduced and
	// the damaged occurrence is resolved, never silently inherited.
	parse := func(rules []ruleXML) []Finding {
		t.Helper()
		findings, err := ParseFailedChecks([]byte(buildReport(false, rules)))
		if err != nil {
			t.Fatal(err)
		}
		return findings
	}
	altered := ruleXML{"8.11.1", "2", "XMPDoc", "hasMetadata == false", []checkXML{{"ctx-a", "a different message"}}}
	got := ClassifyFindings(parse([]ruleXML{ruleA}), parse([]ruleXML{altered}))
	if len(got.Inherited) != 0 || len(got.Introduced) != 1 || len(got.Resolved) != 1 {
		t.Fatalf("ClassifyFindings() = inherited %d, introduced %d, resolved %d; want 0/1/1",
			len(got.Inherited), len(got.Introduced), len(got.Resolved))
	}
	// Whitespace wrapping of the same text is the same finding.
	wrapped := ruleXML{"8.11.1", "2", "XMPDoc", "hasMetadata == false", []checkXML{{"ctx-a", "missing\n   XMP"}}}
	spaced := ruleXML{"8.11.1", "2", "XMPDoc", "hasMetadata == false", []checkXML{{"ctx-a", "missing XMP"}}}
	got = ClassifyFindings(parse([]ruleXML{wrapped}), parse([]ruleXML{spaced}))
	if len(got.Inherited) != 1 || len(got.Introduced) != 0 {
		t.Fatalf("whitespace variant: ClassifyFindings() = inherited %d, introduced %d; want 1/0",
			len(got.Inherited), len(got.Introduced))
	}
}

func TestAssessEligibilityZeroIntroducedRequired(t *testing.T) {
	damaged := buildReport(false, []ruleXML{ruleA, ruleB})
	eligible, err := AssessEligibility([]byte(damaged), []byte(buildReport(false, []ruleXML{ruleA, ruleB})))
	if err != nil || !eligible.Eligible {
		t.Fatalf("identical findings: AssessEligibility() = %#v, %v; want eligible", eligible, err)
	}
	resolved, err := AssessEligibility([]byte(damaged), []byte(buildReport(false, []ruleXML{ruleA})))
	if err != nil || !resolved.Eligible || len(resolved.Classification.Resolved) != 1 {
		t.Fatalf("resolved finding: AssessEligibility() = %#v, %v; want eligible with one resolved", resolved, err)
	}
	introduced, err := AssessEligibility([]byte(damaged), []byte(buildReport(false, []ruleXML{ruleA, ruleB, ruleC})))
	if err != nil || introduced.Eligible || len(introduced.Classification.Introduced) != 1 {
		t.Fatalf("introduced finding: AssessEligibility() = %#v, %v; want ineligible with one introduced", introduced, err)
	}
	compliant, err := AssessEligibility([]byte(damaged), []byte(buildReport(true, nil)))
	if err != nil || !compliant.Eligible {
		t.Fatalf("compliant candidate: AssessEligibility() = %#v, %v; want eligible", compliant, err)
	}
}

func TestAssessEligibilityRejectsUnpinnedProfileOrVersion(t *testing.T) {
	// Decision 0026 requires a pinned ua2 run: both reports must carry the
	// explicit PDF/UA-2 profile and the pinned validator application
	// version. Substituting another profile (the Codex repro: a PDF/A-1B
	// profile string in both slots) or another application version must
	// fail closed instead of returning an eligibility verdict.
	valid := buildReport(false, []ruleXML{ruleA})
	wrongProfile := strings.Replace(valid, `profileName="`+UA2ProfileName+`"`, `profileName="PDF/A-1B validation profile"`, 1)
	wrongVersion := strings.Replace(valid, `id="apps" version="1.30.3"`, `id="apps" version="1.30.2"`, 1)
	if wrongProfile == valid || wrongVersion == valid {
		t.Fatal("test fixture substitution did not change the report")
	}
	cases := map[string]struct{ damaged, candidate string }{
		"candidate profile is not the pinned ua2 profile": {valid, wrongProfile},
		"damaged profile is not the pinned ua2 profile":   {wrongProfile, valid},
		"both profiles substituted":                       {wrongProfile, wrongProfile},
		"candidate application version is not pinned":     {valid, wrongVersion},
		"damaged application version is not pinned":       {wrongVersion, valid},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eligibility, err := AssessEligibility([]byte(tc.damaged), []byte(tc.candidate))
			if err == nil {
				t.Fatalf("AssessEligibility() = %#v, nil; want an error", eligibility)
			}
			if eligibility.Eligible {
				t.Fatalf("AssessEligibility() returned eligible for an unpinned report: %#v", eligibility)
			}
		})
	}
}

func TestClassifyFindingsIgnoresVolatileContext(t *testing.T) {
	// A finding's context is validator location data: content-stream
	// operator and glyph indices and the font-object identifier embedded
	// in a glyph context all change when the bounded repair rewrites the
	// content stream or the font dictionary, without changing the
	// finding. The same failed check (same specification, clause, test
	// number, description, object, test, and error message) at a shifted
	// context is the same finding.
	parse := func(rules []ruleXML) []Finding {
		t.Helper()
		findings, err := ParseFailedChecks([]byte(buildReport(false, rules)))
		if err != nil {
			t.Fatal(err)
		}
		return findings
	}
	damagedGlyph := ruleXML{"8.4.5.8", "1", "Glyph", "toUnicode != null", []checkXML{
		{"root/document[0]/pages[0](3 0 obj PDPage)/contentStream[0](4 0 obj PDSemanticContentStream)/operators[37]/usedGlyphs[0](Helvetica Helvetica 68 0 753631393 0 true)", "The glyph can not be mapped to Unicode"},
		{"root/document[0]/pages[0](3 0 obj PDPage)/contentStream[0](4 0 obj PDSemanticContentStream)/operators[37]/usedGlyphs[1](Helvetica Helvetica 97 0 753631393 0 true)", "The glyph can not be mapped to Unicode"},
	}}
	candidateGlyph := ruleXML{"8.4.5.8", "1", "Glyph", "toUnicode != null", []checkXML{
		{"root/document[0]/pages[0](3 0 obj PDPage)/contentStream[0](4 0 obj PDSemanticContentStream)/operators[41]/usedGlyphs[0](Helvetica Helvetica 68 0 1155769010 0 true)", "The glyph can not be mapped to Unicode"},
		{"root/document[0]/pages[0](3 0 obj PDPage)/contentStream[0](4 0 obj PDSemanticContentStream)/operators[41]/usedGlyphs[1](Helvetica Helvetica 97 0 1155769010 0 true)", "The glyph can not be mapped to Unicode"},
	}}
	got := ClassifyFindings(parse([]ruleXML{damagedGlyph}), parse([]ruleXML{candidateGlyph}))
	if len(got.Inherited) != 2 || len(got.Introduced) != 0 || len(got.Resolved) != 0 {
		t.Fatalf("ClassifyFindings() = inherited %d, introduced %d, resolved %d; want 2/0/0",
			len(got.Inherited), len(got.Introduced), len(got.Resolved))
	}
	// A genuinely new test under the same clause is still introduced:
	// context insensitivity must not widen into clause-level matching.
	newTest := ruleXML{"8.4.5.8", "2", "Glyph", "someOtherTest == true", []checkXML{{"ctx-new", "a new failure"}}}
	got = ClassifyFindings(parse([]ruleXML{damagedGlyph}), parse([]ruleXML{candidateGlyph, newTest}))
	if len(got.Inherited) != 2 || len(got.Introduced) != 1 || got.Introduced[0].TestNumber != "2" {
		t.Fatalf("ClassifyFindings() with a new test = inherited %d, introduced %#v; want 2 inherited and the new test introduced",
			len(got.Inherited), got.Introduced)
	}
}

func TestVerifyDamagedBundleBindsExactDamagedInput(t *testing.T) {
	report := bundleTestReport("false")
	metadata, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-statement-damaged", statementV2Damaged))
	result, err := VerifyDamagedBundle(metadata, []byte(report), "financial-v2-statement-damaged")
	if err != nil || result.CandidateSHA256 != statementV2Damaged || result.Outcome != OutcomeFinding {
		t.Fatalf("VerifyDamagedBundle() = %#v, %v", result, err)
	}
	wrongHash, _ := json.Marshal(bundleForControl(report, 1, "financial-v2-statement-damaged", statementV2Candidate))
	if _, err := VerifyDamagedBundle(wrongHash, []byte(report), "financial-v2-statement-damaged"); err == nil {
		t.Fatal("accepted repaired candidate bytes in the damaged slot")
	}
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
