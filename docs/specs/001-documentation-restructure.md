# Documentation Restructure

## Status

Draft for PR planning.

## Background

`pls7-cli` currently has useful project documentation, but the information is spread across overlapping files:

- `README.md` mixes user-facing setup, CLI usage, development commands, and documentation links.
- `docs/GEMINI.md` and `docs/GEMINI_en.md` contain agent rules, but they are tool-specific and duplicate each other.
- `docs/architecture.md` and `docs/directory_structure.md` overlap on package responsibilities and project layout.
- Korean translations live beside English source documents without a clear canonical source.
- `docs/development_plan.md`, `docs/roadmap_v20250827.md`, and `docs/issue-history/` mix completed history, roadmap ideas, bug notes, and implementation plans.

The repository targets international contributors and users. English should be the canonical documentation language, with Korean retained as supporting material where useful.

This PR follows the repository bootstrapping guidance from `20260528-repo-bootstrapping-guide-01`: keep one primary home for each kind of information, keep root-level documents focused, and use `AGENTS.md` as the tool-neutral source of agent instructions.

## Goals

- Make English the canonical language for project documentation.
- Add root-level `AGENTS.md` as the tool-neutral source of rules for coding agents and human contributors.
- Keep `README.md` short and user-first: what the project is, quick start, usage, and where to go next.
- Promote architecture documentation to a root-level `ARCHITECTURE.md`.
- Add root-level `DEVELOPMENT.md` for daily contributor commands and workflow.
- Introduce `docs/specs/` for feature and documentation specs.
- Introduce `docs/adr/` for future architecture decision records.
- Introduce `docs/archive/` for historical plans, issue notes, and superseded docs.
- Introduce `docs/ko/` for Korean supporting translations or references.
- Reduce duplicate content by using one canonical home plus links from other documents.
- Preserve project-specific agent rules, including TDD expectations, log handling, player-name test fixtures, and commit/PR language rules.

## Non-goals

- Do not change Go source code or gameplay behavior.
- Do not rewrite the poker rules or user guides linked from external blog posts.
- Do not remove historical issue notes; move or index them so they remain discoverable.
- Do not create a visual design system document because this is a CLI project and has no product UI design system yet.
- Do not add CI enforcement for documentation in this PR unless it is already trivial and low-risk.

## Proposed Documentation Map

### Root documents

| Path | Primary audience | Purpose |
| --- | --- | --- |
| `README.md` | Users, new contributors | Project summary, quick start, CLI usage, documentation map |
| `AGENTS.md` | Agents, contributors | Repository rules, coding workflow, safety rules, verification expectations |
| `ARCHITECTURE.md` | Contributors, agents | System structure, package boundaries, dependency flow, core data model |
| `DEVELOPMENT.md` | Contributors, agents | Local setup, build/test/run commands, debugging, commit and PR workflow |

### `docs/`

| Path | Purpose |
| --- | --- |
| `docs/specs/` | Active feature and documentation specs with goals, non-goals, acceptance criteria, and open questions |
| `docs/adr/` | Architecture decision records for durable design decisions |
| `docs/archive/` | Historical plans, issue notes, and superseded documents |
| `docs/ko/` | Korean supporting translations and references |

## Migration Plan

1. Create root-level canonical documents:
   - `AGENTS.md`
   - `ARCHITECTURE.md`
   - `DEVELOPMENT.md`
2. Rewrite `README.md` as the user-facing entry point:
   - Keep a concise project description.
   - Add a quick start that can be copied and run.
   - Keep CLI usage and examples.
   - Link to the canonical docs.
3. Consolidate architecture information:
   - Use `docs/architecture.md` as the base for `ARCHITECTURE.md`.
   - Fold non-duplicative directory structure details into `ARCHITECTURE.md`.
   - Archive or replace old directory structure documents with links.
4. Consolidate agent rules:
   - Use `docs/GEMINI_en.md` and `docs/GEMINI.md` as source material.
   - Move tool-neutral rules into `AGENTS.md`.
   - Keep Gemini-specific files only as short pointers if needed.
5. Split development history from active workflow:
   - Move daily commands and workflow rules into `DEVELOPMENT.md`.
   - Move completed development plans and old roadmap content under `docs/archive/`.
6. Organize Korean supporting material:
   - Move Korean architecture and directory structure references under `docs/ko/`.
   - Make it clear that English root documents are canonical.
7. Add ADR/spec scaffolding:
   - Keep this spec in `docs/specs/`.
   - Add an ADR README or template if it helps future decisions.
8. Verify documentation links and project commands:
   - Check markdown links for moved files.
   - Run Go tests to ensure the PR did not affect source behavior.

## Acceptance Criteria

- `README.md` is shorter, user-first, and links to the new canonical documentation map.
- `AGENTS.md` exists at the repository root and contains the canonical agent/contributor rules.
- `ARCHITECTURE.md` exists at the repository root and includes the package responsibility and dependency map.
- `DEVELOPMENT.md` exists at the repository root and contains local setup, run, build, test, debug, commit, and PR guidance.
- `docs/specs/001-documentation-restructure.md` documents this PR scope.
- `docs/adr/`, `docs/archive/`, and `docs/ko/` exist where needed.
- Old duplicated documents are either moved, archived, or replaced with clear pointers.
- Korean materials are retained as supporting references, not as canonical sources.
- Markdown links in `README.md` and root-level docs resolve to existing files.
- `go test ./...` passes after the documentation-only changes.

## Migration Decisions

- Keep `docs/GEMINI.md` and `docs/GEMINI_en.md` as short compatibility pointer files if Gemini CLI users still expect those paths. They should point to root `AGENTS.md` and should not duplicate the full rule set.
- Move historical issue notes under `docs/archive/issue-history/` and keep them discoverable from an archive index.
- Move old Korean documents under `docs/ko/` as supporting references. They are not canonical unless manually refreshed to match the English root documents.

## Implementation Notes

- Keep document content in English unless the document lives under `docs/ko/`.
- Do not add persona, chat-style wording, or agent-specific humor to repository documentation.
- Prefer links over repeated paragraphs.
- Keep command blocks executable and avoid placeholder repository URLs.
- Preserve all project-specific rules from the existing Gemini documents unless they conflict with the new English-first documentation model.
