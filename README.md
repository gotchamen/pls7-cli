# pls7-cli

**A terminal poker game engine for PLS7, PLS, NLH, PLO, and PLO8.**

[![Go](https://github.com/gotchamen/pls7-cli/actions/workflows/go.yml/badge.svg)](https://github.com/gotchamen/pls7-cli/actions/workflows/go.yml)
![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)
[![License](https://img.shields.io/github/license/gotchamen/pls7-cli)](https://github.com/gotchamen/pls7-cli/blob/main/LICENSE)

## Overview

pls7-cli is a Go CLI poker game with AI opponents, save/load support, configurable blinds and chip stacks, and YAML-defined poker variants.

Supported variants:

| Variant | Full Name | Hole Cards | Betting | Hi-Lo |
| --- | --- |:---:| --- |:---:|
| PLS7 | Pot-Limit Sampyeong 7-or-Better | 3 | Pot-Limit | Yes (7) |
| PLS | Pot-Limit Sampyeong | 3 | Pot-Limit | No |
| NLH | No-Limit Texas Hold'em | 2 | No-Limit | No |
| PLO | Pot-Limit Omaha | 4 | Pot-Limit | No |
| PLO8 | Pot-Limit Omaha 8-or-Better | 4 | Pot-Limit | Yes (8) |

PLS7 is a Hi-Lo split poker variant popular in the Korean poker community. Learn more in the [English guide](https://philipjkim.github.io/posts/20250729-pls7-english-guide/) or [Korean guide](https://philipjkim.github.io/posts/20250724-sampyeong-holdem-guide-v1-4/).

## Quick Start

Go 1.23 or newer is required.

```bash
git clone https://github.com/gotchamen/pls7-cli.git
cd pls7-cli
go mod download
go run main.go
```

## Usage

```bash
go run main.go [flags]
go run main.go saves [list|validate|delete] [args]
go run main.go debug-hands
```

Common examples:

```bash
go run main.go
go run main.go --rule nlh --difficulty easy --outs
go run main.go --rule plo --difficulty hard
go run main.go --initial-chips 500000 --big-blind 2000
go run main.go --blind-up 0
go run main.go --dev --debug-hand river_flush_draw
go run main.go --load
go run main.go --load-file save_20260101_120000
```

Important flags:

| Flag | Short | Default | Description |
| --- |:---:| --- | --- |
| `--rule` | `-r` | `pls7` | Game variant: `pls7`, `pls`, `nlh`, `plo`, or `plo8` |
| `--difficulty` | `-d` | `medium` | AI difficulty: `easy`, `medium`, or `hard` |
| `--big-blind` | | `1000` | Big blind amount. Must be even and at least 2. |
| `--blind-up` | | `10` | Hands per blind level. Use `0` to disable blind increases. |
| `--initial-chips` | | `300000` | Starting chips per player |
| `--dev` | | `false` | Development mode with verbose logging |
| `--debug-hand` | | `""` | Select a debug hand key. Requires `--dev`. |
| `--outs` | | `false` | Show outs for the human player |
| `--load` | `-l` | `false` | Load the most recent saved game |
| `--load-file` | | `""` | Load a specific saved game file |
| `--save-dir` | | `saves` | Directory for saved games |

## Save Management

Games are saved as JSON files in `saves/` unless `--save-dir` is changed.

```bash
go run main.go saves list
go run main.go saves validate <filename>
go run main.go saves delete <filename>
```

## Documentation

- [Architecture](./ARCHITECTURE.md)
- [Development Guide](./DEVELOPMENT.md)
- [Agent and Contributor Rules](./AGENTS.md)
- [Documentation Specs](./docs/specs/)
- [Architecture Decision Records](./docs/adr/)
- [Archive](./docs/archive/)
- [Korean Supporting References](./docs/ko/)

## License

See [LICENSE](./LICENSE) for details.
