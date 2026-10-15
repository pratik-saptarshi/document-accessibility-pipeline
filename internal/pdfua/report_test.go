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

func TestParseReportRedactsMalformedReleaseDetails(t *testing.T) {
	marker := "VERSION_SECRET_8173"
	report := contractReport(true, `<releaseDetails id="apps" version="`+marker+`"`)
	_, _, _, err := ParseReport([]byte(report))
	if err == nil || strings.Contains(err.Error(), marker) {
		t.Fatalf("ParseReport error = %v, expected safe malformed XML rejection", err)
	}
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
		if declares {
			if sumC != "1" || sumNC != "0" || sumFJ != "0" {
				t.Fatalf("accepted report declares compliance but batch summary counts disagree (compliant=%q nonCompliant=%q failedJobs=%q)", sumC, sumNC, sumFJ)
			}
		} else if sumC != "0" || sumNC != "1" || sumFJ != "0" {
			t.Fatalf("accepted report declares non-compliance but batch summary counts disagree (compliant=%q nonCompliant=%q failedJobs=%q)", sumC, sumNC, sumFJ)
		}
	})
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
