package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProposedExamplesCoverBothLanesAndAllFamilies(t *testing.T) {
	root := filepath.Join("..", "..", "docs", "profiles", "examples")
	lanes := []string{"generation", "remediation"}
	families := []string{"statements", "invoices", "correspondence", "reports"}
	for _, lane := range lanes {
		for _, family := range families {
			t.Run(lane+"/"+family, func(t *testing.T) {
				path := filepath.Join(root, lane, family+".json")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var p Profile
				if err := json.Unmarshal(data, &p); err != nil {
					t.Fatal(err)
				}
				if problems := Validate(p); len(problems) != 0 {
					t.Fatalf("invalid proposed example: %v", problems)
				}
				if p.Status != "proposed" || p.Lane != lane || p.Family != family {
					t.Fatalf("example lost proposed scope: %#v", p)
				}
			})
		}
	}
}

func TestRejectsInvalidOrAmbiguousProfiles(t *testing.T) {
	base := Profile{
		SchemaVersion: "1.1", Status: "proposed", Lane: "remediation", Family: "statements",
		Language:        Decision{Value: "en", Status: "proposed"},
		Confidentiality: Decision{Value: "unclassified", Status: "open"},
		Retention:       Decision{Value: "open", Status: "open"},
		Accessibility:   Decision{Value: "semantic expectations open", Status: "open"},
		PDFUATarget:     Decision{Value: "PDF/UA-2", Status: "accepted"},
		HumanReview:     Decision{Value: "review required; protocol open", Status: "open"},
		SourceFields:    []Field{{Name: "account_reference", Type: "string", Status: "proposed"}},
	}
	cases := []struct {
		name   string
		mutate func(*Profile)
	}{
		{"missing family", func(p *Profile) { p.Family = "" }},
		{"unknown family", func(p *Profile) { p.Family = "claims" }},
		{"unsupported schema version", func(p *Profile) { p.SchemaVersion = "2.0" }},
		{"invalid lane", func(p *Profile) { p.Lane = "shared" }},
		{"unmarked decision", func(p *Profile) { p.Retention.Status = "approved" }},
		{"ambiguous blank language", func(p *Profile) { p.Language.Value = " " }},
		{"missing family source fields", func(p *Profile) { p.SourceFields = nil }},
		{"invalid field type", func(p *Profile) { p.SourceFields[0].Type = "blob" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			p.SourceFields = append([]Field(nil), base.SourceFields...)
			tc.mutate(&p)
			if len(Validate(p)) == 0 {
				t.Fatal("expected profile validation errors")
			}
		})
	}
}

func TestSourceFieldNamesMustBeCanonicalAndUnique(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "profiles", "examples", "remediation", "statements.json"))
	if err != nil {
		t.Fatal(err)
	}
	var base Profile
	if err := json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		mutate          func(*Profile)
		wantProblemPart string
	}{
		{
			name:            "padded name",
			wantProblemPart: "name must not have leading or trailing whitespace",
			mutate: func(p *Profile) {
				p.SourceFields[0].Name = " " + p.SourceFields[0].Name
			},
		},
		{
			name:            "duplicate canonical name",
			wantProblemPart: "name must be unique",
			mutate: func(p *Profile) {
				p.SourceFields = append(p.SourceFields, p.SourceFields[0])
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			p.SourceFields = append([]Field(nil), base.SourceFields...)
			tc.mutate(&p)
			problems := strings.Join(Validate(p), "\n")
			if !strings.Contains(problems, tc.wantProblemPart) {
				t.Fatalf("expected validation problem containing %q, got %q", tc.wantProblemPart, problems)
			}
		})
	}
}

func TestSourceFieldWhitespaceBoundariesFollowTrimSpace(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "profiles", "examples", "remediation", "statements.json"))
	if err != nil {
		t.Fatal(err)
	}
	var base Profile
	if err := json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}
	name := base.SourceFields[0].Name
	tests := []struct {
		name      string
		fieldName string
		valid     bool
	}{
		{name: "leading U+0085", fieldName: "\u0085" + name},
		{name: "trailing U+0085", fieldName: name + "\u0085"},
		{name: "leading NBSP", fieldName: "\u00a0" + name},
		{name: "trailing NBSP", fieldName: name + "\u00a0"},
		{name: "leading em space", fieldName: "\u2003" + name},
		{name: "trailing ideographic space", fieldName: name + "\u3000"},
		{name: "trailing line feed", fieldName: name + "\n"},
		{name: "trailing carriage return", fieldName: name + "\r"},
		{name: "trailing CRLF", fieldName: name + "\r\n"},
		{name: "leading FEFF is not trimmed by Go", fieldName: "\ufeff" + name, valid: true},
		{name: "trailing FEFF is not trimmed by Go", fieldName: name + "\ufeff", valid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			p.SourceFields = append([]Field(nil), base.SourceFields...)
			p.SourceFields[0].Name = tc.fieldName
			problems := strings.Join(Validate(p), "\n")
			gotValid := !strings.Contains(problems, "name must not have leading or trailing whitespace")
			if gotValid != tc.valid {
				t.Fatalf("field name %q valid = %t, want %t; problems: %s", tc.fieldName, gotValid, tc.valid, problems)
			}
		})
	}
}

func TestLaneSchemasUseTrimSpaceBoundarySet(t *testing.T) {
	const wantPattern = `^(?:[\u0009-\u000D\u0020\u0085\u00A0\u1680\u2000-\u200A\u2028\u2029\u202F\u205F\u3000]|[\s\S]*[\u0009-\u000D\u0020\u0085\u00A0\u1680\u2000-\u200A\u2028\u2029\u202F\u205F\u3000](?![\s\S]))`
	for _, lane := range []string{"generation", "remediation"} {
		t.Run(lane, func(t *testing.T) {
			path := filepath.Join("..", "..", "docs", "workflows", lane, "profiles", "profile.schema.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Defs struct {
					Field struct {
						Properties struct {
							Name struct {
								Not struct {
									Pattern string `json:"pattern"`
								} `json:"not"`
							} `json:"name"`
						} `json:"properties"`
					} `json:"field"`
				} `json:"$defs"`
			}
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatal(err)
			}
			pattern := schema.Defs.Field.Properties.Name.Not.Pattern
			if pattern != wantPattern {
				t.Fatalf("source field boundary pattern = %q, want %q", pattern, wantPattern)
			}
		})
	}
}

func TestLaneSchemasRejectWhitespaceOnlyDecisionValues(t *testing.T) {
	whitespace := []string{" ", "\t", "\n", "\v", "\f", "\r", "\u0085", "\u00a0", "\u1680", "\u2000", "\u200a", "\u2028", "\u2029", "\u202f", "\u205f", "\u3000"}
	for _, lane := range []string{"generation", "remediation"} {
		t.Run(lane, func(t *testing.T) {
			path := filepath.Join("..", "..", "docs", "workflows", lane, "profiles", "profile.schema.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Defs struct {
					Decision struct {
						Properties struct {
							Value struct {
								Pattern string `json:"pattern"`
							} `json:"value"`
						} `json:"properties"`
					} `json:"decision"`
				} `json:"$defs"`
			}
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatal(err)
			}
			if schema.Defs.Decision.Properties.Value.Pattern == "" {
				t.Fatal("decision.value schema is missing its non-whitespace pattern")
			}
			pattern, err := regexp.Compile(schema.Defs.Decision.Properties.Value.Pattern)
			if err != nil {
				t.Fatalf("compile decision.value schema pattern: %v", err)
			}
			for _, value := range whitespace {
				if pattern.MatchString(value) {
					t.Errorf("schema pattern accepts whitespace-only decision value %q", value)
				}
				profile := Profile{
					SchemaVersion: "1.1", Status: "proposed", Lane: lane, Family: "statements",
					Language:        Decision{Value: value, Status: "proposed"},
					Confidentiality: Decision{Value: "open", Status: "open"},
					Retention:       Decision{Value: "open", Status: "open"},
					Accessibility:   Decision{Value: "review remains required", Status: "open"},
					PDFUATarget:     Decision{Value: "PDF/UA-2", Status: "accepted"},
					HumanReview:     Decision{Value: "protocol open", Status: "open"},
					SourceFields:    []Field{{Name: "unknown", Type: "unknown", Status: "open"}},
				}
				if problems := Validate(profile); len(problems) == 0 {
					t.Errorf("Go validation accepts whitespace-only decision value %q", value)
				}
			}
			if !pattern.MatchString("open") {
				t.Fatal("schema pattern must accept an explicit non-whitespace value")
			}
		})
	}
}

func TestAcceptedPDFUATargetDoesNotApproveOtherPolicies(t *testing.T) {
	base := Profile{
		SchemaVersion: "1.1", Status: "proposed", Lane: "remediation", Family: "statements",
		Language:        Decision{Value: "en", Status: "proposed"},
		Confidentiality: Decision{Value: "open", Status: "open"},
		Retention:       Decision{Value: "open", Status: "open"},
		Accessibility:   Decision{Value: "semantic and assistive-technology review remain required", Status: "open"},
		HumanReview:     Decision{Value: "review protocol and authority open", Status: "open"},
		PDFUATarget:     Decision{Value: "PDF/UA-2", Status: "accepted"},
		SourceFields:    []Field{{Name: "unknown", Type: "unknown", Status: "open"}},
	}
	if problems := Validate(base); len(problems) != 0 {
		t.Fatalf("accepted UA-2 target should be independently representable: %v", problems)
	}
	base.Retention.Status = "accepted"
	if problems := Validate(base); len(problems) == 0 {
		t.Fatal("accepting the PDF/UA target must not accept retention policy")
	}
	base.Retention.Status = "open"
	base.PDFUATarget.Value = "PDF/UA-1"
	if problems := Validate(base); len(problems) == 0 {
		t.Fatal("accepted PDF target must be exactly PDF/UA-2")
	}
}
