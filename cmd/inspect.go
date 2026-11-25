package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/AnuragRaut08/dataquarry/internal/parquet"
	"github.com/spf13/cobra"
)

var inspectFormat string

var inspectCmd = &cobra.Command{
	Use:   "inspect <file.parquet>",
	Short: "Inspect Parquet metadata",
	Long: `Inspect a Parquet file and display its structural metadata,
including row counts, schema, row groups, and compression statistics.`,
	Args: requireExactlyOneInput,
	RunE: runInspect,
}

type inspectJSONColumn struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	LogicalType      string `json:"logical_type"`
	CompressedSize   int64  `json:"compressed_size"`
	UncompressedSize int64  `json:"uncompressed_size"`
	NullCount        int64  `json:"null_count"`
	NullCountKnown   bool   `json:"null_count_known"`
	CompressionCodec string `json:"compression_codec"`
}

type inspectJSONOutput struct {
	NumRows           int64               `json:"num_rows"`
	NumColumns        int                 `json:"num_columns"`
	NumRowGroups      int                 `json:"num_row_groups"`
	CompressedSize    int64               `json:"compressed_size"`
	UncompressedSize  int64               `json:"uncompressed_size"`
	CompressionRatio  float64             `json:"compression_ratio"`
	CompressionCodecs []string            `json:"compression_codecs"`
	Schema            []string            `json:"schema"`
	Columns           []inspectJSONColumn `json:"columns"`
}

func buildInspectJSONResponse(metadata *parquet.Metadata) inspectJSONOutput {
	response := inspectJSONOutput{
		NumRows:           metadata.NumRows,
		NumColumns:        metadata.NumColumns,
		NumRowGroups:      metadata.NumRowGroups,
		CompressedSize:    metadata.CompressedSize,
		UncompressedSize:  metadata.UncompressedSize,
		CompressionRatio:  metadata.CompressionRatio,
		CompressionCodecs: append([]string{}, metadata.CompressionCodecs...),
		Schema:            append([]string{}, metadata.Schema...),
		Columns:           make([]inspectJSONColumn, len(metadata.Columns)),
	}

	for i, column := range metadata.Columns {
		response.Columns[i] = inspectJSONColumn{
			Name:             column.Name,
			Type:             column.Type,
			LogicalType:      column.LogicalType,
			CompressedSize:   column.CompressedSize,
			UncompressedSize: column.UncompressedSize,
			NullCount:        column.NullCount,
			NullCountKnown:   column.NullCountKnown,
			CompressionCodec: column.CompressionCodec,
		}
	}

	return response
}

func runInspect(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("read output format: %w", err)
	}

	if format != "text" && format != "json" {
		return fmt.Errorf("unsupported output format %q: supported formats are text and json", format)
	}

	metadata, err := parquet.ReadMetadata(args[0])
	if err != nil {
		return fmt.Errorf("inspect %q: %w", args[0], err)
	}

	out := cmd.OutOrStdout()

	if format == "json" {
		response := buildInspectJSONResponse(metadata)

		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("write JSON output: %w", err)
		}
		return nil
	}

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
	if len(metadata.Columns) == 0 {
		_, err := fmt.Fprintln(out, "• No columns")
		return err
	}
	for _, column := range metadata.Columns {
		typeLabel := column.Type
		if column.LogicalType != "" {
			if typeLabel != "" {
				typeLabel += " (" + column.LogicalType + ")"
			} else {
				typeLabel = column.LogicalType
			}
		}

		if typeLabel == "" {
			if _, err := fmt.Fprintf(out, "• %s\n", column.Name); err != nil {
				return err
			}
		} else if _, err := fmt.Fprintf(out, "• %s  %s\n", column.Name, typeLabel); err != nil {
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
	inspectCmd.Flags().StringVar(
		&inspectFormat,
		"format",
		"text",
		"Output format (text or json)",
	)
	rootCmd.AddCommand(inspectCmd)
}
