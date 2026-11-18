package parquet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AnuragRaut08/dataquarry/internal/summary"
	parquetgo "github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/format"
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
	Columns           []summary.ColumnSummary
}

type columnTypeInfo struct {
	PhysicalType string
	LogicalType  string
}

// compressionCodecName returns a stable display name for known codecs and
// preserves unknown codec values without relying on generated String methods.
func compressionCodecName(codec format.CompressionCodec) string {
	switch codec {
	case format.Uncompressed:
		return "UNCOMPRESSED"
	case format.Snappy:
		return "SNAPPY"
	case format.Gzip:
		return "GZIP"
	case format.LZO:
		return "LZO"
	case format.Brotli:
		return "BROTLI"
	case format.Lz4:
		return "LZ4"
	case format.Zstd:
		return "ZSTD"
	case format.Lz4Raw:
		return "LZ4_RAW"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int32(codec))
	}
}

// collectColumnTypes walks the Parquet schema and records types for leaf columns.
func collectColumnTypes(node parquetgo.Node, path []string, result map[string]columnTypeInfo) {
	if node.Leaf() {
		typ := node.Type()
		info := columnTypeInfo{
			PhysicalType: typ.Kind().String(),
		}

		if logical := typ.LogicalType(); logical != nil && logical.Value != nil {
			info.LogicalType = logical.Value.String()
		}

		result[strings.Join(path, ".")] = info
		return
	}

	for _, field := range node.Fields() {
		childPath := append(append([]string(nil), path...), field.Name())
		collectColumnTypes(field, childPath, result)
	}
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
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv", ".tsv", ".json", ".jsonl", ".ndjson", ".avro", ".orc":
		return nil, fmt.Errorf(
			"unsupported input format %q for %q: only Parquet files are supported",
			ext,
			path,
		)
	}

	if info.Size() < 8 {
		return nil, fmt.Errorf("invalid parquet file %q: file is too small", path)
	}

	pf, err := parquetgo.OpenFile(file, info.Size())
	if err != nil {
		return nil, fmt.Errorf("read parquet metadata from %q: %w", path, err)
	}

	footer := pf.Metadata()
	parquetSchema := pf.Schema()
	schemaColumns := parquetSchema.Columns()
	schema := make([]string, 0, len(schemaColumns))
	columnSummaries := make([]summary.ColumnSummary, 0, len(schemaColumns))
	columnIndexes := make(map[string]int, len(schemaColumns))
	rowGroupsWithColumn := make([]int, len(schemaColumns))
	rowGroupsWithNullStats := make([]int, len(schemaColumns))

	columnTypes := make(map[string]columnTypeInfo, len(schemaColumns))
	collectColumnTypes(parquetSchema, nil, columnTypes)

	for _, pathParts := range schemaColumns {
		name := strings.Join(pathParts, ".")
		columnIndexes[name] = len(columnSummaries)
		schema = append(schema, name)

		typeInfo := columnTypes[name]
		columnSummaries = append(columnSummaries, summary.ColumnSummary{
			Name:        name,
			Type:        typeInfo.PhysicalType,
			LogicalType: typeInfo.LogicalType,
		})
	}

	codecSet := make(map[string]struct{})
	var compressedSize, uncompressedSize int64

	for _, rowGroup := range footer.RowGroups {
		for _, column := range rowGroup.Columns {
			columnMetadata := column.MetaData
			name := strings.Join(columnMetadata.PathInSchema, ".")
			codec := compressionCodecName(columnMetadata.Codec)

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

				rowGroupsWithColumn[index]++

				if columnMetadata.Statistics.NullCount > 0 {
					columnSummaries[index].NullCount += columnMetadata.Statistics.NullCount
					rowGroupsWithNullStats[index]++
				}
			}
		}
	}

	for i := range columnSummaries {
		columnSummaries[i].NullCountKnown =
			rowGroupsWithColumn[i] > 0 &&
				rowGroupsWithColumn[i] == rowGroupsWithNullStats[i]
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
		Columns:           columnSummaries,
	}, nil
}
