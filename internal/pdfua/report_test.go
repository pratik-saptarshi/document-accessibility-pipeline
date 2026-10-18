package pdfua

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func TestParseReportReturnsApplicationVersionForPassAndFinding(t *testing.T) {
	for _, tc := range []struct {
		name       string
		compliant  bool
		components string
	}{
		{name: "pass", compliant: true, components: `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`},
		{name: "finding", compliant: false, components: `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`},
		{name: "reordered components", compliant: true, components: `<releaseDetails id="validation-model" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="core" version="1.30.3"/>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := contractReport(tc.compliant, tc.components)
			profile, version, compliant, err := ParseReport([]byte(report))
			if err != nil {
				t.Fatal(err)
			}
			if profile != UA2ProfileName || version != "1.30.3" || compliant != tc.compliant {
				t.Fatalf("ParseReport() = (%q, %q, %v); want (%q, %q, %v)", profile, version, compliant, UA2ProfileName, "1.30.3", tc.compliant)
			}
		})
	}
}

func TestParseReportPreservesApplicationVersionForExactPinComparison(t *testing.T) {
	report := contractReport(true, `<releaseDetails id="apps" version=" 1.30.3 "/>`)
	_, version, _, err := ParseReport([]byte(report))
	if err != nil {
		t.Fatal(err)
	}
	if version != " 1.30.3 " {
		t.Fatalf("ParseReport normalized app version %q; exact pin comparison requires the observed value", version)
	}
}

func TestParseReportRequiresUnambiguousApplicationReleaseDetails(t *testing.T) {
	valid := contractReport(true, `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`)
	mutate := func(from, to string) string {
		t.Helper()
		changed := strings.Replace(valid, from, to, 1)
		if changed == valid {
			t.Fatalf("test mutation did not find %q", from)
		}
		return changed
	}
	versionless := strings.Replace(valid, `<buildInformation><releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/></buildInformation>`, "", 1)
	tests := []struct {
		name  string
		input string
	}{
		{"missing buildInformation", strings.Replace(valid, `<buildInformation>`+`<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`+`</buildInformation>`, "", 1)},
		{"duplicate buildInformation", strings.Replace(valid, `</buildInformation>`, `</buildInformation><buildInformation><releaseDetails id="apps" version="1.30.3"/></buildInformation>`, 1)},
		{"missing apps with other components present", strings.Replace(valid, ` id="apps" version="1.30.3"`, ` id="other" version="1.30.3"`, 1)},
		{"only other components", strings.Replace(valid, `<releaseDetails id="apps" version="1.30.3"/>`, "", 1)},
		{"missing id", mutate(`id="apps" `, "")},
		{"empty id", mutate(`id="apps"`, `id=""`)},
		{"whitespace id", mutate(`id="apps"`, `id="  "`)},
		{"app ID with ASCII padding", strings.Replace(valid, `id="apps"`, `id=" apps "`, 1)},
		{"app ID with Unicode padding", strings.Replace(valid, `id="apps"`, "id=\"\u00a0apps\u2003\"", 1)},
		{"missing version", mutate(` version="1.30.3"`, "")},
		{"empty version", mutate(`version="1.30.3"`, `version=""`)},
		{"whitespace version", mutate(`version="1.30.3"`, `version="  "`)},
		{"duplicate apps identical", strings.Replace(valid, `</buildInformation>`, `<releaseDetails id="apps" version="1.30.3"/></buildInformation>`, 1)},
		{"duplicate apps conflicting", strings.Replace(valid, `</buildInformation>`, `<releaseDetails id="apps" version="9.9.9"/></buildInformation>`, 1)},
		{"duplicate other ID identical", strings.Replace(valid, `</buildInformation>`, `<releaseDetails id="core" version="1.30.3"/></buildInformation>`, 1)},
		{"duplicate other ID conflicting", strings.Replace(valid, `</buildInformation>`, `<releaseDetails id="core" version="9.9.9"/></buildInformation>`, 1)},
		{"namespaced metadata", strings.Replace(valid, `<buildInformation>`, `<buildInformation xmlns="urn:shadow">`, 1)},
		{"shadowed releaseDetails", strings.Replace(valid, `<releaseDetails id="apps"`, `<x:releaseDetails xmlns:x="urn:shadow" id="apps"`, 1)},
		{"duplicate version attribute", strings.Replace(valid, `version="1.30.3"`, `version="1.30.3" version="9.9.9"`, 1)},
		{"versionless report", versionless},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
				t.Fatal("ParseReport accepted missing, malformed, or ambiguous application provenance")
			}
		})
	}
}

func TestParseReportRequiresExactlyOneValidationReportsElement(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	for _, compliant := range []bool{true, false} {
		state := "finding"
		if compliant {
			state = "pass"
		}
		t.Run(state, func(t *testing.T) {
			valid := contractReport(compliant, components)
			if _, _, gotCompliant, err := ParseReport([]byte(valid)); err != nil || gotCompliant != compliant {
				t.Fatalf("ParseReport(valid %s) = compliant %v, error %v; want compliant %v and no error", state, gotCompliant, err, compliant)
			}

			duplicate := strings.Replace(valid, `</batchSummary>`, `<validationReports compliant="0" nonCompliant="1" failedJobs="0">1</validationReports></batchSummary>`, 1)
			if duplicate == valid {
				t.Fatal("could not add a duplicate validationReports element")
			}
			if _, _, _, err := ParseReport([]byte(duplicate)); err == nil {
				t.Fatalf("ParseReport accepted %s report with multiple validationReports elements", state)
			}
		})
	}
}

func TestParseReportRequiresValidationReportsContentCountOne(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	for _, compliant := range []bool{true, false} {
		state := "finding"
		if compliant {
			state = "pass"
		}
		t.Run(state, func(t *testing.T) {
			valid := contractReport(compliant, components)
			if _, _, gotCompliant, err := ParseReport([]byte(valid)); err != nil || gotCompliant != compliant {
				t.Fatalf("ParseReport(valid %s) = compliant %v, error %v; want compliant %v and no error", state, gotCompliant, err, compliant)
			}

			for _, count := range []string{"0", "2"} {
				invalid := strings.Replace(valid, `>1</validationReports>`, `>`+count+`</validationReports>`, 1)
				if invalid == valid {
					t.Fatalf("could not set validationReports content count to %q", count)
				}
				t.Run("count_"+count, func(t *testing.T) {
					if _, _, _, err := ParseReport([]byte(invalid)); err == nil {
						t.Fatalf("ParseReport accepted %s report with validationReports content count %q", state, count)
					}
				})
			}
		})
	}
}

func TestParseReportRedactsMalformedReleaseDetails(t *testing.T) {
	marker := "VERSION_SECRET_8173"
	report := contractReport(true, `<releaseDetails id="apps" version="`+marker+`"`)
	_, _, _, err := ParseReport([]byte(report))
	if err == nil || strings.Contains(err.Error(), marker) {
		t.Fatalf("ParseReport error = %v, expected safe malformed XML rejection", err)
	}
}

type invalidBatchCounterCase struct {
	compliant bool
	counter   string
	mutation  string
	input     string
}

func TestInvalidBatchSummaryCounterCorpusContainsFortyDistinctRejectedReports(t *testing.T) {
	cases := invalidBatchSummaryCounterCases()
	if len(cases) != 40 {
		t.Fatalf("invalid batch-summary counter corpus has %d inputs, want 40", len(cases))
	}

	seen := make(map[string]struct{}, len(cases))
	for _, tc := range cases {
		if _, duplicate := seen[tc.input]; duplicate {
			t.Fatalf("duplicate negative corpus input for compliant=%v counter=%s mutation=%s", tc.compliant, tc.counter, tc.mutation)
		}
		seen[tc.input] = struct{}{}
		if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
			t.Fatalf("ParseReport accepted compliant=%v counter=%s mutation=%s", tc.compliant, tc.counter, tc.mutation)
		}
	}
}

func TestParseReportRejectsInvalidBatchSummaryCounters(t *testing.T) {
	for _, tc := range invalidBatchSummaryCounterCases() {
		state := "finding"
		if tc.compliant {
			state = "pass"
		}
		t.Run(state+"/"+tc.counter+"/"+tc.mutation, func(t *testing.T) {
			if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
				t.Fatalf("ParseReport accepted %s counter %s", tc.mutation, tc.counter)
			}
		})
	}
}

func invalidBatchSummaryCounterCases() []invalidBatchCounterCase {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	counters := []string{"totalJobs", "failedToParse", "encrypted", "outOfMemory", "veraExceptions"}
	mutations := []struct {
		name string
		bad  func(counter, expected string) (string, string)
	}{
		{name: "missing", bad: func(counter, expected string) (string, string) {
			return ` ` + counter + `="` + expected + `"`, ""
		}},
		{name: "empty", bad: func(counter, expected string) (string, string) {
			return counter + `="` + expected + `"`, counter + `=""`
		}},
		{name: "malformed", bad: func(counter, expected string) (string, string) {
			return counter + `="` + expected + `"`, counter + `="not-a-number"`
		}},
		{name: "wrong-numeric", bad: func(counter, expected string) (string, string) {
			return counter + `="` + expected + `"`, counter + `="2"`
		}},
	}
	var cases []invalidBatchCounterCase
	for _, compliant := range []bool{true, false} {
		valid := contractReport(compliant, components)
		for _, counter := range counters {
			expected := expectedBatchCounter(counter)
			for _, mutation := range mutations {
				from, to := mutation.bad(counter, expected)
				input := strings.Replace(valid, from, to, 1)
				cases = append(cases, invalidBatchCounterCase{
					compliant: compliant,
					counter:   counter,
					mutation:  mutation.name,
					input:     input,
				})
			}
		}
	}
	return cases
}

func expectedBatchCounter(name string) string {
	if name == "totalJobs" {
		return "1"
	}
	return "0"
}

func contractReport(compliant bool, releaseDetails string) string {
	compliantText, compliantCount, noncompliantCount := "false", "0", "1"
	if compliant {
		compliantText, compliantCount, noncompliantCount = "true", "1", "0"
	}
	return `<?xml version="1.0"?><report><buildInformation>` + releaseDetails + `</buildInformation><jobs><job><validationReport profileName="` + UA2ProfileName + `" isCompliant="` + compliantText + `"/></job></jobs><batchSummary totalJobs="1" failedToParse="0" encrypted="0" outOfMemory="0" veraExceptions="0"><validationReports compliant="` + compliantCount + `" nonCompliant="` + noncompliantCount + `" failedJobs="0">1</validationReports></batchSummary></report>`
}

func FuzzParseReport(f *testing.F) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	f.Add([]byte(contractReport(true, components)))
	f.Add([]byte(contractReport(false, components)))
	decoy := `<ignored><validationReport isCompliant="false"/><batchSummary><validationReports compliant="0" nonCompliant="1" failedJobs="0">1</validationReports></batchSummary></ignored>`
	f.Add([]byte(strings.Replace(contractReport(true, components), `<report>`, `<report>`+decoy, 1)))
	f.Add([]byte(strings.Replace(contractReport(true, components), `isCompliant="true"`, `isCompliant="maybe"`, 1)))
	f.Add([]byte(strings.Replace(contractReport(true, components), `compliant="1" nonCompliant="0"`, `compliant="0" nonCompliant="1"`, 1)))
	f.Add([]byte(strings.Replace(contractReport(true, components), `<report>`, `<report><ignored><batchSummary totalJobs="9" failedToParse="1" encrypted="1" outOfMemory="1" veraExceptions="1"/></ignored>`, 1)))
	for _, tc := range invalidBatchSummaryCounterCases() {
		f.Add([]byte(tc.input))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		// A panic on hostile input fails the fuzz test.
		_, _, compliant, err := ParseReport(data)
		if err != nil {
			return
		}
		// A report must never be classed compliant unless the validator's
		// own isCompliant attribute says exactly "true" AND the batch
		// summary counts independently agree with that declaration. The
		// summary re-derivation below is independent of the parser's
		// consistency gate, so it can falsify a gate bug.
		declares := reportDeclaresCompliant(data)
		if compliant != declares {
			t.Fatalf("ParseReport classed compliant=%v but the isCompliant attribute declares %v", compliant, declares)
		}
		sumC, sumNC, sumFJ := reportSummaryCounts(data)
		batchCounters, foundBatchSummary := reportBatchSummaryCounters(data)
		if !foundBatchSummary {
			t.Fatal("accepted report has no independently observed root batchSummary")
		}
		for name, want := range map[string]string{
			"totalJobs": "1", "failedToParse": "0", "encrypted": "0",
			"outOfMemory": "0", "veraExceptions": "0",
		} {
			if got := batchCounters[name]; got != want {
				t.Fatalf("accepted report has batchSummary %s=%q, want %q", name, got, want)
			}
		}
		if declares {
			if sumC != "1" || sumNC != "0" || sumFJ != "0" {
				t.Fatalf("accepted report declares compliance but batch summary counts disagree (compliant=%q nonCompliant=%q failedJobs=%q)", sumC, sumNC, sumFJ)
			}
		} else if sumC != "0" || sumNC != "1" || sumFJ != "0" {
			t.Fatalf("accepted report declares non-compliance but batch summary counts disagree (compliant=%q nonCompliant=%q failedJobs=%q)", sumC, sumNC, sumFJ)
		}
	})
}

// reportBatchSummaryCounters independently reads the five classifier-critical
// counters from the exact report/batchSummary path. It is meaningful for
// accepted input, for which ParseReport guarantees one un-namespaced summary
// without duplicate attributes.
func reportBatchSummaryCounters(data []byte) (map[string]string, bool) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var path []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, false
		}
		switch tok := token.(type) {
		case xml.StartElement:
			path = append(path, tok.Name.Local)
			if len(path) == 2 && path[0] == "report" && path[1] == "batchSummary" {
				counters := make(map[string]string, 5)
				for _, attr := range tok.Attr {
					switch attr.Name.Local {
					case "totalJobs", "failedToParse", "encrypted", "outOfMemory", "veraExceptions":
						counters[attr.Name.Local] = attr.Value
					}
				}
				return counters, true
			}
		case xml.EndElement:
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
		}
	}
}

func TestReportFuzzOraclesFollowAcceptedStructure(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	decoy := `<ignored><validationReport isCompliant="false"/><batchSummary><validationReports compliant="0" nonCompliant="1" failedJobs="0">1</validationReports></batchSummary></ignored>`
	data := []byte(strings.Replace(contractReport(true, components), `<report>`, `<report>`+decoy, 1))
	_, _, compliant, err := ParseReport(data)
	if err != nil {
		t.Fatalf("ParseReport rejected unrelated nested decoys: %v", err)
	}
	if !compliant || !reportDeclaresCompliant(data) {
		t.Fatal("fuzz oracle did not find the accepted jobs/job/validationReport value")
	}
	gotCompliant, gotNonCompliant, gotFailed := reportSummaryCounts(data)
	if gotCompliant != "1" || gotNonCompliant != "0" || gotFailed != "0" {
		t.Fatalf("fuzz oracle summary = (%q, %q, %q), want (1, 0, 0)", gotCompliant, gotNonCompliant, gotFailed)
	}
	counters, found := reportBatchSummaryCounters(data)
	if !found || counters["totalJobs"] != "1" || counters["failedToParse"] != "0" || counters["encrypted"] != "0" || counters["outOfMemory"] != "0" || counters["veraExceptions"] != "0" {
		t.Fatalf("fuzz oracle root batchSummary counters = (%v, found=%v), want (1, 0, 0, 0, 0)", counters, found)
	}
}

// reportSummaryCounts independently re-derives the summary counts from the
// exact report/batchSummary/validationReports path. It is meaningful for
// inputs ParseReport accepted, which guarantees an un-namespaced report with
// no duplicate attributes.
func reportSummaryCounts(data []byte) (compliant, nonCompliant, failedJobs string) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var path []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", "", ""
		}
		switch tok := token.(type) {
		case xml.StartElement:
			path = append(path, tok.Name.Local)
			if len(path) == 3 && path[0] == "report" && path[1] == "batchSummary" && path[2] == "validationReports" {
				for _, attr := range tok.Attr {
					switch attr.Name.Local {
					case "compliant":
						compliant = attr.Value
					case "nonCompliant":
						nonCompliant = attr.Value
					case "failedJobs":
						failedJobs = attr.Value
					}
				}
				return compliant, nonCompliant, failedJobs
			}
		case xml.EndElement:
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
		}
	}
}

// reportDeclaresCompliant independently reads the declaration at the exact
// report/jobs/job/validationReport path. It is meaningful for accepted input,
// where ParseReport guarantees one un-namespaced report with no duplicate
// attributes.
func reportDeclaresCompliant(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var path []string
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		switch tok := token.(type) {
		case xml.StartElement:
			path = append(path, tok.Name.Local)
			if len(path) == 4 && path[0] == "report" && path[1] == "jobs" && path[2] == "job" && path[3] == "validationReport" {
				for _, attr := range tok.Attr {
					if attr.Name.Local == "isCompliant" {
						return attr.Value == "true"
					}
				}
				return false
			}
		case xml.EndElement:
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
		}
	}
}

func TestParseReportRejectsStrayProcessingInstruction(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	valid := contractReport(true, components)
	cases := []struct {
		name  string
		input string
	}{
		{"after report start", strings.Replace(valid, `<report>`, `<report><?pi?>`, 1)},
		{"inside jobs", strings.Replace(valid, `<jobs>`, `<jobs><?pi?>`, 1)},
		{"trailing inside report", strings.Replace(valid, `</batchSummary></report>`, `</batchSummary><?pi?></report>`, 1)},
		{"misplaced xml declaration in body", strings.Replace(valid, `<report>`, `<report><?xml version="1.0"?>`, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
				t.Fatal("ParseReport accepted a report containing a stray processing instruction")
			}
		})
	}
}

func TestParseReportRejectsDoctypeDirective(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	valid := contractReport(true, components)
	cases := []struct {
		name  string
		input string
	}{
		{"after report start", strings.Replace(valid, `<report>`, `<report><!DOCTYPE report>`, 1)},
		{"inside batchSummary", strings.Replace(valid, `<batchSummary`, `<!DOCTYPE report><batchSummary`, 1)},
		{"trailing inside report", strings.Replace(valid, `</batchSummary></report>`, `</batchSummary><!DOCTYPE report></report>`, 1)},
		{"with internal entity subset", strings.Replace(valid, `<report>`, `<report><!DOCTYPE report [<!ENTITY x "y">]>`, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
				t.Fatal("ParseReport accepted a report containing a DOCTYPE directive")
			}
		})
	}
}

func TestParseReportRejectsTrailingBodyText(t *testing.T) {
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	valid := contractReport(true, components)
	cases := []struct {
		name  string
		input string
	}{
		{"after last element", strings.Replace(valid, `</batchSummary></report>`, `</batchSummary>trailing</report>`, 1)},
		{"after report start", strings.Replace(valid, `<report>`, `<report>trailing`, 1)},
		{"inside jobs", strings.Replace(valid, `<jobs>`, `<jobs>trailing`, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := ParseReport([]byte(tc.input)); err == nil {
				t.Fatal("ParseReport accepted a report containing stray character data")
			}
		})
	}
}

func TestParseReportAcceptsDetailsText(t *testing.T) {
	// Regression test: real veraPDF reports carry rule descriptions inside
	// validationReport/details. The stray-token rejection must not reject
	// this legitimate text (CI failure 2026-10-18: "non-whitespace data in
	// validator report" on downloaded CI bundles).
	components := `<releaseDetails id="core" version="1.30.3"/><releaseDetails id="apps" version="1.30.3"/><releaseDetails id="validation-model" version="1.30.3"/>`
	base := contractReport(true, components)
	withDetails := strings.Replace(base,
		`</validationReport>`,
		`<details><rule specification="PDF/UA-2" clause="7.1" testNumber="1" status="passed"><description>Document contains tags</description></rule></details></validationReport>`, 1)
	if _, _, _, err := ParseReport([]byte(withDetails)); err != nil {
		t.Fatalf("ParseReport rejected legitimate details text: %v", err)
	}
	// The R5-07 protections still hold: stray text outside details is rejected.
	for _, bad := range []string{
		strings.Replace(base, `<report>`, `<report>trailing`, 1),
		strings.Replace(base, `<jobs>`, `<jobs>trailing`, 1),
	} {
		if _, _, _, err := ParseReport([]byte(bad)); err == nil {
			t.Fatal("ParseReport accepted stray text outside details")
		}
	}
}
