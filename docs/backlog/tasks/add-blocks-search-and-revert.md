---
title: M2 — `blocks search` + `blocks revert`
type: task
priority: P1
status: todo
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
