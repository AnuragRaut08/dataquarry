package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	parquetgo "github.com/parquet-go/parquet-go"
)

type inspectTestRecord struct {
	ID   int64  `parquet:"id"`
	Name string `parquet:"name"`
}

func createInspectFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "sales.parquet")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[inspectTestRecord](file)
	_, writeErr := writer.Write([]inspectTestRecord{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	})
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

func TestRunInspect(t *testing.T) {
	path := createInspectFixture(t)
	var output bytes.Buffer

	command := inspectCmd
	command.SetOut(&output)
	t.Cleanup(func() {
		command.SetOut(nil)
	})

	if err := runInspect(command, []string{path}); err != nil {
		t.Fatalf("runInspect returned an error: %v", err)
	}

	got := output.String()
	for _, expected := range []string{
		"Dataset Summary — sales.parquet",
		"Rows:",
		"2",
		"Columns:",
		"Row Groups:",
		"Compressed Size:",
		"Uncompressed Size:",
		"Compression Ratio:",
		"Compression Codecs",
		"Schema",
		"id",
		"name",
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, got)
		}
	}
}

func TestRunInspectMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.parquet")
	command := inspectCmd

	err := runInspect(command, []string{path})
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if !strings.Contains(err.Error(), "open parquet file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInspectRequiresExactlyOneArgument(t *testing.T) {
	if err := inspectCmd.Args(inspectCmd, nil); err == nil {
		t.Fatal("expected an error when no file is provided")
	}
	if err := inspectCmd.Args(inspectCmd, []string{"one.parquet", "two.parquet"}); err == nil {
		t.Fatal("expected an error when multiple files are provided")
	}
}

func TestRunInspectInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.parquet")

	if err := os.WriteFile(path, []byte("not a valid parquet file"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := runInspect(inspectCmd, []string{path})
	if err == nil {
		t.Fatal("expected an error for an invalid Parquet file")
	}

	if !strings.Contains(err.Error(), "inspect") {
		t.Fatalf("expected inspect context in error, got: %v", err)
	}
}
