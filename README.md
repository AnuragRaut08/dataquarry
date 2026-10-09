# DataQuarry

> Understand your data before you move it.

DataQuarry is an open-source CLI project for data engineers, focused on understanding data files and identifying storage inefficiencies using evidence instead of guesswork.

The project is in early development. The current codebase includes a Go CLI foundation, structured diagnostic models, and a reusable Parquet metadata reader. The end-to-end `inspect` and `diagnose` workflows are still under development.

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

- Connecting Parquet metadata to the `inspect` command
- Adding reusable Parquet test fixtures
- Implementing diagnostic rules
- Improving CLI error handling

The `inspect` and `diagnose` commands are not yet complete end-to-end workflows. Their current output is placeholder text.

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

The `inspect` command is intended to report structural information about a data file, beginning with Parquet metadata.

```bash
go run . inspect
```

At the current development stage, the command prints placeholder output. Passing a file path and displaying real metadata are part of the ongoing implementation.

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

- **Inspection:** display Parquet schema, row counts, row groups, and compression metadata
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
