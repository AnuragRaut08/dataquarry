package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func requireExactlyOneInput(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf(
			"expected exactly one input file\\nUsage: %s",
			cmd.UseLine(),
		)
	}
	return nil
}
