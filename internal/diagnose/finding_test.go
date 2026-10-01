package diagnose

import "testing"

func TestNewFinding(t *testing.T) {
	f := NewFinding(
		"Row groups are much smaller than recommended",
		SeverityHigh,
		"Average row group size is 96.1 KiB across 40 row groups.",
		"Rewrite the file using a larger target row group size.",
	)

	if f.Title != "Row groups are much smaller than recommended" {
		t.Fatal("unexpected title")
	}

	if f.Severity != SeverityHigh {
		t.Fatal("unexpected severity")
	}

	if f.Evidence == "" {
		t.Fatal("evidence should not be empty")
	}

	if f.Fix == "" {
		t.Fatal("fix should not be empty")
	}
}

func TestSeverityValues(t *testing.T) {
	values := []Severity{
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}

	if len(values) != 4 {
		t.Fatal("unexpected severity count")
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.want {
			t.Errorf("%q.String() = %q, want %q", tt.severity, got, tt.want)
		}
	}
}
