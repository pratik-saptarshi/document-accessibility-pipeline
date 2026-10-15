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
	f.Add([]byte(strings.Replace(contractReport(true, components), `isCompliant="true"`, `isCompliant="maybe"`, 1)))
	f.Fuzz(func(t *testing.T, data []byte) {
		// A panic on hostile input fails the fuzz test.
		_, _, compliant, err := ParseReport(data)
		if err != nil {
			return
		}
		// A report must never be classed compliant unless the validator's
		// own isCompliant attribute says exactly "true". The parser's
		// consistency gate then requires the batch summary to agree, so
		// inconsistent counts can never yield compliant=true either.
		if compliant && !reportDeclaresCompliant(data) {
			t.Fatalf("ParseReport classed a report compliant whose isCompliant attribute is not \"true\"")
		}
	})
}

// reportDeclaresCompliant re-derives the validator's compliance declaration
// independently of ParseReport. It is only meaningful for inputs ParseReport
// accepted, which guarantees exactly one un-namespaced validationReport with
// no duplicate attributes.
func reportDeclaresCompliant(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "validationReport" {
			continue
		}
		for _, attr := range start.Attr {
			if attr.Name.Local == "isCompliant" {
				return attr.Value == "true"
			}
		}
		return false
	}
}
