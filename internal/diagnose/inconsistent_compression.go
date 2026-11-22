package diagnose

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

// InconsistentColumnCompressionFindings detects columns that use
// different compression codecs across their row groups.
func InconsistentColumnCompressionFindings(dataset summary.Summary) []Finding {
	var findings []Finding

	for _, column := range dataset.Columns {
		codecs := column.CompressionCodecs

		// A column with zero or one codec is consistent.
		if len(codecs) <= 1 {
			continue
		}

		// Sort codecs for deterministic output.
		sortedCodecs := append([]string(nil), codecs...)
		sort.Strings(sortedCodecs)

		findings = append(findings, NewFinding(
			"Inconsistent column compression",
			SeverityMedium,
			fmt.Sprintf(
				"Column %q uses multiple compression codecs: %s.",
				column.Name,
				strings.Join(sortedCodecs, ", "),
			),
			fmt.Sprintf(
				"Configure column %q to use a consistent compression codec across all row groups.",
				column.Name,
			),
		))

	}

	return findings
}
