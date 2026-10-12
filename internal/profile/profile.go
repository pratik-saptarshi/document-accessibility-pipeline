// Package profile validates proposed document-family profiles. It records
// explicit open decisions; it does not resolve policy or approve a profile.
package profile

import (
	"fmt"
	"sort"
	"strings"
)

var families = map[string]bool{"statements": true, "invoices": true, "correspondence": true, "reports": true}

type Profile struct {
	SchemaVersion   string   `json:"schema_version"`
	Status          string   `json:"status"`
	Lane            string   `json:"lane"`
	Family          string   `json:"family"`
	Language        Decision `json:"language"`
	Confidentiality Decision `json:"confidentiality"`
	Retention       Decision `json:"retention"`
	Accessibility   Decision `json:"accessibility"`
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
	if p.SchemaVersion != "1.0" {
		problems = append(problems, "schema_version must be 1.0")
	}
	if p.Status != "proposed" {
		problems = append(problems, "status must remain proposed")
	}
	if p.Lane != "generation" && p.Lane != "remediation" {
		problems = append(problems, "lane must be generation or remediation")
	}
	if !families[p.Family] {
		problems = append(problems, "family must be statements, invoices, correspondence, or reports")
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
	if len(p.SourceFields) == 0 {
		problems = append(problems, "source_fields must identify family-specific fields or explicitly unknown fields")
	}
	seen := make(map[string]bool)
	for i, field := range p.SourceFields {
		prefix := fmt.Sprintf("source_fields[%d]", i)
		if strings.TrimSpace(field.Name) == "" {
			problems = append(problems, prefix+" name is required")
		}
		if seen[field.Name] {
			problems = append(problems, prefix+" name must be unique")
		}
		seen[field.Name] = true
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
