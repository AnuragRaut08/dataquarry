package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AnuragRaut08/dataquarry/internal/parquet"
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
		"id  INT64",
		"name  BYTE_ARRAY (STRING)",
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

func TestRunInspectDoesNotPrintSummaryOnInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.parquet")
	if err := os.WriteFile(path, []byte("not a valid parquet file"), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := inspectCmd
	command.SetOut(&output)
	t.Cleanup(func() {
		command.SetOut(nil)
	})

	err := runInspect(command, []string{path})
	if err == nil {
		t.Fatal("expected an error for an invalid Parquet file")
	}
	if output.Len() != 0 {
		t.Fatalf("expected no partial output after failure, got: %s", output.String())
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("expected error to contain input path, got: %v", err)
	}
}

func TestRunInspectDefaultFormat(t *testing.T) {
	path := createInspectFixture(t)

	command := inspectCmd
	formatFlag := command.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag was not registered")
	}
	if formatFlag.Changed {
		t.Fatal("format flag was explicitly changed; expected its default value")
	}

	format, err := command.Flags().GetString("format")
	if err != nil {
		t.Fatalf("could not read default format: %v", err)
	}
	if format != "text" {
		t.Fatalf("default format = %q, want %q", format, "text")
	}

	var output bytes.Buffer
	command.SetOut(&output)
	t.Cleanup(func() {
		command.SetOut(nil)
	})

	if err := runInspect(command, []string{path}); err != nil {
		t.Fatalf("runInspect returned an error: %v", err)
	}

	metadata, err := parquet.ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata returned an error: %v", err)
	}

	var expected strings.Builder
	fmt.Fprintf(&expected,
		"Dataset Summary — sales.parquet\n"+
			"--------------------------------------------\n"+
			"Rows:               %s\n"+
			"Columns:            %d\n"+
			"Row Groups:         %d\n"+
			"Compressed Size:    %s\n"+
			"Uncompressed Size:  %s\n"+
			"Compression Ratio:  %.1f×\n",
		formatCount(metadata.NumRows),
		metadata.NumColumns,
		metadata.NumRowGroups,
		formatBytes(metadata.CompressedSize),
		formatBytes(metadata.UncompressedSize),
		metadata.CompressionRatio,
	)

	expected.WriteString("Compression Codecs\n")
	if len(metadata.CompressionCodecs) == 0 {
		expected.WriteString("• None\n")
	} else {
		for _, codec := range metadata.CompressionCodecs {
			fmt.Fprintf(&expected, "• %s\n", codec)
		}
	}

	expected.WriteString("Schema\n")
	if len(metadata.Columns) == 0 {
		expected.WriteString("• No columns\n")
	} else {
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
				fmt.Fprintf(&expected, "• %s\n", column.Name)
			} else {
				fmt.Fprintf(&expected, "• %s  %s\n", column.Name, typeLabel)
			}
		}
	}

	if got := output.String(); got != expected.String() {
		t.Errorf("default text output mismatch:\n--- got ---\n%s\n--- want ---\n%s",
			got, expected.String())
	}
}

func TestRunInspectJSON(t *testing.T) {
	path := createInspectFixture(t)
	var output bytes.Buffer

	command := inspectCmd
	formatFlag := command.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag was not registered")
	}
	originalFormat, err := command.Flags().GetString("format")
	if err != nil {
		t.Fatal(err)
	}
	originalChanged := formatFlag.Changed

	if err := command.Flags().Set("format", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = command.Flags().Set("format", originalFormat)
		formatFlag.Changed = originalChanged
	})

	command.SetOut(&output)
	t.Cleanup(func() {
		command.SetOut(nil)
	})

	if err := runInspect(command, []string{path}); err != nil {
		t.Fatalf("runInspect returned an error: %v", err)
	}

	var got inspectJSONOutput
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, output.String())
	}

	metadata, err := parquet.ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata returned an error: %v", err)
	}

	if got.NumRows != metadata.NumRows ||
		got.NumColumns != metadata.NumColumns ||
		got.NumRowGroups != metadata.NumRowGroups ||
		got.CompressedSize != metadata.CompressedSize ||
		got.UncompressedSize != metadata.UncompressedSize ||
		got.CompressionRatio != metadata.CompressionRatio {
		t.Errorf("JSON metadata does not match Parquet metadata: got %+v, want %+v", got, metadata)
	}

	if !reflect.DeepEqual(got.CompressionCodecs, metadata.CompressionCodecs) {
		t.Errorf("CompressionCodecs = %v, want %v", got.CompressionCodecs, metadata.CompressionCodecs)
	}
	if !reflect.DeepEqual(got.Schema, metadata.Schema) {
		t.Errorf("Schema = %v, want %v", got.Schema, metadata.Schema)
	}
	if len(got.Columns) != len(metadata.Columns) {
		t.Fatalf("got %d JSON columns, want %d", len(got.Columns), len(metadata.Columns))
	}
	for i, column := range metadata.Columns {
		want := inspectJSONColumn{
			Name:             column.Name,
			Type:             column.Type,
			LogicalType:      column.LogicalType,
			CompressedSize:   column.CompressedSize,
			UncompressedSize: column.UncompressedSize,
			NullCount:        column.NullCount,
			NullCountKnown:   column.NullCountKnown,
			CompressionCodec: column.CompressionCodec,
		}
		if got.Columns[i] != want {
			t.Errorf("JSON column %d = %+v, want %+v", i, got.Columns[i], want)
		}
	}

	if got.NumRows != 2 {
		t.Errorf("NumRows = %d, want 2", got.NumRows)
	}
	if got.NumColumns != 2 {
		t.Errorf("NumColumns = %d, want 2", got.NumColumns)
	}
	if got.NumRowGroups < 1 {
		t.Errorf("NumRowGroups = %d, want at least 1", got.NumRowGroups)
	}
	if got.CompressedSize <= 0 {
		t.Errorf("CompressedSize = %d, want positive value", got.CompressedSize)
	}
	if got.UncompressedSize <= 0 {
		t.Errorf("UncompressedSize = %d, want positive value", got.UncompressedSize)
	}
	if got.CompressionRatio <= 0 {
		t.Errorf("CompressionRatio = %f, want positive value", got.CompressionRatio)
	}
	if len(got.CompressionCodecs) == 0 {
		t.Error("expected at least one compression codec")
	}
	if len(got.Schema) != 2 || got.Schema[0] != "id" || got.Schema[1] != "name" {
		t.Errorf("Schema = %v, want [id name]", got.Schema)
	}
	if len(got.Columns) != 2 {
		t.Fatalf("got %d column summaries, want 2", len(got.Columns))
	}
	if got.Columns[0].Name != "id" || got.Columns[1].Name != "name" {
		t.Errorf("unexpected column summaries: %+v", got.Columns)
	}

	if strings.Contains(output.String(), "Dataset Summary") {
		t.Error("JSON output contains human-readable text")
	}
}

func TestRunInspectUnsupportedFormat(t *testing.T) {
	command := inspectCmd
	formatFlag := command.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag was not registered")
	}
	originalFormat, err := command.Flags().GetString("format")
	if err != nil {
		t.Fatal(err)
	}
	originalChanged := formatFlag.Changed

	if err := command.Flags().Set("format", "xml"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = command.Flags().Set("format", originalFormat)
		formatFlag.Changed = originalChanged
	})

	err = runInspect(command, []string{"sales.parquet"})
	if err == nil {
		t.Fatal("expected an error for unsupported output format")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildInspectJSONResponseEmptySlices(t *testing.T) {
	metadata := parquet.Metadata{}

	response := buildInspectJSONResponse(&metadata)

	if response.CompressionCodecs == nil {
		t.Fatal("CompressionCodecs should be an empty slice, not nil")
	}
	if response.Schema == nil {
		t.Fatal("Schema should be an empty slice, not nil")
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal JSON response: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if string(got["compression_codecs"]) != "[]" {
		t.Errorf("compression_codecs = %s, want []", got["compression_codecs"])
	}
	if string(got["schema"]) != "[]" {
		t.Errorf("schema = %s, want []", got["schema"])
	}
}
