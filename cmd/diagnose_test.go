
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

func TestRunDiagnoseReportsFindings(t *testing.T) {
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
		"Diagnostic Results",
		"high null ratio",
		"2 nulls out of 3 rows (66.7%)",
		"Row groups are much smaller than recommended",
		"Severity: medium",
		"Severity: high",
		"Recommendation:",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestRunDiagnoseNoFindings(t *testing.T) {
	email := "user@example.com"
	path := createDiagnoseFixture(t, []diagnoseTestRecord{
		{ID: 1, Email: &email},
		{ID: 2, Email: &email},
	})

	output, err := runDiagnoseForTest(t, path)
	if err != nil {
		t.Fatalf("runDiagnose returned an error: %v", err)
	}

	if !strings.Contains(output, "No diagnostic findings.") {
		t.Fatalf("expected no-findings message, got:\n%s", output)
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
