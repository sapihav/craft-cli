# Craft Capability Parity Matrix

Last audited: 2026-04-25 (M3 shipped)

This document maps the official Craft MCP server tools (`mcp.craft.do/my/mcp`,
as exposed in this Claude session as `mcp__claude_ai_Craft__*`) to the
shipped `craft` CLI commands, plus a handful of CLI-only conveniences that
don't have an MCP analogue.

## Sources used for this audit

- **MCP tool list** — the 37 deferred tools matching `mcp__claude_ai_Craft__*`
  exposed in this session. Treated as the authoritative MCP tool inventory.
- **CLI command list** — `craft schema` JSON output (run from
  `/Users/vlads/src/clis/craft-cli`).
- **Backlog** — `docs/backlog/_index.md` and `docs/backlog/tasks/M{1..4}*.md`.
- **CLI source** — `cmd/*.go` and `internal/api/client.go` cross-checked when
  the schema name and the MCP name didn't line up obviously.

## REST API column

Craft does publish a REST API (endpoint:
`https://connect.craft.do/links/HHRuPxZZTJ6/api/v1`, docs at
`https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1`, per `AGENTS.md`), but the
MCP tool surface is the practical superset and is what the backlog tracks
parity against. A separate REST column would duplicate the MCP one for every
row, so it's omitted. If a behaviour is REST-only (no MCP wrapper), it's
called out under "CLI-only / REST-only" below.

## Status legend

- `shipped` — command exists and exercises the same endpoint as the MCP tool.
- `planned (M#)` — listed as acceptance criteria in the named milestone.
- `skipped` — deliberately not on the roadmap; reason given.
- `n/a` — not a CLI-shaped capability.
- `?` — uncertain, needs verification.

## Capability matrix

### Blocks

| MCP tool         | CLI command                                   | Status        | Notes |
|------------------|------------------------------------------------|---------------|-------|
| `blocks_add`     | `craft blocks add`                             | shipped       | Full flag surface (text, code, richUrl, image, file, line, page). Also covered by `markdown_add` route via `--markdown`. |
| `blocks_get`     | `craft blocks get`                             | shipped       | `--depth`, `--metadata`, daily-note `--date` supported. |
| `blocks_update`  | `craft blocks update`                          | shipped       | |
| `blocks_move`    | `craft blocks move`                            | shipped       | |
| `blocks_delete`  | `craft blocks delete`                          | shipped       | |
| `blocks_revert`  | `craft blocks revert`                          | shipped       | M2. `POST /blocks/{id}/revert` (path inferred — mirrors `/collections/{id}/schema` and `/whiteboards/{id}/elements` sub-resource pattern; not in public docs). Supports `--dry-run`, `--quiet`. |
| `markdown_add`   | `craft blocks add --markdown` / `--stdin`      | shipped       | Folded into `blocks add` rather than a separate subcommand — convert markdown to blocks server-side. |
| `image_view`     | `craft blocks image`                           | shipped       | M3. `GET /blocks/{id}/image` (path inferred — mirrors `/blocks/{id}/revert`). Binary stdout by default; `--out FILE` writes file + emits JSON envelope. Follows JSON-envelope redirects (`assetUrl`/`url`) one hop. Refuses to write binary to a TTY. |

### Documents

| MCP tool             | CLI command                | Status   | Notes |
|----------------------|----------------------------|----------|-------|
| `documents_list`     | `craft list`               | shipped  | Returns ALL docs by default; use `--limit`/`--folder`. |
| `documents_create`   | `craft create`             | shipped  | |
| `documents_delete`   | `craft delete`             | shipped  | Soft-delete (moves to trash). |
| `documents_move`     | `craft move`               | shipped  | `--folder`/`--location` (unsorted, trash, templates, daily_notes). |
| `documents_search`   | `craft search`             | shipped  | Document-level fuzzy search. API caps at 20 results. |
| `document_search`    | `craft search --document`  | shipped  | Block-level search inside a single doc. Wires `client.SearchBlocks`. M2 originally proposed a dedicated `blocks search` alias; **skipped** — would be a redundant thin wrapper since `craft search --document` already covers the use case. No agent affordance lost. |

### Folders

| MCP tool          | CLI command            | Status  | Notes |
|-------------------|------------------------|---------|-------|
| `folders_list`    | `craft folders list`   | shipped | |
| `folders_create`  | `craft folders create` | shipped | |
| `folders_move`    | `craft folders move`   | shipped | |
| `folders_delete`  | `craft folders delete` | shipped | |

### Collections

| MCP tool                   | CLI command                  | Status        | Notes |
|----------------------------|------------------------------|---------------|-------|
| `collections_list`         | `craft collections list`     | shipped       | `--document` filter supported. |
| `collections_create`       | `craft collections create`         | shipped       | M1. POST /collections. |
| `collectionSchema_get`     | `craft collections schema`         | shipped       | |
| `collectionSchema_update`  | `craft collections schema update`  | shipped       | M1. PUT /collections/{id}/schema. |
| `collectionItems_get`      | `craft collections items`    | shipped       | List form; per-item GET not exposed. ? — verify whether MCP `collectionItems_get` is per-item or list. |
| `collectionItems_add`      | `craft collections add`      | shipped       | |
| `collectionItems_update`   | `craft collections update`   | shipped       | |
| `collectionItems_delete`   | `craft collections delete`   | shipped       | |

### Tasks

| MCP tool         | CLI command            | Status        | Notes |
|------------------|------------------------|---------------|-------|
| `tasks_add`      | `craft tasks add`      | shipped       | |
| `tasks_get`      | `craft tasks get`      | shipped       | M3. `GET /tasks/{id}` (path inferred — REST conventions). JSON envelope `{"result":{"task":{...}}}`. |
| `tasks_update`   | `craft tasks update`   | shipped       | |
| `tasks_delete`   | `craft tasks delete`   | shipped       | |
| (list)           | `craft tasks list`     | CLI-only      | Backed by `GetDocumentTasks`; no direct MCP equivalent — MCP exposes per-id only. |

### Whiteboards

| MCP tool                    | CLI command                  | Status  | Notes |
|-----------------------------|------------------------------|---------|-------|
| `whiteboard_create`         | `craft whiteboards create`   | shipped | Experimental. |
| `whiteboardElements_get`    | `craft whiteboards get`      | shipped | |
| `whiteboardElements_add`    | `craft whiteboards add`      | shipped | Excalidraw-format; element `id` required. |
| `whiteboardElements_update` | `craft whiteboards update`   | shipped | |
| `whiteboardElements_delete` | `craft whiteboards delete`   | shipped | |

### Misc

| MCP tool          | CLI command            | Status  | Notes |
|-------------------|------------------------|---------|-------|
| `comments_add`    | `craft comments add`   | shipped | Experimental. |
| `connection_info` | `craft connection`     | shipped | Also surfaced by `craft info`. |

## CLI-only commands (no MCP analogue)

These exist for ergonomics, agent affordances, or local-app integration. Not
parity gaps — out of MCP scope by design.

| CLI command       | Purpose                                                                 |
|-------------------|-------------------------------------------------------------------------|
| `craft schema`    | JSON manifest of the entire command tree (self-describing contract).    |
| `craft llm`       | Machine-readable command + styling reference for LLMs.                  |
| `craft llm styles`| Styling/formatting reference for Craft documents.                       |
| `craft info`      | API information and scope.                                              |
| `craft limits`    | Known API/CLI limits.                                                   |
| `craft docs`      | Show available docs help.                                               |
| `craft setup`     | Interactive first-time setup.                                           |
| `craft config *`  | Profile management (add/use/list/remove/reset/format).                  |
| `craft completion`| Shell completion scripts.                                               |
| `craft upgrade`   | Self-upgrade.                                                           |
| `craft version`   | Version info.                                                           |
| `craft clear`     | Delete all content blocks in a document. Convenience over loop of `blocks_delete`. |
| `craft update`    | Update a document (title etc). Maps to a documents endpoint not exposed as a discrete MCP tool. ? |
| `craft upload`    | Upload a file into a document (returns asset URL/block). MCP exposes file blocks via `blocks_add` but not a binary upload tool — CLI is more capable here. |
| `craft local *`   | macOS-only deep-link helpers (open, append, today, tomorrow, search, space). Out of scope for MCP. |

## Gaps

### The "4 MCP gaps" the backlog set out to close

Checked against the current `cmd/` tree. As of 2026-04-25, M1, M2, and M3 have
shipped and closed all genuine gaps:

| # | MCP tool                  | Milestone | Open?  |
|---|---------------------------|-----------|--------|
| 1 | `collections_create`      | M1        | closed |
| 2 | `collectionSchema_update` | M1        | closed |
| 3 | `blocks_revert`           | M2        | closed |
| 4 | `tasks_get`               | M3        | closed |
| 5 | `image_view`              | M3        | closed |
| 6 | `blocks search` wiring    | M2        | closed (no work needed) — `SearchBlocks` was already wired via `craft search --document`. The redundant `blocks search` alias proposed in the M2 task file was **skipped**: it would have been a thin wrapper duplicating an existing path with no agent affordance gained. |

All documented gaps are closed.

### Additional gaps not mentioned in the backlog

- **`collectionItems_get`** — per-item GET. CLI has `collections items` (list)
  but no single-item getter. Not in any milestone. ? — verify MCP semantics
  (the tool name is singular; could be either shape).
- **`--format mcp` envelope** — not a capability gap, but a shape gap. M4.

### Out of MCP, present in CLI

`craft upload` (binary file upload) has no MCP wrapper — agents that want to
push a file into a Craft doc must use the CLI or REST directly.

## Maintenance

When adding a CLI subcommand or noticing a new MCP tool, update both the
relevant section above and the `Last audited` date. The backlog idea
(`docs/backlog/_index.md` → "Schema-driven parity test") would automate this
check; until it lands, this doc is the source of truth.
