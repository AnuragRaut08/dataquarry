package parquet

import (
	"testing"

	"github.com/parquet-go/parquet-go/format"
)

func TestCompressionCodecName(t *testing.T) {
	tests := []struct {
		name  string
		codec format.CompressionCodec
		want  string
	}{
		{"uncompressed", format.Uncompressed, "UNCOMPRESSED"},
		{"snappy", format.Snappy, "SNAPPY"},
		{"gzip", format.Gzip, "GZIP"},
		{"lzo", format.LZO, "LZO"},
		{"brotli", format.Brotli, "BROTLI"},
		{"lz4", format.Lz4, "LZ4"},
		{"zstd", format.Zstd, "ZSTD"},
		{"lz4 raw", format.Lz4Raw, "LZ4_RAW"},
		{"unknown", format.CompressionCodec(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compressionCodecName(tt.codec); got != tt.want {
				t.Errorf("compressionCodecName(%d) = %q, want %q",
					tt.codec, got, tt.want)
			}
		})
	}
}
