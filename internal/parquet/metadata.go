package parquet

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
	parquetgo "github.com/parquet-go/parquet-go"
)

// Metadata contains structural measurements extracted from a Parquet footer.
type Metadata struct {
	NumRows           int64
	NumColumns        int
	NumRowGroups      int
	CompressedSize    int64
	UncompressedSize  int64
	CompressionRatio  float64
	Schema            []string
	CompressionCodecs []string
}

// ReadMetadata reads structural metadata from a Parquet file footer.
func ReadMetadata(path string) (*Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open parquet file %q: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat parquet file %q: %w", path, err)
	}
	if info.Size() < 8 {
		return nil, fmt.Errorf("invalid parquet file %q: file is too small", path)
	}

	pf, err := parquetgo.OpenFile(file, info.Size())
	if err != nil {
		return nil, fmt.Errorf("read parquet metadata from %q: %w", path, err)
	}

	footer := pf.Metadata()
	schemaColumns := pf.Schema().Columns()
	schema := make([]string, 0, len(schemaColumns))
	columnSummaries := make([]summary.ColumnSummary, 0, len(schemaColumns))
	columnIndexes := make(map[string]int, len(schemaColumns))

	for _, pathParts := range schemaColumns {
		name := strings.Join(pathParts, ".")
		columnIndexes[name] = len(columnSummaries)
		schema = append(schema, name)
		columnSummaries = append(columnSummaries, summary.ColumnSummary{Name: name})
	}

	codecSet := make(map[string]struct{})
	var compressedSize, uncompressedSize int64

	for _, rowGroup := range footer.RowGroups {
		for _, column := range rowGroup.Columns {
			columnMetadata := column.MetaData
			name := strings.Join(columnMetadata.PathInSchema, ".")
			codec := columnMetadata.Codec.String()

			compressedSize += columnMetadata.TotalCompressedSize
			uncompressedSize += columnMetadata.TotalUncompressedSize

			if codec != "" {
				codecSet[codec] = struct{}{}
			}

			if index, ok := columnIndexes[name]; ok {
				columnSummaries[index].CompressedSize += columnMetadata.TotalCompressedSize
				columnSummaries[index].UncompressedSize += columnMetadata.TotalUncompressedSize
				columnSummaries[index].Type = columnMetadata.Type.String()
				columnSummaries[index].CompressionCodec = codec
			}
		}
	}

	codecs := make([]string, 0, len(codecSet))
	for codec := range codecSet {
		codecs = append(codecs, codec)
	}
	sort.Strings(codecs)

	// Keep the shared summary model as the source of derived measurements.
	datasetSummary := summary.NewSummary(
		"parquet",
		footer.NumRows,
		len(footer.RowGroups),
		columnSummaries,
	)

	// Footer column-chunk totals also cover files with multiple row groups.
	// Use these totals directly so each chunk is counted exactly once.
	ratio := 0.0
	if compressedSize > 0 {
		ratio = float64(uncompressedSize) / float64(compressedSize)
	}

	return &Metadata{
		NumRows:           datasetSummary.NumRows,
		NumColumns:        len(schemaColumns),
		NumRowGroups:      datasetSummary.NumRowGroups,
		CompressedSize:    compressedSize,
		UncompressedSize:  uncompressedSize,
		CompressionRatio:  ratio,
		Schema:            schema,
		CompressionCodecs: codecs,
	}, nil
}
