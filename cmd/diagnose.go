package cmd

import (
	"fmt"

	"github.com/AnuragRaut08/dataquarry/internal/diagnose"
	"github.com/AnuragRaut08/dataquarry/internal/parquet"
	"github.com/AnuragRaut08/dataquarry/internal/summary"
	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose <file.parquet>",
	Short: "Analyze a dataset for storage issues",
	Long: `Analyze a Parquet dataset and report structured storage findings.

DataQuarry identifies inefficient storage patterns and provides
evidence-backed recommendations without modifying the dataset.`,
	Args: cobra.ExactArgs(1),
	RunE: runDiagnose,
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	metadata, err := parquet.ReadMetadata(args[0])
	if err != nil {
		return fmt.Errorf("diagnose %q: %w", args[0], err)
	}

	dataset := summary.NewSummary(
		"parquet",
		metadata.NumRows,
		metadata.NumRowGroups,
		metadata.Columns,
	)

	findings := diagnose.HighNullRatioFindings(
		metadata.NumRows,
		metadata.Columns,
	)

	findings = append(
		findings,
		diagnose.SmallRowGroupFindings(dataset)...,
	)

	out := cmd.OutOrStdout()

	if len(findings) == 0 {
		_, err := fmt.Fprintln(out, "No diagnostic findings.")
		return err
	}

	if _, err := fmt.Fprintf(out, "Diagnostic Results — %s\n", args[0]); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "--------------------------------------------"); err != nil {
		return err
	}

	for i, finding := range findings {
		if _, err := fmt.Fprintf(
			out,
			"Finding %d: %s\nSeverity: %s\nEvidence: %s\nRecommendation: %s\n",
			i+1,
			finding.Title,
			finding.Severity,
			finding.Evidence,
			finding.Fix,
		); err != nil {
			return err
		}

		if i < len(findings)-1 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)
}
