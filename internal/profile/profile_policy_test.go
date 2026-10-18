package profile

// Policy tests for Codex findings #2, #3, #5 on PR #145: the lane JSON
// schemas must match the Go validator on document families (finding #2),
// Validate must refuse the deferred restricted confidentiality class
// (finding #3), and both must enforce the English-only launch contract
// (finding #5). Decisions 0014 (accepted families) and 0016 (deferred
// restricted class and Spanish) are the sources of truth.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

// TestLaneSchemaFamilyEnumsMatchValidatorAcceptedSet pins the Decision 0014
// accepted family set and requires both lane JSON schemas to enumerate
// exactly the families Validate accepts, so the Go and JSON validators
// cannot diverge (Codex finding #2 on PR #145).
func TestLaneSchemaFamilyEnumsMatchValidatorAcceptedSet(t *testing.T) {
	want := []string{"correspondence", "forms", "hr_documents", "invoices", "marketing", "presentations", "reports", "statements"}
	goFamilies := make([]string, 0, len(families))
	for family := range families {
		goFamilies = append(goFamilies, family)
	}
	sort.Strings(goFamilies)
	if !slices.Equal(goFamilies, want) {
		t.Fatalf("Validate accepts families %v, want Decision 0014 accepted set %v", goFamilies, want)
	}
	for _, lane := range []string{"generation", "remediation"} {
		t.Run(lane, func(t *testing.T) {
			path := filepath.Join("..", "..", "docs", "workflows", lane, "profiles", "profile.schema.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			}
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatal(err)
			}
			got := append([]string(nil), schema.Properties["family"].Enum...)
			sort.Strings(got)
			if !slices.Equal(got, want) {
				t.Fatalf("%s schema family enum = %v, want Decision 0014 accepted set %v", lane, got, want)
			}
		})
	}
}

// validPolicyProfile returns a profile that satisfies every Validate rule:
// accepted eight-family member, English language, non-restricted
// confidentiality, and the Decision 0014 structural fields for forms.
func validPolicyProfile() Profile {
	return Profile{
		SchemaVersion: "1.1", Status: "proposed", Lane: "generation", Family: "forms",
		Language:        Decision{Value: "en", Status: "proposed"},
		Confidentiality: Decision{Value: "internal", Status: "proposed"},
		Retention:       Decision{Value: "transient", Status: "proposed"},
		Accessibility:   Decision{Value: "semantic expectations open", Status: "open"},
		PDFUATarget:     Decision{Value: "PDF/UA-2", Status: "accepted"},
		HumanReview:     Decision{Value: "review required; protocol open", Status: "open"},
		SourceFields: []Field{
			{Name: "form_fields", Type: "string", Status: "proposed"},
			{Name: "instructions", Type: "string", Status: "proposed"},
			{Name: "sections", Type: "string", Status: "proposed"},
		},
	}
}

// TestValidateRejectsRestrictedConfidentiality requires Validate to refuse
// the deferred restricted confidentiality class with a named error citing
// Decision 0016, whether the decision status is proposed or open
// (Codex finding #3 on PR #145).
func TestValidateRejectsRestrictedConfidentiality(t *testing.T) {
	for _, status := range []string{"proposed", "open"} {
		t.Run("status="+status, func(t *testing.T) {
			p := validPolicyProfile()
			p.Confidentiality = Decision{Value: "restricted", Status: status}
			problems := strings.Join(Validate(p), "\n")
			if !strings.Contains(problems, `"restricted"`) || !strings.Contains(problems, "Decision 0016") {
				t.Fatalf("restricted confidentiality (status %q) must be refused citing Decision 0016, got: %q", status, problems)
			}
		})
	}
	if problems := Validate(validPolicyProfile()); len(problems) != 0 {
		t.Fatalf("non-restricted profile must validate: %v", problems)
	}
}

// TestValidateRejectsNonEnglishLanguage_AllowlistEnEnUSEnGBPerDecision0016
// enforces the English-only launch contract. Decision 0016 defers Spanish
// without enumerating tags, so the allowlist is exactly the BCP 47 tags
// en, en-US, and en-GB; es and es-US are rejected (Codex finding #5 on
// PR #145).
func TestValidateRejectsNonEnglishLanguage_AllowlistEnEnUSEnGBPerDecision0016(t *testing.T) {
	for _, lang := range []string{"es", "es-US", "fr"} {
		t.Run("reject/"+lang, func(t *testing.T) {
			p := validPolicyProfile()
			p.Language = Decision{Value: lang, Status: "proposed"}
			problems := strings.Join(Validate(p), "\n")
			if !strings.Contains(problems, "Decision 0016") {
				t.Fatalf("language %q must be rejected citing Decision 0016, got: %q", lang, problems)
			}
		})
	}
	for _, lang := range []string{"en", "en-US", "en-GB"} {
		t.Run("accept/"+lang, func(t *testing.T) {
			p := validPolicyProfile()
			p.Language = Decision{Value: lang, Status: "proposed"}
			if problems := Validate(p); len(problems) != 0 {
				t.Fatalf("language %q must validate: %v", lang, problems)
			}
		})
	}
}

// TestLaneSchemaLanguageEnumsMatchEnglishAllowlist requires both lane JSON
// schemas to constrain the language decision value to the same English-only
// allowlist Validate enforces, so the Go and JSON validators cannot diverge
// (Codex finding #5 on PR #145).
func TestLaneSchemaLanguageEnumsMatchEnglishAllowlist(t *testing.T) {
	want := []string{"en", "en-GB", "en-US"}
	goLangs := make([]string, 0, len(englishLanguages))
	for lang := range englishLanguages {
		goLangs = append(goLangs, lang)
	}
	sort.Strings(goLangs)
	if !slices.Equal(goLangs, want) {
		t.Fatalf("Validate language allowlist = %v, want %v", goLangs, want)
	}
	for _, lane := range []string{"generation", "remediation"} {
		t.Run(lane, func(t *testing.T) {
			path := filepath.Join("..", "..", "docs", "workflows", lane, "profiles", "profile.schema.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Defs struct {
					LanguageDecision struct {
						Properties struct {
							Value struct {
								Enum []string `json:"enum"`
							} `json:"value"`
						} `json:"properties"`
					} `json:"languageDecision"`
				} `json:"$defs"`
			}
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatal(err)
			}
			got := append([]string(nil), schema.Defs.LanguageDecision.Properties.Value.Enum...)
			sort.Strings(got)
			if !slices.Equal(got, want) {
				t.Fatalf("%s schema language enum = %v, want English allowlist %v", lane, got, want)
			}
		})
	}
}

// TestValidateRejectsRestrictedConfidentialityWhitespaceAndCaseVariants
// requires the restricted-class refusal to normalize the input before
// comparing: boundary whitespace or case variants must not bypass the
// hard out-of-scope policy (Codex P2 on PR #154).
func TestValidateRejectsRestrictedConfidentialityWhitespaceAndCaseVariants(t *testing.T) {
	for _, value := range []string{"restricted ", " restricted", "\trestricted\n", "RESTRICTED", "Restricted", "ReStRiCtEd"} {
		t.Run("value="+value, func(t *testing.T) {
			p := validPolicyProfile()
			p.Confidentiality = Decision{Value: value, Status: "proposed"}
			problems := strings.Join(Validate(p), "\n")
			if !strings.Contains(problems, "Decision 0016") {
				t.Fatalf("confidentiality %q must be refused citing Decision 0016, got: %q", value, problems)
			}
		})
	}
}
