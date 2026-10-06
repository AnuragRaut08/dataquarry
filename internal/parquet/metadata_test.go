package parquet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
