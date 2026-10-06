package parquet

import (
	"fmt"
	"os"

	parquetgo "github.com/parquet-go/parquet-go"
)

// Metadata contains structural metadata extracted from a Parquet file.
type Metadata struct {
	NumRows           int64
	NumRowGroups      int
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

	metadata := pf.Metadata()

	schema := make([]string, 0)
	for i, element := range metadata.Schema {
		if i == 0 {
			continue
		}
		if element.Name != "" {
			schema = append(schema, element.Name)
		}
	}

	codecSet := make(map[string]struct{})
	codecs := make([]string, 0)

	for _, rowGroup := range metadata.RowGroups {
		for _, column := range rowGroup.Columns {
			codec := column.MetaData.Codec.String()
			if codec == "" {
				continue
			}

			if _, exists := codecSet[codec]; exists {
				continue
			}

			codecSet[codec] = struct{}{}
			codecs = append(codecs, codec)
		}
	}

	return &Metadata{
		NumRows:           metadata.NumRows,
		NumRowGroups:      len(metadata.RowGroups),
		Schema:            schema,
		CompressionCodecs: codecs,
	}, nil
}
