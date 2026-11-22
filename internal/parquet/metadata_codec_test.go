package parquet

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	parquetgo "github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/encoding/thrift"
	"github.com/parquet-go/parquet-go/format"
)

func TestReadMetadataUnknownCompressionCodec(t *testing.T) {
	dir := t.TempDir()
	originalPath := filepath.Join(dir, "original.parquet")
	unknownPath := filepath.Join(dir, "unknown-codec.parquet")

	file, err := os.Create(originalPath)
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

	data, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) < 8 {
		t.Fatal("generated Parquet file is too small")
	}

	footerSize := int(binary.LittleEndian.Uint32(data[len(data)-8 : len(data)-4]))
	footerStart := len(data) - 8 - footerSize
	if footerStart < 4 {
		t.Fatalf("invalid footer size: %d", footerSize)
	}

	protocol := &thrift.CompactProtocol{}
	var footer format.FileMetaData

	if err := thrift.Unmarshal(protocol, data[footerStart:footerStart+footerSize], &footer); err != nil {
		t.Fatalf("decode Parquet footer: %v", err)
	}

	modifiedColumns := 0
	for i := range footer.RowGroups {
		for j := range footer.RowGroups[i].Columns {
			footer.RowGroups[i].Columns[j].MetaData.Codec = format.CompressionCodec(99)
			modifiedColumns++
		}
	}

	if modifiedColumns == 0 {
		t.Fatal("generated Parquet file contains no column chunks")
	}

	encodedFooter, err := thrift.Marshal(protocol, &footer)
	if err != nil {
		t.Fatalf("encode modified Parquet footer: %v", err)
	}

	modifiedData := make([]byte, 0, footerStart+len(encodedFooter)+8)
	modifiedData = append(modifiedData, data[:footerStart]...)
	modifiedData = append(modifiedData, encodedFooter...)

	var sizeBytes [4]byte
	binary.LittleEndian.PutUint32(sizeBytes[:], uint32(len(encodedFooter)))
	modifiedData = append(modifiedData, sizeBytes[:]...)
	modifiedData = append(modifiedData, []byte("PAR1")...)

	if err := os.WriteFile(unknownPath, modifiedData, 0o644); err != nil {
		t.Fatal(err)
	}

	metadata, err := ReadMetadata(unknownPath)
	if err != nil {
		t.Fatalf("ReadMetadata failed for unknown codec: %v", err)
	}

	if len(metadata.CompressionCodecs) != 1 ||
		metadata.CompressionCodecs[0] != "UNKNOWN(99)" {
		t.Fatalf("dataset codecs = %v, want [UNKNOWN(99)]", metadata.CompressionCodecs)
	}

	if len(metadata.Columns) != 2 {
		t.Fatalf("got %d column summaries, want 2", len(metadata.Columns))
	}

	for _, column := range metadata.Columns {
		if column.CompressionCodec != "UNKNOWN(99)" {
			t.Errorf(
				"column %q codec = %q, want UNKNOWN(99)",
				column.Name,
				column.CompressionCodec,
			)
		}
	}
}
