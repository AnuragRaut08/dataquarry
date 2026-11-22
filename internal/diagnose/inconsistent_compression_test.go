package diagnose

import (
	"strings"
	"testing"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

func TestInconsistentColumnCompressionFindings(t *testing.T) {
	tests := []struct {
		name           string
		columns        []summary.ColumnSummary
		wantFindings   int
		wantColumnName string
	}{
		{
			name: "consistent compression",
			columns: []summary.ColumnSummary{
				{
					Name:              "name",
					CompressionCodecs: []string{"SNAPPY"},
				},
			},
			wantFindings: 0,
		},
		{
			name: "inconsistent compression",
			columns: []summary.ColumnSummary{
				{
					Name:              "salary",
					CompressionCodecs: []string{"SNAPPY", "ZSTD"},
				},
			},
			wantFindings:   1,
			wantColumnName: "salary",
		},
		{
			name: "missing codec metadata",
			columns: []summary.ColumnSummary{
				{
					Name:              "department",
					CompressionCodecs: []string{},
				},
			},
			wantFindings: 0,
		},
		{
			name: "only inconsistent column triggers finding",
			columns: []summary.ColumnSummary{
				{
					Name:              "id",
					CompressionCodecs: []string{"SNAPPY"},
				},
				{
					Name:              "salary",
					CompressionCodecs: []string{"SNAPPY", "GZIP"},
				},
			},
			wantFindings:   1,
			wantColumnName: "salary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := summary.Summary{
				Columns: tt.columns,
			}

			findings := InconsistentColumnCompressionFindings(dataset)

			if len(findings) != tt.wantFindings {
				t.Fatalf(
					"got %d findings, want %d",
					len(findings),
					tt.wantFindings,
				)
			}

			if tt.wantFindings > 0 {
				if tt.name == "inconsistent compression" {
					if !strings.Contains(findings[0].Evidence, "SNAPPY") ||
						!strings.Contains(findings[0].Evidence, "ZSTD") {
						t.Errorf(
							"evidence %q should contain both SNAPPY and ZSTD",
							findings[0].Evidence,
						)
					}
				}

				if tt.wantColumnName != "" &&
					!strings.Contains(findings[0].Evidence, tt.wantColumnName) {
					t.Errorf(
						"evidence %q does not contain column name %q",
						findings[0].Evidence,
						tt.wantColumnName,
					)
				}
			}
		})
	}
}
