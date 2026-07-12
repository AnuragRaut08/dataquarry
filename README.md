# DataQuarry

> Understand your data before you move it.

DataQuarry is an open-source CLI for **data engineers** that inspects datasets, diagnoses storage inefficiencies, and recommends better data formats using measurable evidence—not guesswork.

Instead of asking:

> "Should I use Parquet or Avro?"

DataQuarry helps answer:

> "Based on *my actual data*, what storage format and configuration should I use, and why?"

---

## Why DataQuarry?

Every day, data engineers make decisions like:

- Should this dataset stay as JSON?
- Is Parquet actually the right choice?
- Why are my Parquet queries slow?
- Is my compression optimal?
- Are my row groups too small?
- Should I partition this dataset?
- Is this file suitable for analytics or streaming?

Today, these decisions usually rely on experience, blog posts, or trial and error.

**DataQuarry aims to make them evidence-based.**

---

# Vision

Our mission is simple:

> **Help data engineers make better storage decisions using evidence instead of guesswork.**

Rather than becoming "just another converter", DataQuarry aims to become a practical assistant that explains:

- what your data looks like
- why performance suffers
- how storage can be improved
- which format best matches your workload

---

# Features

## Current

- CLI foundation
- Project architecture
- CI pipeline
- Extensible command structure

---

## Planned

### Inspect

```bash
dataquarry inspect sales.parquet
```

Analyze a dataset and display:

- schema
- row count
- column count
- null percentage
- nested fields
- compression
- row groups
- statistics

Example:

```
File: sales.parquet

Rows: 12,487,391
Columns: 18

Compression:
  ZSTD

Row Groups:
  8

Schema:
  id
  customer_id
  amount
  created_at

Recommendation

✓ Good compression
⚠ Row groups are smaller than recommended
```

---

### Diagnose

```bash
dataquarry diagnose sales.parquet
```

Explain performance issues.

Example:

```
Problem:
Small row groups detected.

Impact:
More metadata reads
Poor scan performance

Recommendation:
Rewrite with 128MB row groups.

Estimated improvement:
+34% read performance
```

---

### Compare

```bash
dataquarry compare orders.json
```

Estimate storage across formats.

Example

| Format | Estimated Size | Best For |
|---------|----------------|-----------|
| JSON | 340 MB | APIs |
| Avro | 118 MB | Streaming |
| Parquet | 61 MB | Analytics |
| ORC | 58 MB | Warehouses |

---

### Convert

```bash
dataquarry convert orders.json --to parquet
```

Supported formats (planned)

- JSON
- CSV
- Parquet
- Avro
- ORC

---

### Benchmark

```bash
dataquarry bench orders.parquet
```

Measure

- read speed
- query latency
- scan throughput
- compression efficiency

---

### Optimize

```bash
dataquarry optimize sales.parquet
```

Example

```
Current

Compression:
SNAPPY

Row Groups:
32MB

Recommendation

Compression:
ZSTD

Row Groups:
128MB

Estimated improvement

Storage:
-27%

Read Speed:
+39%
```

---

### Quality

```bash
dataquarry quality orders.parquet
```

Detect

- missing values
- duplicate IDs
- skewed columns
- timestamp anomalies
- outliers
- invalid values

---

# Long-Term Vision

Eventually, DataQuarry will evolve from a storage inspection tool into a data engineering copilot capable of answering questions like:

```
What is the best storage format?

Should I partition this dataset?

Why is this Parquet file slow?

Is DuckDB better than Snowflake for this workload?

What compression codec should I use?

How can I reduce storage cost?

What is the expected performance improvement?
```

The goal is not to replace existing data engines.

The goal is to help engineers make better decisions before data reaches those engines.

---

# Roadmap

## v0.1

- CLI foundation
- Parquet inspection
- Schema extraction
- Metadata reporting

---

## v0.2

- Diagnose command
- Performance recommendations
- Better statistics

---

## v0.3

- Format comparison
- Conversion

---

## v0.4

- Benchmarking

---

## v0.5

- Optimization engine

---

## v1.0

- Data quality
- Recommendation engine
- Rich terminal output

---

# Why Go?

DataQuarry is written in Go because it provides

- fast startup
- single static binaries
- cross-platform support
- simple installation
- excellent CLI ecosystem

---

# Project Structure

```
cmd/
    CLI commands

internal/
    parquet/
    core/
    common/
    cli/

docs/
    architecture
    knowledge

examples/

testdata/
```

---

# Installation

Coming soon.

---

# Development

Clone the repository

```bash
git clone https://github.com/AnuragRaut08/dataquarry.git
```

Run

```bash
go build ./...
```

Tests

```bash
go test ./...
```

---

# Contributing

Contributions are welcome.

Whether you're fixing bugs, improving documentation, adding new storage formats, or suggesting better recommendations, we'd love your help.

Please read:

- CONTRIBUTING.md

before opening a pull request.

---

# Inspiration

DataQuarry is inspired by real-world challenges faced by data engineers working with:

- Apache Arrow
- Apache Parquet
- DuckDB
- DataFusion
- Snowflake
- BigQuery
- Redshift
- Databricks

---

# License

MIT License

---

## Author

**Anurag Raut**:anuragtraut2003@gmail.com

GitHub:https://github.com/AnuragRaut08

---

⭐ If you find this project useful, consider starring the repository.
