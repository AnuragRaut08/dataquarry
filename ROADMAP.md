# DataQuarry Roadmap

## v1.0.0 Release Milestone

**Target release date:** Saturday, 7 November 2026  
**Status:** In progress

DataQuarry v1.0.0 aims to deliver a reliable, well-documented CLI for inspecting Parquet files and identifying evidence-backed storage inefficiencies.

This is a target date for contributors to coordinate around, not a guarantee. Release readiness depends on passing the project's quality checks and resolving critical issues.

### Release goals

- [ ] Reliable Parquet metadata inspection.
- [ ] Accurate diagnostics with clear severity and supporting evidence.
- [ ] Useful errors for missing, invalid, and corrupted input files.
- [ ] Automated tests covering normal operation and important edge cases.
- [ ] Clear CLI help, usage examples, and version reporting.
- [ ] Updated README and technical documentation.
- [ ] CI checks passing for the release commit.
- [ ] License and release metadata verified.
- [ ] Final release notes prepared.

### Proposed schedule

| Date | Goal |
|---|---|
| 26–28 October | Audit current implementation and identify release blockers |
| 29–31 October | Improve correctness, diagnostics, and error handling |
| 1–3 November | Finalize CLI behavior and edge-case tests |
| 4–5 November | Complete documentation and CI verification |
| 6 November | Release candidate review and feature freeze |
| 7 November | Target date for v1.0.0 release |

### Scope boundaries

**Included in v1.0.0**
- Parquet metadata inspection.
- Evidence-backed diagnostics.
- Reliable error handling.
- Automated testing and documentation.

**Deferred until after v1.0.0**
- File conversion.
- Benchmarking workflows.
- Automated optimization and rewriting.
- Broader format support unless required to fix a release blocker.

### How contributors can help

- Check existing issues before starting new work.
- Focus contributions on the release goals above.
- Add or update tests when changing behavior.
- Keep pull requests focused and explain how changes were validated.
- Report bugs and release blockers with reproducible examples.
- Avoid introducing unrelated features during the release freeze.

### Release criteria

The release should proceed only when the release goals have been reviewed, required checks pass, critical issues are resolved, and the documentation accurately describes the shipped functionality.

If a critical blocker remains on the target date, prioritize correctness over meeting the date.

### After v1.0.0

Future milestones may cover additional file formats, conversion, benchmarking, and optimization workflows. These will be planned separately after the initial release.

 
