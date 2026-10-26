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
// identified by the failed rule (specification, clause, and test number)
// and the exact failed check (object, test expression, context, and error
// message), so two reports can be compared check by check.
type Finding struct {
	Specification string
	Clause        string
	TestNumber    string
	Object        string
	Test          string
	Context       string
	ErrorMessage  string
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
// fails closed when either report is malformed, is not an explicit PDF/UA-2
// report from the pinned validator shape, or does not enumerate its failed
// checks completely.
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
					FailedRules  string `xml:"failedRules,attr"`
					FailedChecks string `xml:"failedChecks,attr"`
					Rules        []struct {
						Specification string `xml:"specification,attr"`
						Clause        string `xml:"clause,attr"`
						TestNumber    string `xml:"testNumber,attr"`
						Status        string `xml:"status,attr"`
						FailedChecks  string `xml:"failedChecks,attr"`
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

// ParseFailedChecks returns every failed check enumerated in a validator
// report, in report order. The report is first validated by ParseReport,
// so only the exact single-job explicit PDF/UA-2 shape is accepted. A
// noncompliant report must carry a details element whose declared failed
// rule and check counts match the enumerated failed rules and checks;
// otherwise the findings cannot be enumerated completely and the report
// is rejected.
func ParseFailedChecks(data []byte) ([]Finding, error) {
	_, _, compliant, err := ParseReport(data)
	if err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root reportDetails
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("malformed validator XML report")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("malformed validator XML report")
	}
	if len(root.Jobs.Job) != 1 || len(root.Jobs.Job[0].ValidationReports) != 1 {
		return nil, fmt.Errorf("report must contain exactly one job and one validationReport")
	}
	validationReport := root.Jobs.Job[0].ValidationReports[0]
	if compliant {
		if validationReport.Details == nil {
			return nil, nil
		}
	}
	if validationReport.Details == nil {
		return nil, fmt.Errorf("noncompliant validator report does not enumerate its failed checks")
	}
	details := validationReport.Details
	declaredFailedChecks, err := strconv.Atoi(details.FailedChecks)
	if err != nil || declaredFailedChecks < 0 {
		return nil, fmt.Errorf("validator report details contain an invalid failedChecks count")
	}
	declaredFailedRules, err := strconv.Atoi(details.FailedRules)
	if err != nil || declaredFailedRules < 0 {
		return nil, fmt.Errorf("validator report details contain an invalid failedRules count")
	}
	var findings []Finding
	failedRules := 0
	for _, rule := range details.Rules {
		switch rule.Status {
		case "passed":
			if len(rule.Checks) != 0 {
				return nil, fmt.Errorf("passed validator rule enumerates checks")
			}
			continue
		case "failed":
			failedRules++
		default:
			return nil, fmt.Errorf("validator rule has an unsupported status %q", rule.Status)
		}
		ruleFailedChecks, err := strconv.Atoi(rule.FailedChecks)
		if err != nil || ruleFailedChecks < 0 {
			return nil, fmt.Errorf("validator rule contains an invalid failedChecks count")
		}
		failedInRule := 0
		for _, check := range rule.Checks {
			if check.Status != "failed" {
				return nil, fmt.Errorf("failed validator rule enumerates a non-failed check")
			}
			failedInRule++
			findings = append(findings, Finding{
				Specification: strings.TrimSpace(rule.Specification),
				Clause:        strings.TrimSpace(rule.Clause),
				TestNumber:    strings.TrimSpace(rule.TestNumber),
				Object:        strings.TrimSpace(rule.Object),
				Test:          strings.TrimSpace(rule.Test),
				Context:       strings.TrimSpace(check.Context),
				ErrorMessage:  strings.TrimSpace(check.ErrorMessage),
			})
		}
		if failedInRule != ruleFailedChecks {
			return nil, fmt.Errorf("validator rule failedChecks count does not match its enumerated failed checks")
		}
	}
	if failedRules != declaredFailedRules {
		return nil, fmt.Errorf("validator report failedRules count does not match its enumerated failed rules")
	}
	if len(findings) != declaredFailedChecks {
		return nil, fmt.Errorf("validator report failedChecks count does not match its enumerated failed checks")
	}
	if compliant && len(findings) != 0 {
		return nil, fmt.Errorf("compliant validator report enumerates failed checks")
	}
	return findings, nil
}
