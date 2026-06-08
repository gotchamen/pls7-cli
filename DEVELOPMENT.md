# Development Guide

This document covers daily contributor workflow for pls7-cli.

## Prerequisites

- Go 1.23 or newer
- Git
- Optional: `make`

## Setup

```bash
git clone https://github.com/gotchamen/pls7-cli.git
cd pls7-cli
go mod download
go test ./...
```

## Run

```bash
go run main.go
go run main.go --rule nlh --difficulty easy
go run main.go --rule plo --difficulty hard
go run main.go --dev
```

With `make`:

```bash
make run
make run-dev
make run-nlh
```

## Build

```bash
go build -v ./...
go build -o pls7 main.go
./pls7
```

With `make`:

```bash
make build
```

## Test

```bash
go test ./...
go test -v ./...
go test -v ./pkg/engine/...
go test -v -run TestFuncName ./pkg/engine/
go test -race ./...
```

With `make`:

```bash
make test
make test-v
make test-race
make check
```

`make check` matches the CI build-and-test workflow.

## Debug

Development mode enables verbose logging:

```bash
go run main.go --dev
```

Debug hand scenarios can be listed and selected:

```bash
go run main.go debug-hands
go run main.go --dev --debug-hand <key>
```

Do not delete or weaken existing `logrus` statements. Add structured log context when new complex state transitions or nested loops need debugging visibility.

## Save Files

Runtime save files are JSON files under `saves/` by default:

```bash
go run main.go saves list
go run main.go saves validate <filename>
go run main.go saves delete <filename>
```

Use `--save-dir` when tests or manual runs should not touch the default save directory.

## Code Workflow

1. Start from an up-to-date branch.
2. Write or confirm tests before changing code.
3. Make the smallest useful change.
4. Run narrow package tests.
5. Run `go test ./...` before finishing.
6. Keep commits focused and in English.

For code changes, follow [AGENTS.md](./AGENTS.md). Documentation-only changes do not need new Go tests, but `go test ./...` should still pass before completion.

## Commit and PR Workflow

- Use conventional commit messages in English.
- Keep PR titles and descriptions in English.
- Mention the verification commands that passed.
- Do not include chat persona text, local-only notes, or tool-specific humor in commits, PRs, or repository documentation.
