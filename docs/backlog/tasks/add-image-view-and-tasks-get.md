---
title: M3 — `blocks image` (image_view) + `tasks get`
type: task
priority: P2
status: todo
created: 2026-04-18
---

# M3 — `blocks image` + `tasks get`

## Problem Statement

- MCP `image_view` fetches binary image content for image blocks. No CLI equivalent — `upload` is insert-only.
- MCP `tasks_get` fetches a single task by id. CLI has `tasks list` (via `GetDocumentTasks`) but no single-task getter subcommand.

## Acceptance Criteria

- `craft blocks image <block_id> [--out <file>]` — fetches image binary.
  - Default: writes bytes to stdout (suitable for piping to `file -`, ImageMagick, etc.).
  - With `--out`: writes to file, prints a minimal JSON envelope `{result:{path, content_type, size_bytes}}` to stdout.
  - `--format` (accept header preference, e.g. `png`, `jpeg`).
- `craft tasks get <task_id>` — GET single task.
  - Envelope: `result.{task: {id, title, status, due_date?, subtasks[]?, ...}}`.

## Context / Notes

- Image endpoint may require `Accept: image/*` handling — verify in API docs.
- Budget: ~200 LoC.
- Depends on: nothing.
