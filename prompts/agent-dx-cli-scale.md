# Agent DX CLI Scale

> Human DX optimizes for discoverability and forgiveness.
> Agent DX optimizes for predictability and defense-in-depth.
> — [You Need to Rewrite Your CLI for AI Agents](https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/)

Use this scale to evaluate any CLI against agent-first design principles. Score each axis 0-3, sum for a total between 0-21.

---

## Scoring Axes

### 1. Machine-Readable Output

| Score | Criteria |
|-------|----------|
| 0 | Human-only output (tables, color codes, prose). No structured format. |
| 1 | `--output json` exists but is incomplete or inconsistent across commands. |
| 2 | Consistent JSON output across all commands. Errors also return structured JSON. |
| 3 | NDJSON streaming for paginated results. Structured output is the default in non-TTY contexts. |

### 2. Raw Payload Input

| Score | Criteria |
|-------|----------|
| 0 | Only bespoke flags. No way to pass structured input. |
| 1 | Accepts `--json` or stdin JSON for some commands, but most require flags. |
| 2 | All mutating commands accept a raw JSON payload that maps to the API schema. |
| 3 | Raw payload is first-class alongside convenience flags. Agent can use API schema as docs with zero translation loss. |

### 3. Schema Introspection

| Score | Criteria |
|-------|----------|
| 0 | Only `--help` text. No machine-readable schema. |
| 1 | `--help --json` or a `describe` command for some surfaces. |
| 2 | Full schema introspection for all commands — params, types, required fields — as JSON. |
| 3 | Live, runtime-resolved schemas that reflect the current API version. Includes scopes, enums, nested types. |

### 4. Context Window Discipline

| Score | Criteria |
|-------|----------|
| 0 | Returns full API responses with no way to limit fields or paginate. |
| 1 | Supports `--fields` or field masks on some commands. |
| 2 | Field masks on all read commands. Pagination with `--page-all` or equivalent. |
| 3 | Streaming pagination (NDJSON per page). Explicit guidance on field mask usage. CLI actively protects agent from token waste. |

### 5. Input Hardening

| Score | Criteria |
|-------|----------|
| 0 | No input validation beyond basic type checks. |
| 1 | Validates some inputs, but not agent-specific hallucination patterns. |
| 2 | Rejects control characters, path traversals, percent-encoded segments, embedded query params in resource IDs. |
| 3 | Comprehensive hardening plus output path sandboxing, HTTP-layer encoding, and explicit security posture. |

### 6. Safety Rails

| Score | Criteria |
|-------|----------|
| 0 | No dry-run mode. No response sanitization. |
| 1 | `--dry-run` exists for some mutating commands. |
| 2 | `--dry-run` for all mutating commands. Agent can validate without side effects. |
| 3 | Dry-run plus response sanitization to defend against prompt injection in API data. |

### 7. Agent Knowledge Packaging

| Score | Criteria |
|-------|----------|
| 0 | Only `--help` and a docs site. No agent-specific context files. |
| 1 | A `CONTEXT.md` or `AGENTS.md` with basic usage guidance. |
| 2 | Structured skill files covering per-command or per-API-surface workflows and invariants. |
| 3 | Comprehensive skill library encoding agent-specific guardrails. Skills are versioned and discoverable. |

---

## Interpreting the Total

| Range | Rating | Description |
|-------|--------|-------------|
| 0-5 | **Human-only** | Built for humans. Agents struggle with parsing, hallucinate inputs, lack safety rails. |
| 6-10 | **Agent-tolerant** | Agents can use it but waste tokens, make avoidable errors, require heavy prompt engineering. |
| 11-15 | **Agent-ready** | Solid agent support. Structured I/O, input validation, some introspection. |
| 16-21 | **Agent-first** | Purpose-built for agents. Full schema introspection, comprehensive hardening, safety rails, packaged knowledge. |

---

## Bonus: Multi-Surface Readiness

- [ ] **MCP (stdio JSON-RPC)** — typed tool invocation, no shell escaping
- [ ] **Extension / plugin install** — agent treats CLI as native capability
- [ ] **Headless auth** — env vars for tokens/credentials, no browser redirect
