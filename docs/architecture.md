# DataQuarry Architecture

## Overview

DataQuarry is an evidence-based storage diagnostics tool for modern data engineering.

Instead of telling users to "use Parquet" or "switch to ZSTD", DataQuarry first measures the dataset, then produces recommendations backed by those measurements.

Every command follows the same philosophy:

> **Measure first. Diagnose second. Recommend last.**

This architecture keeps the project modular, testable, and easy to extend as new file formats and storage engines are added.

---

# High-Level Architecture

```
                   +-----------------------+
                   |   Parquet File        |
                   +-----------+-----------+
                               |
                               v
                +----------------------------+
                | Metadata Reader            |
                | (Footer Parser)            |
                +------------+---------------+
                             |
                             v
                +----------------------------+
                | Dataset Summary Builder    |
                +------------+---------------+
                             |
              +--------------+--------------+
              |                             |
              v                             v
     +------------------+          +------------------+
     | inspect          |          | diagnose         |
     +------------------+          +------------------+
              |                             |
              +--------------+--------------+
                             |
                             v
                   Terminal / JSON Output
```

---

# Design Principles

## 1. Separation of Responsibilities

Every layer has a single responsibility.

### Metadata Reader

Responsible for reading the Parquet file.

It knows:

- magic bytes
- footer
- schema
- row groups
- compression codecs

It does **not** know:

- CLI
- recommendations
- diagnostics
- formatting

---

### Summary Builder

Responsible for converting raw metadata into meaningful measurements.

Examples:

- number of rows
- number of columns
- compression ratio
- compressed size
- row group count

It does **not** know:

- terminal output
- colors
- formatting
- recommendation rules

---

### Inspect Command

Responsible for presenting measured information.

Examples:

- dataset summary
- schema
- compression codecs
- row groups

It never performs file parsing directly.

---

### Diagnose Command

Responsible for analysing measurements.

Examples:

- row groups too small
- inefficient compression
- poor compression ratio
- missing statistics

Every recommendation must be backed by measured evidence.

---

# Data Flow

```
Parquet File
      │
      ▼
Read footer
      │
      ▼
Extract metadata
      │
      ▼
Build summary
      │
      ▼
Run diagnostics
      │
      ▼
Render output
```

Every layer depends only on the previous layer.

No layer skips ahead.

For example:

- `diagnose` never reads files.
- `inspect` never parses Parquet.
- CLI commands never manipulate raw footer structures.

This keeps the architecture clean and avoids duplicated logic.

---

# Project Structure

```
dataquarry/

cmd/
    Root CLI entrypoint

internal/

    parquet/
        Footer reader
        Metadata parser
        File structures

    summary/
        Dataset summary builder

    diagnose/
        Diagnostic rules

    common/
        Shared models
        Utilities
        Errors

docs/
    Architecture
    Technical notes

examples/
    Example datasets

testdata/
    Sample Parquet files
```

---

# Dependency Graph

```
CLI
 │
 ▼
Commands
 │
 ▼
Summary Layer
 │
 ▼
Metadata Reader
 │
 ▼
Parquet File
```

Dependencies always point downward.

Lower layers never import higher layers.

---

# Why This Design?

Many existing tools expose raw metadata.

For example:

```
Rows: 1000000

Codec:
SNAPPY

Row Groups:
8
```

That tells users **what** exists.

DataQuarry aims to answer something more useful:

```
Rows: 1,000,000

Compression Ratio:
1.3×

Diagnosis

⚠ Compression ratio is significantly lower than expected.

Evidence:
Measured ratio is 1.3×.

Typical analytical datasets compressed with ZSTD
often achieve between 3× and 6×.

Recommendation:
Consider rewriting the dataset using ZSTD compression.
```

The architecture reflects this distinction.

Measurement is separated from interpretation.

---

# Why Not Parse Files Inside Each Command?

Because it would duplicate logic.

Without this architecture:

```
inspect
   ↓
Read footer

diagnose
   ↓
Read footer

benchmark
   ↓
Read footer
```

Every command would reimplement the same functionality.

Instead:

```
Read footer

↓

Metadata

↓

Shared by every command
```

One implementation.

Many consumers.

---

# Future Extensions

The architecture is intentionally designed to support future capabilities without changing existing layers.

Examples include:

- ORC support
- Avro support
- Delta Lake metadata
- Iceberg metadata
- Storage benchmarking
- Data quality analysis
- Warehouse recommendations
- Cloud cost estimation

These features should build on the existing metadata and summary layers rather than introducing new parsing logic.

---

# Long-Term Vision

DataQuarry is not intended to be another file conversion utility.

The long-term goal is to become an engineering assistant for storage optimisation.

Instead of asking:

> "Can you convert JSON to Parquet?"

Users should eventually be able to ask:

- Why is this Parquet file slow?
- Why is compression poor?
- Which format best fits my workload?
- How should I partition this dataset?
- What changes would improve scan performance?
- How can I reduce storage costs?

Every answer should be based on measured evidence rather than generic best practices.

---

# Guiding Philosophy

DataQuarry follows one simple principle:

> **Measure. Explain. Recommend.**

Every recommendation should be traceable to real metadata extracted from the dataset.

No guesses.

No magic scores.

No black-box heuristics.

Only Evidence.
