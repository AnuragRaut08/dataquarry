package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/AnuragRaut08/dataquarry/internal/parquet"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <file.parquet>",
	Short: "Inspect Parquet metadata",
	Long: `Inspect a Parquet file and display its structural metadata,
including row counts, schema, row groups, and compression statistics.`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

func runInspect(cmd *cobra.Command, args []string) error {
	metadata, err := parquet.ReadMetadata(args[0])
	if err != nil {
		return fmt.Errorf("inspect %q: %w", args[0], err)
	}

	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out,
		"Dataset Summary — %s\n"+
			"--------------------------------------------\n"+
			"Rows:               %s\n"+
			"Columns:            %d\n"+
			"Row Groups:         %d\n"+
			"Compressed Size:    %s\n"+
			"Uncompressed Size:  %s\n"+
			"Compression Ratio:  %.1f×\n",
		filepath.Base(args[0]),
		formatCount(metadata.NumRows),
		metadata.NumColumns,
		metadata.NumRowGroups,
		formatBytes(metadata.CompressedSize),
		formatBytes(metadata.UncompressedSize),
		metadata.CompressionRatio,
	); err != nil {
		return fmt.Errorf("write dataset summary: %w", err)
	}

	if _, err := fmt.Fprintln(out, "Compression Codecs"); err != nil {
		return err
	}
	if len(metadata.CompressionCodecs) == 0 {
		if _, err := fmt.Fprintln(out, "• None"); err != nil {
			return err
		}
	} else {
		for _, codec := range metadata.CompressionCodecs {
			if _, err := fmt.Fprintf(out, "• %s\n", codec); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(out, "Schema"); err != nil {
		return err
	}
	if len(metadata.Schema) == 0 {
		_, err := fmt.Fprintln(out, "• No columns")
		return err
	}
	for _, field := range metadata.Schema {
		if _, err := fmt.Fprintf(out, "• %s\n", field); err != nil {
			return err
		}
	}

	return nil
}

func formatCount(n int64) string {
	// Keep the count readable for large datasets.
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func formatBytes(n int64) string {
	if n < 0 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}

	const unit = 1024
	value := float64(n)
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	index := 0
	for value >= unit && index < len(units)-1 {
		value /= unit
		index++
	}
	return fmt.Sprintf("%.1f %s", value, units[index])
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
