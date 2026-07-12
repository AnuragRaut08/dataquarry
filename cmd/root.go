package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dataquarry",
	Short: "Inspect, diagnose, and optimize data files",
	Long: `DataQuarry is a CLI for data engineers that helps inspect,
diagnose, benchmark, and optimize data files using evidence
instead of guesswork.

Starting with Parquet, DataQuarry provides actionable insights
about schema, storage layout, compression, null ratios, and
performance recommendations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags will go here in future.
}