package corpus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

const schemaVersion = "1.0.0"

var families = []string{"statements", "invoices", "correspondence", "reports"}

var requiredDimensions = []string{
	"visible_text",
	"page_count",
	"page_geometry",
	"reading_order_when_known",
	"links",
	"forms",
	"annotations",
	"language",
	"metadata",
}

var requiredRequirements = []string{"REQ-REM-001", "REQ-REM-002", "REQ-REM-003"}

var requiredBeadsIssues = []string{
	"dap-jo9.1.1",
	"dap-jo9.1.2",
	"dap-jo9.2.1",
	"dap-jo9.2.1.1",
}

// Manifest describes planned synthetic characterization cases, not PDF artifacts or results.
type Manifest struct {
	SchemaVersion  string   `json:"schema_version"`
	CorpusID       string   `json:"corpus_id"`
	Lane           string   `json:"lane"`
	Status         string   `json:"status"`
	Provenance     string   `json:"provenance"`
	RequirementIDs []string `json:"requirement_ids"`
	BeadsIssueIDs  []string `json:"beads_issue_ids"`
	Cases          []Case   `json:"cases"`
}

// Case records a planned fixture and the characterization dimensions to consider.
type Case struct {
	ID             string   `json:"id"`
	Family         string   `json:"family"`
	ProfileExample string   `json:"profile_example"`
	FixtureStatus  string   `json:"fixture_status"`
	Synthetic      bool     `json:"synthetic"`
	RunStatus      string   `json:"run_status"`
	Dimensions     []string `json:"dimensions"`
}

// Load parses a manifest with unknown fields rejected so fixture paths cannot be smuggled in.
func Load(filename string) (Manifest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode corpus manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Manifest{}, errors.New("decode corpus manifest: multiple JSON values")
		}
		return Manifest{}, fmt.Errorf("decode corpus manifest trailing data: %w", err)
	}
	return manifest, nil
}

// Validate checks planning metadata only. It does not inspect PDFs or establish expected behavior.
func Validate(manifest Manifest) []string {
	var problems []string
	if manifest.SchemaVersion != schemaVersion {
		problems = append(problems, "schema_version must be "+schemaVersion)
	}
	if strings.TrimSpace(manifest.CorpusID) == "" {
		problems = append(problems, "corpus_id is required")
	}
	if manifest.Lane != "remediation" {
		problems = append(problems, "lane must be remediation")
	}
	if manifest.Status != "proposed" {
		problems = append(problems, "status must be proposed")
	}
	if manifest.Provenance != "synthetic-only" {
		problems = append(problems, "provenance must be synthetic-only")
	}
	checkRequiredIDs(&problems, "requirement_ids", manifest.RequirementIDs, requiredRequirements)
	checkRequiredIDs(&problems, "beads_issue_ids", manifest.BeadsIssueIDs, requiredBeadsIssues)
	if len(manifest.Cases) != len(families) {
		problems = append(problems, "cases must contain exactly one planned case for each document family")
	}

	seenIDs := make(map[string]bool)
	seenFamilies := make(map[string]bool)
	for index, c := range manifest.Cases {
		prefix := fmt.Sprintf("cases[%d]", index)
		if strings.TrimSpace(c.ID) == "" {
			problems = append(problems, prefix+" id is required")
		} else if seenIDs[c.ID] {
			problems = append(problems, prefix+" case ID must be unique")
		}
		seenIDs[c.ID] = true

		if !contains(families, c.Family) {
			problems = append(problems, prefix+" family is unsupported")
		} else if seenFamilies[c.Family] {
			problems = append(problems, prefix+" must be the only case for its family in v1")
		}
		seenFamilies[c.Family] = true

		wantExample := path.Join("docs", "profiles", "examples", "remediation", c.Family+".json")
		if !safeProfileExample(c.ProfileExample) || c.ProfileExample != wantExample {
			problems = append(problems, prefix+" profile_example must reference the matching synthetic remediation profile")
		}
		if c.FixtureStatus != "planned-not-created" {
			problems = append(problems, prefix+" fixture_status must be planned-not-created")
		}
		if !c.Synthetic {
			problems = append(problems, prefix+" must be marked synthetic")
		}
		if c.RunStatus != "not-run" {
			problems = append(problems, prefix+" run_status must be not-run")
		}
		seenDimensions := make(map[string]bool, len(c.Dimensions))
		for _, dimension := range c.Dimensions {
			if !contains(requiredDimensions, dimension) {
				problems = append(problems, prefix+" dimensions contain unsupported value "+dimension)
			}
			if seenDimensions[dimension] {
				problems = append(problems, prefix+" dimensions must not contain duplicates")
			}
			seenDimensions[dimension] = true
		}
		for _, dimension := range requiredDimensions {
			if !contains(c.Dimensions, dimension) {
				problems = append(problems, prefix+" dimensions must include "+dimension)
			}
		}
	}
	for _, family := range families {
		if !seenFamilies[family] {
			problems = append(problems, "cases must cover family "+family)
		}
	}
	sort.Strings(problems)
	return problems
}

func checkRequiredIDs(problems *[]string, name string, actual, required []string) {
	seen := make(map[string]bool, len(actual))
	for _, id := range actual {
		if seen[id] {
			*problems = append(*problems, name+" must not contain duplicates")
		}
		seen[id] = true
	}
	for _, id := range required {
		if !seen[id] {
			*problems = append(*problems, name+" must include "+id)
		}
	}
}

func safeProfileExample(value string) bool {
	return value != "" && !path.IsAbs(value) && path.Clean(value) == value && !strings.Contains(value, `\`) &&
		strings.HasPrefix(value, "docs/profiles/examples/remediation/") && strings.HasSuffix(value, ".json")
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
