package parquet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
	parquetgo "github.com/parquet-go/parquet-go"
)

func TestReadMetadataMissingFile(t *testing.T) {
	_, err := ReadMetadata(filepath.Join(t.TempDir(), "missing.parquet"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	if !strings.Contains(err.Error(), "open parquet file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadMetadataTooSmallFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.parquet")

	if err := os.WriteFile(path, []byte("PAR1"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadMetadata(path)
	if err == nil {
		t.Fatal("expected error for invalid parquet file")
	}

	if !strings.Contains(err.Error(), "file is too small") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadMetadataCorruptedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupted.parquet")

	data := make([]byte, 8)
	copy(data[:4], []byte("PAR1"))
	copy(data[4:], []byte("PAR1"))

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadMetadata(path)
	if err == nil {
		t.Fatal("expected error for corrupted parquet file")
	}
}

type testRecord struct {
	ID   int64  `parquet:"id"`
	Name string `parquet:"name"`
}

func TestReadMetadataValidParquet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "valid.parquet")

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[testRecord](file)

	_, err = writer.Write([]testRecord{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	})
	if err != nil {
		file.Close()
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		file.Close()
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	metadata, err := ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if metadata.NumRows != 2 {
		t.Fatalf("expected 2 rows, got %d", metadata.NumRows)
	}

	if metadata.NumColumns != 2 {
		t.Fatalf("expected 2 columns, got %d", metadata.NumColumns)
	}

	if metadata.CompressedSize <= 0 {
		t.Fatalf("expected positive compressed size, got %d", metadata.CompressedSize)
	}

	if metadata.UncompressedSize <= 0 {
		t.Fatalf("expected positive uncompressed size, got %d", metadata.UncompressedSize)
	}

	if metadata.CompressionRatio <= 0 {
		t.Fatalf("expected positive compression ratio, got %f", metadata.CompressionRatio)
	}

	expectedRatio := float64(metadata.UncompressedSize) / float64(metadata.CompressedSize)
	if metadata.CompressionRatio != expectedRatio {
		t.Fatalf("expected compression ratio %f, got %f", expectedRatio, metadata.CompressionRatio)
	}

	if metadata.NumRowGroups < 1 {
		t.Fatalf("expected at least 1 row group, got %d", metadata.NumRowGroups)
	}

	if len(metadata.Schema) != 2 {
		t.Fatalf("expected 2 schema fields, got %d: %v", len(metadata.Schema), metadata.Schema)
	}

	if len(metadata.CompressionCodecs) == 0 {
		t.Fatal("expected at least one compression codec")
	}
}

type nullableTestRecord struct {
	ID    int64   `parquet:"id"`
	Email *string `parquet:"email"`
}

func TestReadMetadataNullCountStatistics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "null-counts.parquet")

	email := "alice@example.com"
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[nullableTestRecord](file)
	_, err = writer.Write([]nullableTestRecord{
		{ID: 1, Email: &email},
		{ID: 2, Email: nil},
		{ID: 3, Email: nil},
	})
	if err != nil {
		_ = writer.Close()
		_ = file.Close()
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	metadata, err := ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	columns := make(map[string]summary.ColumnSummary)
	for _, column := range metadata.Columns {
		columns[column.Name] = column
	}

	emailColumn, ok := columns["email"]
	if !ok {
		t.Fatal("email column not found")
	}
	if emailColumn.NullCount != 2 {
		t.Errorf("email null count = %d, want 2", emailColumn.NullCount)
	}
	if !emailColumn.NullCountKnown {
		t.Error("email null count should be known when positive statistics are present")
	}

	idColumn, ok := columns["id"]
	if !ok {
		t.Fatal("id column not found")
	}
	if idColumn.NullCountKnown {
		t.Error("id zero null count should remain unknown with the current conservative implementation")
	}
}

func TestReadMetadataZeroNullCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "zero-nulls.parquet")

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := parquetgo.NewGenericWriter[testRecord](file)
	_, err = writer.Write([]testRecord{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	})
	if err != nil {
		_ = writer.Close()
		_ = file.Close()
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	metadata, err := ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if len(metadata.Columns) == 0 {
		t.Fatal("expected columns in metadata")
	}
	for _, column := range metadata.Columns {
		if column.NullCount != 0 {
			t.Errorf("column %q: expected zero nulls, got %d", column.Name, column.NullCount)
		}
		if column.NullCountKnown {
			t.Errorf("column %q: expected zero null count to remain unknown", column.Name)
		}
	}
}

func TestReadMetadataInspectionFixture(t *testing.T) {
	path := filepath.Join("testdata", "inspection.parquet")

	metadata, err := ReadMetadata(path)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if metadata.NumRows != 6 {
		t.Errorf("NumRows = %d, want 6", metadata.NumRows)
	}
	if metadata.NumColumns != 4 {
		t.Errorf("NumColumns = %d, want 4", metadata.NumColumns)
	}
	if metadata.NumRowGroups != 3 {
		t.Errorf("NumRowGroups = %d, want 3", metadata.NumRowGroups)
	}

	expectedSchema := []string{"id", "name", "department", "salary"}
	if len(metadata.Schema) != len(expectedSchema) {
		t.Fatalf("Schema = %v, want %v", metadata.Schema, expectedSchema)
	}
	for i, expected := range expectedSchema {
		if metadata.Schema[i] != expected {
			t.Errorf("Schema[%d] = %q, want %q",
				i, metadata.Schema[i], expected)
		}
	}

	expectedCodecs := []string{"SNAPPY"}
	if len(metadata.CompressionCodecs) != len(expectedCodecs) {
		t.Fatalf("CompressionCodecs = %v, want %v",
			metadata.CompressionCodecs, expectedCodecs)
	}
	for i, expected := range expectedCodecs {
		if metadata.CompressionCodecs[i] != expected {
			t.Errorf("CompressionCodecs[%d] = %q, want %q",
				i, metadata.CompressionCodecs[i], expected)
		}
	}
}
