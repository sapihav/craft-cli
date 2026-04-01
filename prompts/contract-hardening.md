# Agent-First Contract Hardening

Adapted from [agent-to-bricks](https://github.com/nerveband/agent-to-bricks) contract hardening plan for craft-cli.

---

## Goal

Move craft-cli from "agent-tolerant" toward "agent-first" by making the contract between the CLI, API client, docs, and tests explicit, typed, validated, and enforceable.

## Desired End State

- One canonical contract defines Craft Connect API payloads, outputs, error shapes, and capability metadata
- CLI behavior is validated against the Craft Connect API spec
- Agents can discover capabilities without scraping help text
- Mutating actions have consistent safety rails and predictable structured responses
- CI fails on contract drift between CLI and API

---

## Current Gaps (craft-cli)

### 1. No formal contract
- No `schema.json` or OpenAPI spec
- CLI behavior is defined by Go code only, not a machine-readable contract
- Agent must scrape `--help` to discover capabilities

### 2. Inconsistent structured behavior
- JSON output exists but not all error paths return structured JSON
- Raw payload input partially supported (stdin for some commands)
- Some commands return human-oriented output even with `--format json`

### 3. Missing safety rails
- `--dry-run` exists for some commands but not all mutating ones
- No input hardening for agent hallucination patterns
- No response sanitization against prompt injection in API data

### 4. No runtime introspection
- No `craft schema` command to discover available commands/flags as JSON
- No way for an agent to programmatically learn what the CLI can do

---

## Workstreams

### WS1: Contract Canonicalization
- Create `schema/craft-connect.json` — canonical contract for all supported Craft API operations
- Define JSON shapes for: document CRUD, block CRUD, folder CRUD, task CRUD, collection CRUD, whiteboard CRUD, upload, search, comments
- Define structured error shapes with codes and hints
- Add safety metadata per operation: `readonly`, `destructive`, `idempotent`, `supports_dry_run`

### WS2: CLI Consistency
- Ensure all commands produce structured JSON by default (non-TTY)
- Ensure all error paths return `{"error": {"code": "...", "message": "...", "hint": "..."}}` 
- Add `--fields` flag to all read commands
- Add `--dry-run` to all mutating commands
- Add `--yes` to commands with confirmation prompts
- Accept stdin JSON for all mutating commands

### WS3: Input Hardening
- Validate document/block IDs against hallucination patterns
- Reject path traversals, percent-encoded segments, embedded query params
- Validate JSON input against expected schemas before sending to API

### WS4: Schema Introspection
- Add `craft schema` command that outputs machine-readable command manifest
- Include params, types, required fields, examples for each command
- Validate schema matches actual CLI behavior in CI

### WS5: Agent Knowledge Packaging
- Enhance `AGENTS.md` with agent-specific guardrails
- Create structured skill files for common workflows
- Add examples to every `--help` output
- Ship `prompts/` folder with session prompts for AI coding agents

### WS6: Test Matrix
- Add contract validation tests (CLI output matches schema)
- Add integration tests against live Craft Connect API
- Add input hardening tests (reject bad inputs)
- CI enforces schema validation on every PR

### WS7: Multi-Surface Readiness
- MCP server mode (`craft mcp`) for stdio JSON-RPC invocation
- Headless auth via `CRAFT_API_URL` and `CRAFT_API_KEY` env vars
- Whiteboard commands (new API surface)

---

## Priority Order

1. **WS2** (CLI Consistency) — immediate value, no new infrastructure
2. **WS5** (Agent Knowledge) — already started with prompts/ folder
3. **WS1** (Contract) — foundation for everything else
4. **WS3** (Input Hardening) — safety before new features
5. **WS4** (Schema Introspection) — unlocks agent self-discovery
6. **WS6** (Tests) — ongoing, layer in as workstreams land
7. **WS7** (Multi-Surface) — new capabilities, whiteboards + MCP
