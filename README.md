# DataQuarry

> Understand your data before you move it.

DataQuarry is an open-source CLI project for data engineers, focused on understanding data files and identifying storage inefficiencies using evidence instead of guesswork.

The project is in early development. It includes a Go CLI, a functional Parquet `inspect` command, reusable metadata and diagnostic models, and automated tests. Diagnostic rules and additional inspection capabilities remain under development.

## Project Goals

DataQuarry aims to help data engineers answer questions such as:

- What does this data file contain? 
- What schema and storage layout does it use?
- Are there potential storage inefficiencies?
- Which storage format or configuration might suit a workload?

The long-term goal is to support evidence-based storage decisions, starting with Parquet.

## Current Status

Implemented foundations:

- Go CLI built with Cobra
- CLI version reporting
- Reusable Parquet footer metadata reader
- Metadata extraction for row counts, row groups, schema names, and compression codecs
- Structured diagnostic findings with severity validation
- Automated tests and CI

In progress:

- Adding reusable Parquet test fixtures
- Implementing diagnostic rules
- Improving CLI error handling

The `inspect` command reads Parquet footer metadata and reports row counts, column counts, row groups, compressed and uncompressed column-chunk sizes, compression ratio, codecs, and schema. The `diagnose` workflow remains under development.

## Requirements

- Go version specified in [`go.mod`](go.mod)
- Git

## Getting Started

Clone the repository:

```bash
git clone https://github.com/AnuragRaut08/dataquarry.git
cd dataquarry
```

Build the project:

```bash
go build ./...
```

Run the test suite:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

## CLI Usage

Run the root help:

```bash
go run . --help
```

Display the current version:

```bash
go run . --version
```

### Inspect

The `inspect` command reports structural metadata from a Parquet file.

```bash
go run . inspect sales.parquet
```

Replace `sales.parquet` with the path to your Parquet file. The command reports metadata read from the file footer and does not modify the input file.

### Diagnose

The `diagnose` command is intended to report structured findings about potential storage issues.

```bash
go run . diagnose
```

The command currently prints placeholder output. Diagnostic rules and actionable findings are still under development.

## Development

Format Go code before submitting changes:

```bash
gofmt -w ./cmd ./internal
```

Run tests and build checks:

```bash
go test ./...
go build ./...
go vet ./...
```

Please keep changes focused and add tests for new behavior.

## Roadmap

Planned development includes:

- **Diagnostics:** identify evidence-backed storage issues
- **Error handling:** provide clear errors for missing, invalid, or corrupted input files
- **Documentation:** provide accurate usage examples as workflows become available
- **Future exploration:** format comparison, conversion, benchmarking, and optimization

Features in the roadmap should not be considered available until implemented and documented.

## Why Go?

DataQuarry is written in Go for its fast startup, straightforward deployment, cross-platform support, and mature CLI ecosystem.

## Contributing

Contributions are welcome. Before opening a pull request:

- Keep pull requests focused.
- Add or update tests when behavior changes.
- Run the relevant build and test checks.
- Update documentation to reflect actual behavior.

## License

A license has not yet been specified in the repository. Check back for licensing details before reusing or distributing the project.
