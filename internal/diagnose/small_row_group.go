package diagnose

import (
	"fmt"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
)

const (
	// ReferenceRowGroupSizeBytes is the recommended reference target for
	// analytical Parquet workloads: 128 MiB.
	ReferenceRowGroupSizeBytes int64 = 128 * 1024 * 1024

	// SmallRowGroupThresholdBytes is half the reference target.
	SmallRowGroupThresholdBytes int64 = ReferenceRowGroupSizeBytes / 2
)

// SmallRowGroupFindings reports when the average uncompressed row-group
// size is significantly below the reference target.
func SmallRowGroupFindings(dataset summary.Summary) []Finding {
	if dataset.NumRowGroups <= 0 || dataset.UncompressedSize <= 0 {
		return nil
	}

	averageSize := float64(dataset.UncompressedSize) /
		float64(dataset.NumRowGroups)

	if averageSize >= float64(SmallRowGroupThresholdBytes) {
		return nil
	}

	return []Finding{
		NewFinding(
			"Row groups are much smaller than recommended",
			SeverityHigh,
			fmt.Sprintf(
				"Average row group size is %s across %d row groups.",
				formatBytes(averageSize),
				dataset.NumRowGroups,
			),
			"Rewrite the file using a larger target row group size, for example approximately 128 MiB.",
		),
	}
}

func formatBytes(size float64) string {
	if size < 1024 {
		return fmt.Sprintf("%.1f B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KiB", size/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MiB", size/(1024*1024))
	}
	return fmt.Sprintf("%.1f GiB", size/(1024*1024*1024))
}
