package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
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
