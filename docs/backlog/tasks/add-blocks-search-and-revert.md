---
title: M2 — `blocks search` + `blocks revert`
type: task
priority: P1
status: done
created: 2026-04-18
---

# M2 — `blocks search` + `blocks revert`

## Problem Statement

Two gaps:
1. `internal/api/client.go` already has `SearchBlocks` (in-doc search) but no subcommand wires it up — it's dead code. MCP `document_search` depends on this.
2. MCP `blocks_revert` (undo a block change) has no CLI equivalent at all.

## Acceptance Criteria

- `craft blocks search --doc <doc_id> --query "<text>"` — wires existing `SearchBlocks` client method.
  - Optional: `--limit N`, `--type TEXT|HEADING|CARD|...`.
  - Envelope: `result.blocks[]`.
- `craft blocks revert <block_id>` — new client method + subcommand, hits the revert endpoint.
  - `--dry-run` prints the planned mutation.
  - Returns the post-revert block state.
- `craft schema` reflects both.

## Context / Notes

- Depends on: nothing (M1 not required).
- Budget: ~300 LoC (one new client method, two new cmd files, tests).
- Verify revert endpoint path at implementation time — may be `POST /blocks/{id}/revert` or similar; confirm against `https://connect.craft.do/api-docs`.

## Resolution (2026-04-25, partial → done)

- **`blocks revert`**: shipped. Endpoint `POST /blocks/{id}/revert` (path inferred — mirrors `/collections/{id}/schema` and `/whiteboards/{id}/elements` sub-resource pattern; not in public API docs). New client method `RevertBlock`, new `craft blocks revert <block-id>` subcommand with `--dry-run` and `--quiet`, full ID validation, JSON + text output.
- **`blocks search`**: skipped. `client.SearchBlocks` is already wired via `craft search --document` — the M2 backlog framing it as "dead code" was incorrect (caught during the PARITY audit). A dedicated `blocks search` subcommand would have been a redundant alias delegating to the same function, with no agent affordance gained. Decision documented in `PARITY.md` (Documents row + Gaps section).
- All 4 documented MCP gaps from the parity matrix are now closed.
