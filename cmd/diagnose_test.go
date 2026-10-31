package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	parquetgo "github.com/parquet-go/parquet-go"
)

type diagnoseTestRecord struct {
	ID    int64   `parquet:"id"`
	Email *string `parquet:"email,optional"`
}

func createDiagnoseFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "diagnose.parquet")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[diagnoseTestRecord](file)

	email := "user@example.com"
	records := []diagnoseTestRecord{
		{ID: 1, Email: &email},
		{ID: 2, Email: nil},
		{ID: 3, Email: nil},
	}

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

func TestRunDiagnoseReportsBothFindings(t *testing.T) {
	path := createDiagnoseFixture(t)

	var output bytes.Buffer
	command := diagnoseCmd
	command.SetOut(&output)

	t.Cleanup(func() {
		command.SetOut(nil)
	})

	if err := runDiagnose(command, []string{path}); err != nil {
		t.Fatalf("runDiagnose returned an error: %v", err)
	}

	got := output.String()

	for _, expected := range []string{
		"Diagnostic Results",
		`Column "email" has a high null ratio`,
		"2 nulls out of 3 rows (66.7%)",
		"Severity: medium",
		"Row groups are much smaller than recommended",
		"Severity: high",
		"Recommendation:",
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, got)
		}
	}
}

func TestRunDiagnoseMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.parquet")

	var output bytes.Buffer
	command := diagnoseCmd
	command.SetOut(&output)

	err := runDiagnose(command, []string{path})
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
