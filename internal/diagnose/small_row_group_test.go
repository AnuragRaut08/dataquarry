package diagnose

import (
	"strings"
	"testing"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

func TestSmallRowGroupFindings(t *testing.T) {
	tests := []struct {
		name         string
		uncompressed int64
		rowGroups    int
		wantFindings int
		wantEvidence string
	}{
		{
			name:         "undersized row groups",
			uncompressed: 40 * 96 * 1024,
			rowGroups:    40,
			wantFindings: 1,
			wantEvidence: "Average row group size is 96.0 KiB across 40 row groups.",
		},
		{
			name:         "healthy row groups",
			uncompressed: 128 * 1024 * 1024,
			rowGroups:    1,
			wantFindings: 0,
		},
		{
			name:         "exact threshold",
			uncompressed: SmallRowGroupThresholdBytes,
			rowGroups:    1,
			wantFindings: 0,
		},
		{
			name:         "below threshold",
			uncompressed: SmallRowGroupThresholdBytes - 1,
			rowGroups:    1,
			wantFindings: 1,
		},
		{
			name:         "zero row groups",
			uncompressed: 1024,
			rowGroups:    0,
			wantFindings: 0,
		},
		{
			name:         "zero uncompressed size",
			uncompressed: 0,
			rowGroups:    2,
			wantFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := summary.Summary{
				NumRowGroups:     tt.rowGroups,
				UncompressedSize: tt.uncompressed,
			}

			findings := SmallRowGroupFindings(dataset)

			if len(findings) != tt.wantFindings {
				t.Fatalf("got %d findings, want %d",
					len(findings), tt.wantFindings)
			}

			if tt.wantFindings == 0 {
				return
			}

			finding := findings[0]

			if finding.Severity != SeverityHigh {
				t.Errorf("severity = %q, want %q",
					finding.Severity, SeverityHigh)
			}

			if !strings.Contains(finding.Evidence, tt.wantEvidence) {
				t.Errorf("evidence = %q, want it to contain %q",
					finding.Evidence, tt.wantEvidence)
			}

			if finding.Title == "" || finding.Fix == "" {
				t.Error("expected finding title and fix to be populated")
			}
		})
	}
}
