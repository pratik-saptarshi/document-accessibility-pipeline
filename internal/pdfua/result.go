package pdfua

type Outcome string

const (
	OutcomePass    Outcome = "pass"
	OutcomeFinding Outcome = "finding"
)

type Result struct {
	ValidatorVersion string
	Profile          string
	Outcome          Outcome
	ExitCode         int
	ArtifactSHA256   string
	CandidateSHA256  string
	ReportSHA256     string
	// BundleMetadataSHA256 identifies the exact verified metadata bytes.
	BundleMetadataSHA256 string
}
