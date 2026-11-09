package cmd

import (
	"strings"
	"testing"
)

func TestRequireExactlyOneInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  interface {
			UseLine() string
		}
	}{
		{name: "inspect", cmd: inspectCmd},
		{name: "diagnose", cmd: diagnoseCmd},
	} {
		t.Run(tc.name+"/missing argument", func(t *testing.T) {
			var command = inspectCmd
			if tc.name == "diagnose" {
				command = diagnoseCmd
			}
			err := requireExactlyOneInput(command, nil)
			if err == nil || !strings.Contains(err.Error(), "Usage:") {
				t.Fatalf("expected actionable usage error, got: %v", err)
			}
		})

		t.Run(tc.name+"/extra argument", func(t *testing.T) {
			var command = inspectCmd
			if tc.name == "diagnose" {
				command = diagnoseCmd
			}
			err := requireExactlyOneInput(
				command,
				[]string{"one.parquet", "two.parquet"},
			)
			if err == nil {
				t.Fatal("expected an error for unexpected arguments")
			}
		})
	}
}
