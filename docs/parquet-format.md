

````markdown
# Parquet Format & Footer Parsing

This document explains the subset of the Apache Parquet format that DataQuarry reads to inspect and diagnose Parquet files.

DataQuarry intentionally focuses on **metadata**, not row-level data. The information required for structural inspection and diagnostics is already stored in the Parquet file footer, allowing DataQuarry to analyze files without scanning the complete dataset.

---

## 1. Parquet File Layout

A Parquet file is organized into three main areas:

```text
┌──────────────────────────────────────────────┐
│              Column Chunks                   │
│                                              │
│          Encoded / Compressed Data          │
│                                              │
├──────────────────────────────────────────────┤
│              File Metadata                   │
│                                              │
│  Schema • Row Groups • Column Metadata      │
│  Compression • Statistics • Other Metadata  │
│                                              │
├──────────────────────────────────────────────┤
│           Footer Length (4 bytes)            │
├──────────────────────────────────────────────┤
│                PAR1 Magic                    │
└──────────────────────────────────────────────┘
````

The footer is located at the end of the file and contains the structural metadata needed to understand how the data is organized.

DataQuarry reads this metadata rather than scanning the actual rows.

---

## 2. Magic Bytes

Parquet files use the ASCII sequence:

```text
PAR1
```

as their magic bytes.

The file begins with:

```text
PAR1
```

and also ends with:

```text
PAR1
```

The trailing magic bytes allow a reader to verify that the file follows the expected Parquet container structure.

DataQuarry validates the file magic before attempting to parse the footer.

Conceptually:

```text
File start
    ↓
[PAR1]
    ↓
[Column data ...]
    ↓
[File metadata]
    ↓
[Footer length]
    ↓
[PAR1]
    ↓
File end
```

---

## 3. Locating the Footer

The final eight bytes of a Parquet file contain:

```text
4 bytes   → footer length
4 bytes   → PAR1 magic
```

The footer length is stored as a 32-bit little-endian integer.

To locate the metadata:

1. Open the file.
2. Read the final 8 bytes.
3. Validate the final 4 bytes as `PAR1`.
4. Decode the preceding 4 bytes as the footer length.
5. Move backwards from the end of the file by the footer length.
6. Read the footer metadata bytes.

Conceptually:

```text
file size
    │
    ├── 4 bytes: PAR1
    │
    ├── 4 bytes: footer length
    │
    ├── footer metadata
    │
    └── column data
```

This means the parser can locate the metadata without reading the complete file.

---

## 4. Footer Metadata

The footer contains a serialized `FileMetaData` structure.

At a high level, it provides information such as:

```text
FileMetaData
├── version
├── schema
├── num_rows
├── row_groups
├── key_value_metadata
└── created_by
```

DataQuarry primarily uses:

* `schema`
* `num_rows`
* `row_groups`
* column metadata inside each row group

These structures provide the measurements required by the inspection and diagnostic layers.

---

## 5. Thrift Compact Protocol

Parquet serializes its footer metadata using Apache Thrift.

DataQuarry needs to decode the **Thrift Compact Protocol** representation of the footer.

The parser does not attempt to implement the complete Thrift specification.

Instead, it implements the subset required to read the Parquet structures used by DataQuarry.

The decoder must understand concepts such as:

* field identifiers
* field types
* structs
* lists
* maps
* strings
* binary values
* integers
* booleans
* nested structures

The parsing flow is approximately:

```text
Parquet file
     │
     ▼
Locate footer
     │
     ▼
Read serialized metadata
     │
     ▼
Decode Thrift Compact Protocol
     │
     ▼
FileMetaData
     │
     ├── Schema
     ├── Row Groups
     └── Column Metadata
```

Keeping the decoder focused on the required subset keeps DataQuarry lightweight and avoids pulling a large dependency tree into the CLI.

---

## 6. Schema Metadata

The Parquet schema describes the logical structure of the dataset.

Schema metadata can contain information such as:

* column names
* physical types
* logical types
* repetition information
* nesting relationships

For example:

```text
Schema
├── order_id
├── customer_id
├── amount
├── region
└── created_at
```

DataQuarry uses schema metadata to determine structural properties such as the number of columns and their names.

The schema is metadata only; DataQuarry does not need to decode the actual values stored in the column chunks to perform these measurements.

---

## 7. Row Groups

A Parquet file is divided into one or more row groups.

A row group contains a horizontal partition of the dataset.

Conceptually:

```text
Dataset
│
├── Row Group 1
│   ├── Column Chunk A
│   ├── Column Chunk B
│   └── Column Chunk C
│
├── Row Group 2
│   ├── Column Chunk A
│   ├── Column Chunk B
│   └── Column Chunk C
│
└── Row Group 3
    ├── Column Chunk A
    ├── Column Chunk B
    └── Column Chunk C
```

Each row group contains metadata describing:

* number of rows
* total compressed size
* total uncompressed size
* column chunks

DataQuarry uses this information to calculate measurements such as:

* number of row groups
* average row group size
* compressed dataset size
* uncompressed dataset size

These measurements later feed the diagnostic engine.

---

## 8. Column Chunks

Each column inside a row group is represented by a column chunk.

A column chunk contains metadata describing the encoded and compressed representation of that column.

Important metadata includes:

* physical type
* compression codec
* compressed size
* uncompressed size
* encoding information
* statistics when available

Conceptually:

```text
Row Group
│
├── Column Chunk
│   ├── Compression
│   ├── Encoding
│   ├── Compressed Size
│   └── Uncompressed Size
│
├── Column Chunk
│   └── ...
│
└── Column Chunk
    └── ...
```

DataQuarry aggregates this information across row groups and columns to produce dataset-level measurements.

---

## 9. Compression Codecs

Parquet supports multiple compression codecs.

Examples include:

```text
UNCOMPRESSED
SNAPPY
GZIP
LZO
BROTLI
LZ4
ZSTD
```

The codec used by a column chunk is stored in its metadata.

DataQuarry reads this information directly rather than trying to infer the codec from file size or filename.

This allows the inspection layer to report codec usage such as:

```text
Compression Codecs

SNAPPY
ZSTD
```

It also allows the diagnostic layer to identify potentially problematic configurations.

For example, a diagnostic rule may flag uncompressed column chunks or identify codec choices that may warrant review.

---

## 10. What DataQuarry Extracts

The parser exposes structured metadata to higher-level packages.

The information currently relevant to DataQuarry includes:

| Metadata             | Purpose                      |
| -------------------- | ---------------------------- |
| Number of rows       | Dataset size measurement     |
| Schema               | Column structure             |
| Number of columns    | Dataset dimensions           |
| Number of row groups | File organization            |
| Row group sizes      | Storage layout analysis      |
| Compressed size      | Storage footprint            |
| Uncompressed size    | Original data size           |
| Compression codec    | Compression analysis         |
| Column metadata      | Detailed structural analysis |

The parser should return this information through Go structures rather than printing directly to the terminal.

This separation allows multiple features to consume the same metadata.

```text
                 ┌──────────────┐
                 │ Parquet File │
                 └──────┬───────┘
                        │
                        ▼
              ┌───────────────────┐
              │ Footer Parser     │
              │                   │
              │ Thrift decoding   │
              │ FileMetaData      │
              └─────────┬─────────┘
                        │
                        ▼
              ┌───────────────────┐
              │ Metadata Model    │
              └─────────┬─────────┘
                        │
              ┌─────────┼──────────┐
              ▼         ▼          ▼
           inspect   diagnose   optimize
```

---

## 11. Footer-Only Analysis

One of DataQuarry's core design decisions is to perform structural analysis using footer metadata whenever possible.

The parser does **not** need to scan every row in the dataset to determine:

* row count
* schema
* row group count
* compression codecs
* compressed size
* uncompressed size

These values are already represented in the Parquet metadata.

This provides two important benefits:

### Performance

Large datasets can be inspected without reading their complete contents.

```text
Traditional full scan:

File
 ↓
Read rows
 ↓
Decode data
 ↓
Calculate statistics
 ↓
Produce result

DataQuarry:

File
 ↓
Read footer
 ↓
Decode metadata
 ↓
Produce result
```

### Predictability

The measurements come from the file's own metadata rather than from a sampled subset of rows.

DataQuarry therefore avoids presenting sampled estimates as exact structural measurements.

---

## 12. What DataQuarry Intentionally Does Not Parse

DataQuarry does not attempt to implement the complete Parquet specification.

The parser intentionally avoids responsibilities that belong to other layers or future features.

Currently excluded from the footer parser:

* decoding individual row values
* scanning complete datasets
* decompressing column data
* executing queries
* applying diagnostic rules
* generating recommendations
* rewriting Parquet files
* benchmarking storage engines
* supporting non-Parquet formats

The architectural boundary is:

```text
┌───────────────────────────────────────────────┐
│                  CLI Layer                    │
│ inspect / diagnose / optimize / benchmark     │
└───────────────────────┬───────────────────────┘
                        │
                        ▼
┌───────────────────────────────────────────────┐
│             Measurement Layer                 │
│ Summary / derived dataset measurements        │
└───────────────────────┬───────────────────────┘
                        │
                        ▼
┌───────────────────────────────────────────────┐
│             Metadata Layer                    │
│ Parquet footer → structured metadata          │
└───────────────────────┬───────────────────────┘
                        │
                        ▼
┌───────────────────────────────────────────────┐
│                Parquet File                   │
└───────────────────────────────────────────────┘
```

Keeping these responsibilities separate makes the parser reusable and allows future features to build on the same trusted metadata source.

---

## Summary

The important idea behind DataQuarry's Parquet handling is simple:

```text
Parquet File
     │
     ▼
Locate Footer
     │
     ▼
Decode Thrift Compact Protocol
     │
     ▼
Extract FileMetaData
     │
     ├── Schema
     ├── Row Groups
     └── Column Metadata
              │
              ▼
      Derived Measurements
              │
              ├── inspect
              ├── diagnose
              └── future features
```

DataQuarry treats the Parquet footer as the authoritative source for structural measurements.

The parser is therefore deliberately small, focused, and independent of CLI presentation or diagnostic policy.

Future implementations should preserve this boundary: **the metadata layer extracts facts; higher layers decide how those facts are presented, interpreted, or acted upon.**


