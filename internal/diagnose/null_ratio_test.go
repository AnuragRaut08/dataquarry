package diagnose

import (
	"testing"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

func TestHighNullRatioFindings(t *testing.T) {
	tests := []struct {
		name         string
		numRows      int64
		column       summary.ColumnSummary
		wantFindings int
		wantSeverity Severity
	}{
		{
			name:    "below threshold",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 49, NullCountKnown: true,
			},
			wantFindings: 0,
		},
		{
			name:    "at medium threshold",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 50, NullCountKnown: true,
			},
			wantFindings: 1, wantSeverity: SeverityMedium,
		},
		{
			name:    "at high threshold",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 80, NullCountKnown: true,
			},
			wantFindings: 1, wantSeverity: SeverityHigh,
		},
		{
			name:    "unknown null count ignored",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 90, NullCountKnown: false,
			},
			wantFindings: 0,
		},
		{
			name:    "negative null count ignored",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: -1, NullCountKnown: true,
			},
			wantFindings: 0,
		},
		{
			name:    "null count above row count ignored",
			numRows: 100,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 101, NullCountKnown: true,
			},
			wantFindings: 0,
		},
		{
			name:    "zero rows ignored",
			numRows: 0,
			column: summary.ColumnSummary{
				Name: "email", NullCount: 0, NullCountKnown: true,
			},
			wantFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := HighNullRatioFindings(tt.numRows, []summary.ColumnSummary{tt.column})

			if len(findings) != tt.wantFindings {
				t.Fatalf("got %d findings, want %d", len(findings), tt.wantFindings)
			}

			if tt.wantFindings == 0 {
				return
			}

			got := findings[0]
			if got.Severity != tt.wantSeverity {
				t.Errorf("severity = %q, want %q", got.Severity, tt.wantSeverity)
			}
			if got.Evidence == "" {
				t.Error("expected evidence to be populated")
			}
			if got.Fix == "" {
				t.Error("expected recommendation to be populated")
			}
		})
	}
}
