package pdfua

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	validationBundleSchema = 1
	maxBundleMetadata      = 16 << 10
	maxBundleReport        = 8 << 20
	expectedCLIHash        = "19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70"
	expectedBaseImage      = "eclipse-temurin:21.0.12.1_1-jre-noble@sha256:000fd431958bc81a24abe1e8e5f0f0fd3ae365a594bd50aadb20696805f9408c"
	passControlHash        = "c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b"
	findingControlHash     = "2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0"
)

// ValidationBundle is the versioned, non-secret handoff from the isolated
// veraPDF container to the Go-only report verifier. It contains hashes and
// process metadata, never the validator executable or the PDF itself.
type ValidationBundle struct {
	SchemaVersion    int    `json:"schema_version"`
	Validator        string `json:"validator"`
	ValidatorVersion string `json:"validator_version"`
	ValidatorSHA256  string `json:"validator_sha256"`
	BaseImage        string `json:"base_image"`
	ContainerImageID string `json:"container_image_id"`
	Control          string `json:"control"`
	Profile          string `json:"profile"`
	ExitCode         int    `json:"exit_code"`
	CandidateSHA256  string `json:"candidate_sha256"`
	ReportSHA256     string `json:"report_sha256"`
}

// VerifyBundle checks the contract metadata and raw report independently of
// Java or veraPDF. The outcome is always derived from the XML; metadata cannot
// assert pass/finding on its own.
func VerifyBundle(metadataJSON, reportXML []byte) (Result, error) {
	if len(metadataJSON) == 0 || len(metadataJSON) > maxBundleMetadata {
		return Result{}, fmt.Errorf("validation bundle metadata size is invalid")
	}
	if len(reportXML) == 0 || len(reportXML) > maxBundleReport {
		return Result{}, fmt.Errorf("validation bundle report size is invalid")
	}
	if err := validateBundleKeys(metadataJSON); err != nil {
		return Result{}, err
	}
	var bundle ValidationBundle
	decoder := json.NewDecoder(bytes.NewReader(metadataJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return Result{}, fmt.Errorf("invalid validation bundle metadata")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Result{}, fmt.Errorf("validation bundle must contain exactly one JSON object")
	}
	if bundle.SchemaVersion != validationBundleSchema {
		return Result{}, fmt.Errorf("unsupported validation bundle schema")
	}
	if bundle.Validator != "veraPDF" || bundle.ValidatorVersion != "1.30.3" || bundle.ValidatorSHA256 != expectedCLIHash || bundle.BaseImage != expectedBaseImage || !validImageID(bundle.ContainerImageID) {
		return Result{}, fmt.Errorf("validation bundle validator identity mismatch")
	}
	if bundle.Profile != UA2ProfileName {
		return Result{}, fmt.Errorf("validation bundle profile mismatch")
	}
	if !validSHA256(bundle.CandidateSHA256) || !validSHA256(bundle.ReportSHA256) {
		return Result{}, fmt.Errorf("validation bundle contains an invalid SHA-256")
	}
	if strings.TrimSpace(bundle.Control) == "" {
		return Result{}, fmt.Errorf("validation bundle control is missing")
	}
	reportHash := sha256.Sum256(reportXML)
	actualReportHash := hex.EncodeToString(reportHash[:])
	if actualReportHash != bundle.ReportSHA256 {
		return Result{}, fmt.Errorf("validation bundle report digest mismatch")
	}
	profile, validatorVersion, compliant, err := ParseReport(reportXML)
	if err != nil {
		return Result{}, fmt.Errorf("invalid validator XML report: %w", err)
	}
	if validatorVersion != bundle.ValidatorVersion {
		return Result{}, fmt.Errorf("validator report application version mismatch")
	}
	if profile != bundle.Profile || profile != UA2ProfileName {
		return Result{}, fmt.Errorf("validator report profile mismatch")
	}
	result := Result{
		ValidatorVersion: validatorVersion,
		Profile:          profile,
		ExitCode:         bundle.ExitCode,
		ArtifactSHA256:   bundle.ValidatorSHA256,
		CandidateSHA256:  bundle.CandidateSHA256,
		ReportSHA256:     actualReportHash,
	}
	switch bundle.ExitCode {
	case 0:
		if !compliant {
			return Result{}, fmt.Errorf("validator exit/report mismatch")
		}
		result.Outcome = OutcomePass
	case 1:
		if compliant {
			return Result{}, fmt.Errorf("validator exit/report mismatch")
		}
		result.Outcome = OutcomeFinding
	default:
		return Result{}, fmt.Errorf("validator returned processing exit code %d", bundle.ExitCode)
	}
	return result, nil
}

// VerifyControlBundle verifies a named corpus control against its fixed
// candidate digest and expected pass/finding result. VerifyBundle remains the
// generic verifier for other explicitly identified PDF candidates.
func VerifyControlBundle(metadataJSON, reportXML []byte, expectedControl string) (Result, error) {
	result, err := VerifyBundle(metadataJSON, reportXML)
	if err != nil {
		return Result{}, err
	}
	var bundle ValidationBundle
	decoder := json.NewDecoder(bytes.NewReader(metadataJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return Result{}, fmt.Errorf("invalid validation bundle metadata")
	}
	expectedHash := ""
	expectedOutcome := Outcome("")
	switch expectedControl {
	case "known-pass":
		expectedHash, expectedOutcome = passControlHash, OutcomePass
	case "known-finding":
		expectedHash, expectedOutcome = findingControlHash, OutcomeFinding
	default:
		return Result{}, fmt.Errorf("expected control is unsupported")
	}
	if bundle.Control != expectedControl || bundle.CandidateSHA256 != expectedHash || result.Outcome != expectedOutcome {
		return Result{}, fmt.Errorf("validation bundle does not match the expected control, digest, and outcome")
	}
	return result, nil
}

func validSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == string(bytes.ToLower([]byte(value)))
}

func validImageID(value string) bool {
	return len(value) == len("sha256:")+sha256.Size*2 && len(value) > 7 && value[:7] == "sha256:" && validSHA256(value[7:])
}

func validateBundleKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("validation bundle metadata must be a JSON object")
	}
	allowed := map[string]struct{}{
		"schema_version": {}, "validator": {}, "validator_version": {}, "validator_sha256": {},
		"base_image": {}, "container_image_id": {}, "control": {}, "profile": {},
		"exit_code": {}, "candidate_sha256": {}, "report_sha256": {},
	}
	seen := make(map[string]struct{}, len(allowed))
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return fmt.Errorf("invalid validation bundle metadata")
		}
		key, ok := token.(string)
		if !ok {
			return fmt.Errorf("validation bundle metadata contains an invalid key")
		}
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("validation bundle metadata contains an unknown or noncanonical key")
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("validation bundle metadata contains a duplicate key")
		}
		seen[key] = struct{}{}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return fmt.Errorf("invalid validation bundle metadata")
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("validation bundle metadata contains a null field")
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return fmt.Errorf("invalid validation bundle metadata")
	}
	if len(seen) != len(allowed) {
		return fmt.Errorf("validation bundle metadata is missing a required field")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("validation bundle must contain exactly one JSON object")
	}
	return nil
}
