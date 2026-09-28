package summary

// ColumnSummary contains derived measurements for a dataset column.
type ColumnSummary struct {
	Name             string
	Type             string
	CompressedSize   int64
	UncompressedSize int64
	NullCount        int64
	CompressionCodec string
}

// Summary is a reusable, presentation-independent dataset summary.
type Summary struct {
	Format            string
	NumRows           int64
	NumColumns        int
	NumRowGroups      int
	CompressedSize    int64
	UncompressedSize  int64
	CompressionRatio  float64
	CompressionCodecs []string
	Columns           []ColumnSummary
}

// NewSummary creates a reusable dataset summary.
func NewSummary(
	format string,
	numRows int64,
	numRowGroups int,
	columns []ColumnSummary,
) Summary {
	var compressed int64
	var uncompressed int64

	codecSet := map[string]struct{}{}
	codecs := make([]string, 0)

	for _, c := range columns {
		compressed += c.CompressedSize
		uncompressed += c.UncompressedSize

		if c.CompressionCodec != "" {
			if _, ok := codecSet[c.CompressionCodec]; !ok {
				codecSet[c.CompressionCodec] = struct{}{}
				codecs = append(codecs, c.CompressionCodec)
			}
		}
	}

	ratio := 0.0
	if compressed > 0 {
		ratio = float64(uncompressed) / float64(compressed)
	}

	return Summary{
		Format:            format,
		NumRows:           numRows,
		NumColumns:        len(columns),
		NumRowGroups:      numRowGroups,
		CompressedSize:    compressed,
		UncompressedSize:  uncompressed,
		CompressionRatio:  ratio,
		CompressionCodecs: codecs,
		Columns:           columns,
	}
}
