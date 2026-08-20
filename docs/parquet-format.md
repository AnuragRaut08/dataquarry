
# Parquet Format

This document explains the parts of the Apache Parquet file format that are relevant to DataQuarry.

DataQuarry is primarily interested in the metadata stored at the end of a Parquet file. This metadata contains information about the dataset's schema, row groups, column chunks, compression codecs, and physical layout.

Understanding this structure is important because DataQuarry's inspection and diagnostic features are based on measured metadata rather than guesses or sampling.

---

## 1. What is Parquet?

Apache Parquet is a column-oriented storage format designed for analytical workloads.

Unlike row-oriented formats such as CSV or JSON, Parquet stores data in a structure that allows analytical engines to read only the columns and row groups required by a query.

A simplified representation is:

```text
Parquet File
│
├── Row Group 0
│   ├── Column Chunk: column A
│   ├── Column Chunk: column B
│   └── Column Chunk: column C
│
├── Row Group 1
│   ├── Column Chunk: column A
│   ├── Column Chunk: column B
│   └── Column Chunk: column C
│
├── ...
│
├── File Metadata
│
├── Footer Length
│
└── PAR1
