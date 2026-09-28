package diagnose

// Severity represents the importance of a diagnostic finding.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Finding is a structured, machine-readable diagnostic result.
type Finding struct {
	Title    string
	Severity Severity
	Evidence string
	Fix      string
}

// NewFinding creates a Finding with consistent field ordering.
func NewFinding(title string, severity Severity, evidence string, fix string) Finding {
	return Finding{
		Title:    title,
		Severity: severity,
		Evidence: evidence,
		Fix:      fix,
	}
}
