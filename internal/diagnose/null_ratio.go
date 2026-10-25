

package diagnose

import (
	"fmt"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

const HighNullRatioThreshold = 0.50

// HighNullRatioFindings reports columns whose known null ratio
// meets or exceeds the configured threshold.
func HighNullRatioFindings(
	numRows int64,
	columns []summary.ColumnSummary,
) []Finding {
	if numRows <= 0 {
		return nil
	}

	var findings []Finding

	for _, column := range columns {
		if !column.NullCountKnown {
			continue
		}
		if column.NullCount < 0 || column.NullCount > numRows {
			continue
		}

		ratio := float64(column.NullCount) / float64(numRows)
		if ratio < HighNullRatioThreshold {
			continue
		}

		severity := SeverityMedium
		if ratio >= 0.80 {
			severity = SeverityHigh
		}

		if err := severity.Validate(); err != nil {
			continue
		}

		findings = append(findings, NewFinding(
			fmt.Sprintf("Column %q has a high null ratio", column.Name),
			severity,
			fmt.Sprintf(
				"Column %q has %d nulls out of %d rows (%.1f%%).",
				column.Name,
				column.NullCount,
				numRows,
				ratio*100,
			),
			"Check whether the column is necessary and investigate why values are missing.",
		))
	}

	return findings
}
