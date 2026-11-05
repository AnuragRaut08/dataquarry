package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	parquetgo "github.com/parquet-go/parquet-go"
)

type diagnoseTestRecord struct {
	ID    int64   `parquet:"id"`
	Email *string `parquet:"email,optional"`
}

func createDiagnoseFixture(t *testing.T, records []diagnoseTestRecord) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "diagnose.parquet")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[diagnoseTestRecord](file)

	_, writeErr := writer.Write(records)
	closeWriterErr := writer.Close()
	closeFileErr := file.Close()

	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeWriterErr != nil {
		t.Fatal(closeWriterErr)
	}
	if closeFileErr != nil {
		t.Fatal(closeFileErr)
	}

	return path
}

func runDiagnoseForTest(t *testing.T, path string) (string, error) {
	t.Helper()

	var output bytes.Buffer
	command := diagnoseCmd
	command.SetOut(&output)

	t.Cleanup(func() {
		command.SetOut(nil)
	})

	err := runDiagnose(command, []string{path})
	return output.String(), err
}

func TestRunDiagnoseReportsBothFindings(t *testing.T) {
	email := "user@example.com"

	path := createDiagnoseFixture(t, []diagnoseTestRecord{
		{ID: 1, Email: &email},
		{ID: 2, Email: nil},
		{ID: 3, Email: nil},
	})

	output, err := runDiagnoseForTest(t, path)
	if err != nil {
		t.Fatalf("runDiagnose returned an error: %v", err)
	}

	for _, expected := range []string{
		`Column "email" has a high null ratio`,
		"2 nulls out of 3 rows (66.7%)",
		"Severity: medium",
		"Row groups are much smaller than recommended",
		"Severity: high",
		"Recommendation:",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestRunDiagnoseMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.parquet")

	_, err := runDiagnoseForTest(t, path)
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}

	if !strings.Contains(err.Error(), "open parquet file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiagnoseRequiresExactlyOneArgument(t *testing.T) {
	if err := diagnoseCmd.Args(diagnoseCmd, nil); err == nil {
		t.Fatal("expected an error when no file is provided")
	}

	if err := diagnoseCmd.Args(
		diagnoseCmd,
		[]string{"one.parquet", "two.parquet"},
	); err == nil {
		t.Fatal("expected an error when multiple files are provided")
	}
}

func TestRunDiagnoseDirectoryInput(t *testing.T) {
	path := t.TempDir()

	_, err := runDiagnoseForTest(t, path)
	if err == nil {
		t.Fatal("expected an error when input is a directory")
	}

	if !strings.Contains(err.Error(), "diagnose") {
		t.Fatalf("expected diagnose context in error, got: %v", err)
	}

	if !strings.Contains(err.Error(), path) {
		t.Fatalf("expected error to contain directory path %q, got: %v", path, err)
	}
}

func TestCLICommandsReturnNonZeroExitStatusForDirectoryInput(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "dataquarry")

	build := exec.Command("go", "build", "-o", binaryPath, "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build CLI: %v\n%s", err, output)
	}

	for _, commandName := range []string{"inspect", "diagnose"} {
		t.Run(commandName, func(t *testing.T) {
			directoryPath := t.TempDir()

			command := exec.Command(binaryPath, commandName, directoryPath)
			output, err := command.CombinedOutput()

			if err == nil {
				t.Fatalf(
					"expected %s to exit with a non-zero status for directory %q; output: %s",
					commandName,
					directoryPath,
					output,
				)
			}

			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected process exit error, got: %v", err)
			}

			if exitErr.ExitCode() == 0 {
				t.Fatalf("expected non-zero exit status, got 0")
			}
		})
	}
}
