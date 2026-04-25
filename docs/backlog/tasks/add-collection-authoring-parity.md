---
title: M1 — collection authoring parity (create + schema update)
type: task
priority: P1
status: done
created: 2026-04-18
---

# M1 — collection authoring parity

## Problem Statement

Craft MCP exposes `collections_create` and `collectionSchema_update`. The CLI currently supports collections list / schema read / items CRUD, but **cannot create a new collection** and **cannot update a collection's schema**. Agents that build workspaces end-to-end must drop back to the MCP or GUI.

## Acceptance Criteria

- `craft collections create --name "<name>" --doc <doc_id> [--schema @schema.json]` — creates a new collection in a document.
  - Optional `--icon`, `--description`.
  - `--dry-run` prints the planned POST.
- `craft collections schema update <collection_id> --schema @schema.json` — replace collection schema.
  - Validates schema JSON client-side before sending.
  - `--dry-run` prints the planned PUT.
- Both commands emit the standard envelope.
- `internal/api/client.go` gains two methods: `CreateCollection`, `UpdateCollectionSchema`.
- `craft schema` reflects new commands.

## Context / Notes

- Largest single MCP gap — closes 2 of 4 missing tools in one small PR.
- Budget: ~250 LoC.
- No new models needed if we reuse existing `CollectionSchema` struct from read-path.
