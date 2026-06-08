# Agent and Contributor Rules

This file is the canonical source of repository rules for coding agents and human contributors. It is tool-neutral; tool-specific files should point here instead of duplicating these rules.

## Project Overview

pls7-cli is a Go CLI poker game engine supporting PLS7, PLS, NLH, PLO, and PLO8. It includes AI opponents, save/load functionality, configurable blinds and starting stacks, and YAML-defined rules in `rules/*.yml`.

## Canonical Documentation

- User entry point: [README.md](./README.md)
- Architecture and package boundaries: [ARCHITECTURE.md](./ARCHITECTURE.md)
- Local development workflow: [DEVELOPMENT.md](./DEVELOPMENT.md)
- Active specs: [docs/specs/](./docs/specs/)
- Historical material: [docs/archive/](./docs/archive/)
- Korean supporting references: [docs/ko/](./docs/ko/)

English root-level documents are canonical. Korean documents are supporting references unless manually refreshed against the English source.

## Architecture Rules

The dependency direction is strict:

```text
CLI Layer -> Engine Layer -> Poker Library
```

- `pkg/poker` is the pure poker library and must keep zero internal dependencies.
- `pkg/engine` owns game state, betting flow, AI decisions, save/load, and pot distribution. It may depend on `pkg/poker`.
- `cmd/` and `internal/*` provide application orchestration, terminal UI, configuration loading, and utilities.
- UI code must not leak into `pkg/poker` or `pkg/engine`.
- Game rules belong in `rules/*.yml` when they can be expressed as variant data.

## Development Rules

- TDD is required for code changes: write the test, confirm it fails, implement, confirm it passes, then run the relevant package tests.
- For refactoring, add or confirm coverage for the target behavior before changing implementation.
- Do not delete or modify existing `logrus` statements. Logs are debugging evidence.
- Add `logrus` logs for nested loops or complex state transitions when new complexity is introduced.
- Use English for code comments, documentation, commit messages, PR titles, and PR descriptions.
- Use the test player naming format `[]string{"YOU", "CPU1", "CPU2"}`. Game and test logic depends on this convention.
- Stop on the third failed attempt at the same approach. Report what was tried, what failed, and why.

## Verification

Run the narrowest useful checks during development and the full suite before finishing:

```bash
go build -v ./...
go test ./...
```

CI runs `go build -v ./...` and `go test -v ./...` on Go 1.23.

## Git and Review Rules

- Keep commit messages in English and use conventional commit style.
- Keep PR titles and descriptions in English.
- Prefer small, reviewable changes.
- Do not mix gameplay behavior changes with documentation-only restructuring.
- Do not rewrite historical documents unless the task explicitly asks for it. Move or index historical material so it remains discoverable.

## Documentation Rules

- Keep one canonical home for each topic and link to it elsewhere.
- Keep README focused on users and first-time contributors.
- Keep architecture details in `ARCHITECTURE.md`.
- Keep daily setup, build, test, debug, commit, and PR workflow in `DEVELOPMENT.md`.
- Put active specs in `docs/specs/`.
- Put architecture decision records in `docs/adr/`.
- Put completed plans, old roadmaps, issue notes, and superseded documents in `docs/archive/`.
- Put Korean supporting translations and references in `docs/ko/`.
