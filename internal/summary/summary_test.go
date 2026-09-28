package summary

import "testing"

func TestNewSummary(t *testing.T) {
	cols := []ColumnSummary{
		{
			Name:             "id",
			Type:             "INT64",
			CompressedSize:   100,
			UncompressedSize: 200,
			CompressionCodec: "ZSTD",
		},
		{
			Name:             "name",
			Type:             "BYTE_ARRAY",
			CompressedSize:   50,
			UncompressedSize: 150,
			CompressionCodec: "ZSTD",
		},
	}

	s := NewSummary("parquet", 1000, 4, cols)

	if s.NumRows != 1000 {
		t.Fatal("unexpected row count")
	}
	if s.NumColumns != 2 {
		t.Fatal("unexpected column count")
	}
	if s.NumRowGroups != 4 {
		t.Fatal("unexpected row group count")
	}
	if s.CompressedSize != 150 {
		t.Fatal("unexpected compressed size")
	}
	if s.UncompressedSize != 350 {
		t.Fatal("unexpected uncompressed size")
	}
	if len(s.CompressionCodecs) != 1 {
		t.Fatal("expected one unique codec")
	}
	if s.CompressionRatio <= 0 {
		t.Fatal("invalid compression ratio")
	}
}
