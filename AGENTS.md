# Agents

## Quick start

craft-cli is a Go binary for managing Craft.do documents. JSON output by default. 30+ commands covering documents, blocks, folders, tasks, collections, whiteboards, comments, uploads, and search.

## Auth

No interactive login. Set credentials via:
- `craft config add <name> <url>` then `craft config use <name>`
- Or per-command: `--api-url URL --api-key KEY`

## Agent guardrails

- Always use `--dry-run` before delete, move, or clear
- Use `--quiet` to suppress status messages (cleaner for parsing)
- Use `--id-only` or `--output-only <field>` to reduce output tokens
- Use `--limit N` to cap list/search results
- Use `--max-depth N` on get to control block nesting depth
- Use `--json-errors` for machine-readable errors with hints and retry guidance
- Use `craft schema` to discover commands programmatically (JSON manifest with safety metadata)
- Use `--yes` to skip any confirmation prompts

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad input, missing flag) |
| 2 | API error (network, server, rate limit) |
| 3 | Config error (no profile, bad config) |

## Common pitfalls

- `craft list` returns ALL documents (300+). Use `--limit` or `--folder` to filter.
- `craft get` returns full markdown. For large docs, use `--max-depth 1` to limit nesting.
- `craft delete` is soft-delete (moves to trash). Restore with `craft move ID --location unsorted`.
- Whiteboards require element IDs on add. The API rejects elements without an `id` field.
- The search API caps at 20 results. Use `--limit` for fewer, but you can't get more than 20.
- IDs with path traversals (`../`), query params (`?`), or control characters are rejected.
- Errors include `(retryable)` or `(not retryable)` in the hint. Only retry on rate limits and server errors.

## Debugging

- **Endpoint:** `https://connect.craft.do/links/HHRuPxZZTJ6/api/v1`
- **Auth:** None required (token disabled for this test endpoint)
- **API docs:** `https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1`
