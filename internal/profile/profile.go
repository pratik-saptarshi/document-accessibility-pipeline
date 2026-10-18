// Package profile validates document-family profiles. It records the accepted
// PDF/UA target separately from policy choices that remain open.
package profile

import (
	"fmt"
	"sort"
	"strings"
)

var families = map[string]bool{"statements": true, "invoices": true, "correspondence": true, "reports": true, "presentations": true, "forms": true, "marketing": true, "hr_documents": true}

// englishLanguages is the Decision 0016 launch allowlist: the pipeline
// launches English-only (Decision 0014 deferred Spanish; Decision 0016
// records the deferral). Only these BCP 47 tags are accepted.
var englishLanguages = map[string]bool{"en": true, "en-US": true, "en-GB": true}

// familyRequiredFields holds the structural source fields Decision 0014
// (docs/decisions/0014-profile-policy-decisions.md, section 1) requires each
// family to carry. Only presence is checked: per-subfield rules from the
// decision (form fields' label/field_type/required, transcript/caption for
// embedded media, image alt-text and decorative marking, descriptive link
// text and target, required alt-text on report figures) have no slot in the
// profile Field schema and remain human review per the decision's
// machine/human boundary.
var familyRequiredFields = map[string][]string{
	"statements":     {"period_start", "period_end", "account_holder", "transactions"},
	"invoices":       {"invoice_number", "invoice_date", "bill_to", "line_items"},
	"correspondence": {"title", "date", "sender", "recipient", "body", "signature_block", "language"},
	"reports":        {"title", "report_date", "author", "sections", "tables", "figures", "footnotes"},
	"presentations":  {"slide_count", "slide_titles", "speaker_notes_included", "embedded_media", "charts_per_slide"},
	"forms":          {"form_fields", "instructions", "sections"},
	"marketing":      {"headings", "images", "links", "language"},
	"hr_documents":   {"title", "effective_date", "sections", "tables", "signature_fields"},
}

type Profile struct {
	SchemaVersion   string   `json:"schema_version"`
	Status          string   `json:"status"`
	Lane            string   `json:"lane"`
	Family          string   `json:"family"`
	Language        Decision `json:"language"`
	Confidentiality Decision `json:"confidentiality"`
	Retention       Decision `json:"retention"`
	Accessibility   Decision `json:"accessibility"`
	PDFUATarget     Decision `json:"pdfua_target"`
	HumanReview     Decision `json:"human_review"`
	SourceFields    []Field  `json:"source_fields"`
}

type Decision struct {
	Value  string `json:"value"`
	Status string `json:"status"`
}

type Field struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// Validate requires the profile's unresolved policies to be visible and
// rejects ambiguous omissions. Values remain descriptive proposals.
func Validate(p Profile) []string {
	var problems []string
	if p.SchemaVersion != "1.1" {
		problems = append(problems, "schema_version must be 1.1")
	}
	if p.Status != "proposed" {
		problems = append(problems, "status must remain proposed")
	}
	if p.Lane != "generation" && p.Lane != "remediation" {
		problems = append(problems, "lane must be generation or remediation")
	}
	if !families[p.Family] {
		problems = append(problems, "family must be statements, invoices, correspondence, reports, presentations, forms, marketing, or hr_documents")
	}
	for name, decision := range map[string]Decision{
		"language": p.Language, "confidentiality": p.Confidentiality,
		"retention": p.Retention, "accessibility": p.Accessibility, "human_review": p.HumanReview,
	} {
		if strings.TrimSpace(decision.Value) == "" {
			problems = append(problems, name+" decision value must be explicit")
		}
		if decision.Status != "proposed" && decision.Status != "open" {
			problems = append(problems, name+" decision status must be proposed or open")
		}
	}
	// Decision 0016 defers the restricted confidentiality class: the Q5
	// threat-data controls are proposed but not implemented or independently
	// reviewed, so the pipeline refuses restricted-class documents.
	if p.Confidentiality.Value == "restricted" {
		problems = append(problems, `confidentiality "restricted" is refused per Decision 0016: restricted-class documents are out of scope until Q5 threat-data controls are implemented and independently reviewed`)
	}
	// Decision 0016 keeps the launch contract English-only: Spanish is
	// deferred, so only the englishLanguages allowlist tags are accepted.
	// Blank values are already reported as non-explicit above.
	if v := strings.TrimSpace(p.Language.Value); v != "" && !englishLanguages[v] {
		problems = append(problems, fmt.Sprintf("language %q is not supported per Decision 0016: pipeline launches English-only (en, en-US, en-GB)", p.Language.Value))
	}
	if strings.TrimSpace(p.PDFUATarget.Value) == "" {
		problems = append(problems, "pdfua_target decision value must be explicit")
	} else if p.PDFUATarget.Status == "accepted" {
		if p.PDFUATarget.Value != "PDF/UA-2" {
			problems = append(problems, "accepted pdfua_target must be PDF/UA-2")
		}
	} else if p.PDFUATarget.Status != "proposed" && p.PDFUATarget.Status != "open" {
		problems = append(problems, "pdfua_target status must be accepted, proposed, or open")
	}
	if len(p.SourceFields) == 0 {
		problems = append(problems, "source_fields must identify family-specific fields or explicitly unknown fields")
	}
	if required, ok := familyRequiredFields[p.Family]; ok {
		present := make(map[string]bool, len(p.SourceFields))
		for _, field := range p.SourceFields {
			present[field.Name] = true
		}
		for _, name := range required {
			if !present[name] {
				problems = append(problems, fmt.Sprintf("family %s requires source field %q", p.Family, name))
			}
		}
	}
	seen := make(map[string]bool)
	for i, field := range p.SourceFields {
		prefix := fmt.Sprintf("source_fields[%d]", i)
		name := strings.TrimSpace(field.Name)
		if name == "" {
			problems = append(problems, prefix+" name is required")
		}
		if field.Name != name {
			problems = append(problems, prefix+" name must not have leading or trailing whitespace")
		}
		if seen[name] {
			problems = append(problems, prefix+" name must be unique")
		}
		seen[name] = true
		if field.Type != "string" && field.Type != "number" && field.Type != "date" && field.Type != "boolean" && field.Type != "unknown" {
			problems = append(problems, prefix+" type is unsupported")
		}
		if field.Status != "proposed" && field.Status != "open" {
			problems = append(problems, prefix+" status must be proposed or open")
		}
	}
	sort.Strings(problems)
	return problems
}
