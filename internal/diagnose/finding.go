package diagnose

// Severity represents the importance of a diagnostic finding.
type Severity string

const (
	// SeverityLow indicates a low-impact finding.
	SeverityLow Severity = "low"

	// SeverityMedium indicates a moderate finding.
	SeverityMedium Severity = "medium"

	// SeverityHigh indicates a significant finding.
	SeverityHigh Severity = "high"

	// SeverityCritical indicates a critical finding.
	SeverityCritical Severity = "critical"
)

// Finding is a structured, machine-readable diagnostic result.
type Finding struct {
	// Title is the short summary of the finding.
	Title string

	// Severity indicates the importance of the finding.
	Severity Severity

	// Evidence contains the measured evidence that triggered the finding.
	Evidence string

	// Fix contains a suggested remediation.
	Fix string
}

// NewFinding creates a structured diagnostic finding.
func NewFinding(title string, severity Severity, evidence string, fix string) Finding {
	return Finding{
		Title:    title,
		Severity: severity,
		Evidence: evidence,
		Fix:      fix,
	}
}
