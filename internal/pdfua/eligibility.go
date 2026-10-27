package pdfua

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Finding identifies one failed check in a validator report. A finding is
// identified by the failed rule (specification, clause, test number, and
// description) and the exact failed check (object, test expression,
// context, and error message), so two reports can be compared check by
// check. Text fields are whitespace-normalized: the same validator may
// wrap the same text differently between runs, and wrapping is not a
// different finding.
type Finding struct {
	Specification string
	Clause        string
	TestNumber    string
	Description   string
	Object        string
	Test          string
	Context       string
	ErrorMessage  string
}

func normalizeFindingText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// Classification partitions candidate findings against damaged-input
// findings. Inherited findings occur in both reports, Introduced findings
// occur only in the candidate report, and Resolved findings occur only in
// the damaged-input report. Comparison is a multiset comparison: repeated
// identical checks count individually.
type Classification struct {
	Inherited  []Finding
	Introduced []Finding
	Resolved   []Finding
}

// ClassifyFindings compares the failed checks of a damaged-input report
// with those of a candidate report produced from that input.
func ClassifyFindings(damaged, candidate []Finding) Classification {
	remaining := make(map[Finding]int, len(damaged))
	for _, finding := range damaged {
		remaining[finding]++
	}
	var classification Classification
	for _, finding := range candidate {
		if remaining[finding] > 0 {
			remaining[finding]--
			classification.Inherited = append(classification.Inherited, finding)
			continue
		}
		classification.Introduced = append(classification.Introduced, finding)
	}
	seen := make(map[Finding]bool, len(damaged))
	for _, finding := range damaged {
		if seen[finding] {
			continue
		}
		seen[finding] = true
		for ; remaining[finding] > 0; remaining[finding]-- {
			classification.Resolved = append(classification.Resolved, finding)
		}
	}
	return classification
}

// Eligibility is the Decision 0026 assessed-evidence outcome for one
// candidate: eligible when every candidate finding is inherited from the
// damaged input (zero introduced findings). Compliance is not required
// and is not decided here.
type Eligibility struct {
	Eligible       bool
	Classification Classification
}

// AssessEligibility parses a damaged-input report and a candidate report
// from the pinned validator and classifies every candidate finding. It
// fails closed when either report is malformed, is not an explicit
// PDF/UA-2 report in the pinned validator shape, or does not enumerate
// its failed checks completely.
func AssessEligibility(damagedReport, candidateReport []byte) (Eligibility, error) {
	damagedFindings, err := ParseFailedChecks(damagedReport)
	if err != nil {
		return Eligibility{}, fmt.Errorf("damaged-input report: %w", err)
	}
	candidateFindings, err := ParseFailedChecks(candidateReport)
	if err != nil {
		return Eligibility{}, fmt.Errorf("candidate report: %w", err)
	}
	classification := ClassifyFindings(damagedFindings, candidateFindings)
	return Eligibility{
		Eligible:       len(classification.Introduced) == 0,
		Classification: classification,
	}, nil
}

type reportDetails struct {
	XMLName xml.Name `xml:"report"`
	Jobs    struct {
		Job []struct {
			ValidationReports []struct {
				Compliant string `xml:"isCompliant,attr"`
				Details   *struct {
					PassedRules  string `xml:"passedRules,attr"`
					FailedRules  string `xml:"failedRules,attr"`
					PassedChecks string `xml:"passedChecks,attr"`
					FailedChecks string `xml:"failedChecks,attr"`
					Rules        []struct {
						Specification string `xml:"specification,attr"`
						Clause        string `xml:"clause,attr"`
						TestNumber    string `xml:"testNumber,attr"`
						Status        string `xml:"status,attr"`
						FailedChecks  string `xml:"failedChecks,attr"`
						Description   string `xml:"description"`
						Object        string `xml:"object"`
						Test          string `xml:"test"`
						Checks        []struct {
							Status       string `xml:"status,attr"`
							Context      string `xml:"context"`
							ErrorMessage string `xml:"errorMessage"`
						} `xml:"check"`
					} `xml:"rule"`
				} `xml:"details"`
			} `xml:"validationReport"`
		} `xml:"job"`
	} `xml:"jobs"`
}

// validateDetailsStructure token-walks the report and enforces the exact
// element grammar of the validationReport details subtree: only the
// elements the pinned validator emits there, singleton description,
// object, and test elements per rule, singleton context and errorMessage
// elements per check, and no unknown attributes on details, rule, or
// check elements. ParseReport's token validation has already rejected
// namespaces, duplicate attributes, and stray text; this pass rejects the
// duplicate or unknown structural elements a plain struct decode would
// silently merge or drop.
func validateDetailsStructure(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var stack []string
	detailsSeen := 0
	var ruleSingletonSeen map[string]int
	var checkSingletonSeen map[string]int
	inRule := false
	inCheck := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("malformed validator XML report")
		}
		switch token := token.(type) {
		case xml.StartElement:
			path := append(append([]string{}, stack...), token.Name.Local)
			joined := strings.Join(path, "/")
			switch joined {
			case "report/jobs/job/validationReport/details":
				detailsSeen++
				if detailsSeen > 1 {
					return fmt.Errorf("validator report contains duplicate details elements")
				}
				if err := validateAttrs(token, "passedRules", "failedRules", "passedChecks", "failedChecks"); err != nil {
					return err
				}
			case "report/jobs/job/validationReport/details/rule":
				if inRule {
					return fmt.Errorf("validator report contains nested rule elements")
				}
				inRule = true
				ruleSingletonSeen = map[string]int{}
				if err := validateAttrs(token, "specification", "clause", "testNumber", "status", "failedChecks", "tags"); err != nil {
					return err
				}
			case "report/jobs/job/validationReport/details/rule/description",
				"report/jobs/job/validationReport/details/rule/object",
				"report/jobs/job/validationReport/details/rule/test":
				if !inRule {
					return fmt.Errorf("validator report rule child appears outside a rule")
				}
				ruleSingletonSeen[token.Name.Local]++
				if ruleSingletonSeen[token.Name.Local] > 1 {
					return fmt.Errorf("validator report rule contains duplicate %s elements", token.Name.Local)
				}
			case "report/jobs/job/validationReport/details/rule/check":
				if !inRule || inCheck {
					return fmt.Errorf("validator report check appears outside a rule or nested in a check")
				}
				inCheck = true
				checkSingletonSeen = map[string]int{}
				if err := validateAttrs(token, "status"); err != nil {
					return err
				}
			case "report/jobs/job/validationReport/details/rule/check/context",
				"report/jobs/job/validationReport/details/rule/check/errorMessage":
				if !inCheck {
					return fmt.Errorf("validator report check child appears outside a check")
				}
				checkSingletonSeen[token.Name.Local]++
				if checkSingletonSeen[token.Name.Local] > 1 {
					return fmt.Errorf("validator report check contains duplicate %s elements", token.Name.Local)
				}
			default:
				if strings.HasPrefix(joined, "report/jobs/job/validationReport/details/") {
					return fmt.Errorf("validator report details contain an unsupported element %q", token.Name.Local)
				}
			}
			stack = append(stack, token.Name.Local)
		case xml.EndElement:
			joined := strings.Join(stack, "/")
			switch joined {
			case "report/jobs/job/validationReport/details/rule":
				inRule = false
				for _, required := range []string{"description", "object", "test"} {
					if ruleSingletonSeen[required] != 1 {
						return fmt.Errorf("validator report rule does not contain exactly one %s element", required)
					}
				}
			case "report/jobs/job/validationReport/details/rule/check":
				inCheck = false
				for _, required := range []string{"context", "errorMessage"} {
					if checkSingletonSeen[required] != 1 {
						return fmt.Errorf("validator report check does not contain exactly one %s element", required)
					}
				}
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return nil
}

func validateAttrs(element xml.StartElement, allowed ...string) error {
	seen := make(map[string]bool, len(element.Attr))
	for _, attr := range element.Attr {
		if attr.Name.Space != "" {
			return fmt.Errorf("validator report %s element contains a namespaced attribute", element.Name.Local)
		}
		ok := false
		for _, name := range allowed {
			if attr.Name.Local == name {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("validator report %s element contains an unsupported attribute %q", element.Name.Local, attr.Name.Local)
		}
		seen[attr.Name.Local] = true
	}
	for _, name := range allowed {
		// tags is optional on rule elements; every other listed attribute is required.
		if name == "tags" {
			continue
		}
		if !seen[name] {
			return fmt.Errorf("validator report %s element is missing the %s attribute", element.Name.Local, name)
		}
	}
	return nil
}

// ParseFailedChecks returns every failed check enumerated in a validator
// report, in report order. The report is first validated by ParseReport,
// so only the exact single-job explicit PDF/UA-2 shape is accepted, and
// the details subtree is then structure-checked token by token. In the
// pinned validator's reports the details element enumerates failed rules
// only; a noncompliant report must carry it, and its declared failed rule
// and check counts (and each rule's declared failed check count) must
// match the enumerated failed rules and checks exactly. A compliant
// report must declare zero failed rules and checks. Anything else means
// the findings cannot be enumerated completely and the report is
// rejected.
func ParseFailedChecks(data []byte) ([]Finding, error) {
	_, _, compliant, err := ParseReport(data)
	if err != nil {
		return nil, err
	}
	if err := validateDetailsStructure(data); err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root reportDetails
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("malformed validator XML report")
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("malformed validator XML report")
		}
		switch token := token.(type) {
		case xml.CharData:
			if strings.TrimSpace(string(token)) != "" {
				return nil, fmt.Errorf("unexpected data after validator report")
			}
		case xml.Comment:
			continue
		default:
			return nil, fmt.Errorf("unexpected token after validator report")
		}
	}
	if len(root.Jobs.Job) != 1 || len(root.Jobs.Job[0].ValidationReports) != 1 {
		return nil, fmt.Errorf("report must contain exactly one job and one validationReport")
	}
	validationReport := root.Jobs.Job[0].ValidationReports[0]
	if validationReport.Details == nil {
		if compliant {
			return nil, nil
		}
		return nil, fmt.Errorf("noncompliant validator report does not enumerate its failed checks")
	}
	details := validationReport.Details
	counts := make(map[string]int, 4)
	for name, raw := range map[string]string{
		"passedRules": details.PassedRules, "failedRules": details.FailedRules,
		"passedChecks": details.PassedChecks, "failedChecks": details.FailedChecks,
	} {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("validator report details contain an invalid %s count", name)
		}
		counts[name] = value
	}
	if compliant && (counts["failedRules"] != 0 || counts["failedChecks"] != 0) {
		return nil, fmt.Errorf("compliant validator report declares failed rules or checks in its details")
	}
	if !compliant && (counts["failedRules"] == 0 || counts["failedChecks"] == 0) {
		return nil, fmt.Errorf("noncompliant validator report declares zero failed rules or checks in its details")
	}
	var findings []Finding
	for _, rule := range details.Rules {
		if rule.Status != "failed" {
			return nil, fmt.Errorf("validator report details enumerate a rule with status %q; only failed rules are enumerated", rule.Status)
		}
		ruleFailedChecks, err := strconv.Atoi(rule.FailedChecks)
		if err != nil || ruleFailedChecks <= 0 {
			return nil, fmt.Errorf("validator rule contains an invalid failedChecks count")
		}
		if len(rule.Checks) != ruleFailedChecks {
			return nil, fmt.Errorf("validator rule failedChecks count does not match its enumerated failed checks")
		}
		for _, check := range rule.Checks {
			if check.Status != "failed" {
				return nil, fmt.Errorf("failed validator rule enumerates a non-failed check")
			}
			findings = append(findings, Finding{
				Specification: normalizeFindingText(rule.Specification),
				Clause:        normalizeFindingText(rule.Clause),
				TestNumber:    normalizeFindingText(rule.TestNumber),
				Description:   normalizeFindingText(rule.Description),
				Object:        normalizeFindingText(rule.Object),
				Test:          normalizeFindingText(rule.Test),
				Context:       normalizeFindingText(check.Context),
				ErrorMessage:  normalizeFindingText(check.ErrorMessage),
			})
		}
	}
	if len(details.Rules) != counts["failedRules"] {
		return nil, fmt.Errorf("validator report failedRules count does not match its enumerated failed rules")
	}
	if len(findings) != counts["failedChecks"] {
		return nil, fmt.Errorf("validator report failedChecks count does not match its enumerated failed checks")
	}
	return findings, nil
}
