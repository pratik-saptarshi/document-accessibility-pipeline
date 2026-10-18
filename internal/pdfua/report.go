package pdfua

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const UA2ProfileName = "PDF/UA-2 + Tagged PDF validation profile"

// ParseReport accepts only the exact single-job XML shape emitted by the
// pinned validator. It is independent of Java and process execution.
func ParseReport(data []byte) (string, string, bool, error) {
	if err := validateReportTokens(data); err != nil {
		return "", "", false, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root struct {
		XMLName          xml.Name `xml:"report"`
		BuildInformation []struct {
			ReleaseDetails []struct {
				ID      string `xml:"id,attr"`
				Version string `xml:"version,attr"`
			} `xml:"releaseDetails"`
		} `xml:"buildInformation"`
		Jobs struct {
			Job []struct {
				ValidationReports []struct {
					ProfileName string `xml:"profileName,attr"`
					Compliant   string `xml:"isCompliant,attr"`
				} `xml:"validationReport"`
			} `xml:"job"`
		} `xml:"jobs"`
		BatchSummaries []struct {
			TotalJobs         string `xml:"totalJobs,attr"`
			FailedToParse     string `xml:"failedToParse,attr"`
			Encrypted         string `xml:"encrypted,attr"`
			OutOfMemory       string `xml:"outOfMemory,attr"`
			VeraExceptions    string `xml:"veraExceptions,attr"`
			ValidationReports []struct {
				Compliant    string `xml:"compliant,attr"`
				NonCompliant string `xml:"nonCompliant,attr"`
				FailedJobs   string `xml:"failedJobs,attr"`
				Count        string `xml:",chardata"`
			} `xml:"validationReports"`
		} `xml:"batchSummary"`
	}
	var rootStart xml.StartElement
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", "", false, fmt.Errorf("malformed validator XML report")
		}
		switch token := token.(type) {
		case xml.CharData:
			if strings.TrimSpace(string(token)) != "" {
				return "", "", false, fmt.Errorf("non-whitespace data before report root")
			}
		case xml.Comment:
			continue
		case xml.ProcInst:
			if token.Target != "xml" {
				return "", "", false, fmt.Errorf("unexpected processing instruction before report root")
			}
		case xml.StartElement:
			rootStart = token
		default:
			return "", "", false, fmt.Errorf("unexpected token before report root")
		}
		if rootStart.Name.Local != "" {
			break
		}
	}
	if rootStart.Name.Local != "report" {
		return "", "", false, fmt.Errorf("report root element is missing or invalid")
	}
	if rootStart.Name.Space != "" {
		return "", "", false, fmt.Errorf("namespaced validator reports are not supported")
	}
	if err := decoder.DecodeElement(&root, &rootStart); err != nil {
		return "", "", false, fmt.Errorf("malformed validator XML report")
	}
	if root.XMLName.Local != "report" || len(root.BuildInformation) != 1 || len(root.Jobs.Job) != 1 || len(root.Jobs.Job[0].ValidationReports) != 1 || len(root.BatchSummaries) != 1 {
		return "", "", false, fmt.Errorf("report must contain exactly one buildInformation, one job, one validationReport, and one batchSummary")
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", false, fmt.Errorf("malformed validator XML report")
		}
		switch token := token.(type) {
		case xml.CharData:
			if strings.TrimSpace(string(token)) != "" {
				return "", "", false, fmt.Errorf("non-whitespace data after validation report")
			}
		case xml.Comment:
			continue
		default:
			return "", "", false, fmt.Errorf("unexpected token after validation report")
		}
	}
	componentIDs := make(map[string]struct{}, len(root.BuildInformation[0].ReleaseDetails))
	applicationVersion := ""
	for _, component := range root.BuildInformation[0].ReleaseDetails {
		id := component.ID
		trimmedID := strings.TrimSpace(id)
		version := component.Version
		if trimmedID == "" || id != trimmedID || strings.TrimSpace(version) == "" {
			return "", "", false, fmt.Errorf("validator buildInformation contains an invalid release component ID or version")
		}
		if _, exists := componentIDs[id]; exists {
			return "", "", false, fmt.Errorf("validator buildInformation contains duplicate release component IDs")
		}
		componentIDs[id] = struct{}{}
		if id == "apps" {
			applicationVersion = version
		}
	}
	if applicationVersion == "" {
		return "", "", false, fmt.Errorf("validator application version is missing")
	}
	report := root.Jobs.Job[0].ValidationReports[0]
	if strings.TrimSpace(report.ProfileName) == "" {
		return "", "", false, fmt.Errorf("validationReport profileName is missing")
	}
	var compliant bool
	switch report.Compliant {
	case "true":
		compliant = true
	case "false":
		compliant = false
	default:
		return "", "", false, fmt.Errorf("validationReport isCompliant must be true or false")
	}
	summary := root.BatchSummaries[0]
	if summary.TotalJobs != "1" || summary.FailedToParse != "0" || summary.Encrypted != "0" || summary.OutOfMemory != "0" || summary.VeraExceptions != "0" || len(summary.ValidationReports) != 1 {
		return "", "", false, fmt.Errorf("batchSummary reports a processing failure or unexpected job count")
	}
	validationSummary := summary.ValidationReports[0]
	wantCompliantText, wantNonCompliantText := "0", "1"
	if compliant {
		wantCompliantText, wantNonCompliantText = "1", "0"
	}
	if validationSummary.Compliant != wantCompliantText || validationSummary.NonCompliant != wantNonCompliantText || validationSummary.FailedJobs != "0" || strings.TrimSpace(validationSummary.Count) != "1" {
		return "", "", false, fmt.Errorf("batchSummary and validationReport disagree")
	}
	return report.ProfileName, applicationVersion, compliant, nil
}

// validateReportTokens rejects XML features the classifier does not need,
// rejects stray structural text, and avoids encoding/xml's duplicate-attribute
// last-value-wins behavior.
func validateReportTokens(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	inProlog := true
	prologContentSeen := false
	declarationSeen := false
	var stack []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("malformed validator XML report")
		}
		if proc, ok := token.(xml.ProcInst); ok {
			if proc.Target != "xml" || !inProlog || prologContentSeen || declarationSeen {
				return fmt.Errorf("unexpected processing instruction in validator report")
			}
			declarationSeen = true
			continue
		}
		if _, ok := token.(xml.Directive); ok {
			return fmt.Errorf("unexpected directive in validator report")
		}
		if charData, ok := token.(xml.CharData); ok {
			if inProlog && len(charData) > 0 {
				prologContentSeen = true
			}
			if strings.TrimSpace(string(charData)) != "" {
				path := strings.Join(stack, "/")
				if !isValidatorTextElement(path) {
					return fmt.Errorf("non-whitespace data in validator report")
				}
			}
			continue
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			if _, ok := token.(xml.Comment); ok && inProlog {
				prologContentSeen = true
			}
			if _, ok := token.(xml.EndElement); ok && len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		inProlog = false
		stack = append(stack, start.Name.Local)
		if start.Name.Space != "" {
			return fmt.Errorf("namespaced validator reports are not supported")
		}
		seen := make(map[xml.Name]struct{}, len(start.Attr))
		for _, attr := range start.Attr {
			if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" || attr.Name.Space != "" {
				return fmt.Errorf("namespaced validator reports are not supported")
			}
			if _, exists := seen[attr.Name]; exists {
				return fmt.Errorf("malformed validator XML report")
			}
			seen[attr.Name] = struct{}{}
		}
	}
}

func isValidatorTextElement(path string) bool {
	switch path {
	case "report/jobs/job/item/name",
		"report/jobs/job/duration",
		"report/jobs/job/validationReport/details/rule/description",
		"report/jobs/job/validationReport/details/rule/object",
		"report/jobs/job/validationReport/details/rule/test",
		"report/jobs/job/validationReport/details/rule/check/context",
		"report/jobs/job/validationReport/details/rule/check/errorMessage",
		"report/batchSummary/validationReports",
		"report/batchSummary/featureReports",
		"report/batchSummary/repairReports",
		"report/batchSummary/duration":
		return true
	default:
		return false
	}
}
