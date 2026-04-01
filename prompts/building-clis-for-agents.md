# Building CLIs for Agents

A consolidated guide from multiple sources on making CLIs work well with AI agents.

---

## Eric Zakariasson's 10 Guidelines

*Source: [@ericzakariasson on X](https://x.com/ericzakariasson/status/2036762680401223946) — 1,816 likes, 483K views, March 25, 2026*

> If you've ever watched an agent try to use a CLI, you've seen it get stuck on an interactive prompt it can't answer, or parse a help page with no examples. Most CLIs were built assuming a human is at the keyboard.

### 1. Make it non-interactive
Every input should be passable as a flag. Keep interactive mode as a fallback when flags are missing, not the primary path. If your CLI drops into a prompt mid-execution, an agent is stuck.

### 2. Don't dump all your docs upfront
Let agents discover progressively: `mycli` -> subcommands -> `mycli deploy --help` -> what it needs. No wasted context on commands it won't use.

### 3. Make --help actually useful
Every subcommand gets a `--help`, and every `--help` includes examples. Agents pattern-match off `mycli deploy --env staging --tag v1.2.3` faster than reading descriptions.

### 4. Accept flags and stdin for everything
Agents think in pipelines. They want to chain commands and pipe output between tools. Don't require positional args in weird orders or fall back to interactive prompts.

### 5. Fail fast with actionable errors
If a required flag is missing, don't hang. Error immediately and show the correct invocation. Agents self-correct when you give them something to work with.

### 6. Make commands idempotent
Agents retry constantly. Running the same deploy twice should return "already deployed, no-op", not create a duplicate.

### 7. Add --dry-run for destructive actions
Let agents preview what a deploy or deletion would do before committing. Validate the plan, then run it for real.

### 8. --yes / --force to skip confirmations
Humans get "are you sure?" — agents pass `--yes` to bypass. Make the safe path the default but allow bypassing.

### 9. Predictable command structure
If an agent learns `mycli service list`, it should guess `mycli deploy list` and `mycli config list`. Pick a pattern (resource + verb) and use it everywhere.

### 10. Return data on success
Show the deploy ID and the URL. Return structured data, not just emojis and prose.

---

## Justin Poehnelt's "Rewrite Your CLI for AI Agents"

*Source: [justin.poehnelt.com](https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/)*

The deeper technical framework behind the Agent DX CLI Scale (see `agent-dx-cli-scale.md`). Key additions beyond the tweet guidelines:

### Machine-Readable Output
- JSON should be the default in non-TTY (piped) contexts
- NDJSON streaming for paginated results
- Errors must also be structured JSON, not just stderr prose

### Raw Payload Input
- All mutating commands should accept raw JSON payloads matching the API schema
- The agent should be able to use the API schema as documentation with zero translation loss
- Don't force agents to reverse-engineer flag-to-payload mappings

### Schema Introspection
- Agents need to discover what a CLI accepts at runtime
- `--help --json` or a `describe` command for machine-readable capability discovery
- Live, runtime-resolved schemas that always reflect the current API version

### Context Window Discipline
- `--fields` / field masks to limit response size
- Pagination controls to avoid dumping entire datasets
- CLI should actively protect the agent from token waste

### Input Hardening
- Agents hallucinate — validate for path traversals, embedded query params, double encoding
- The agent is not a trusted operator
- Sandbox output paths to CWD

### Safety Rails
- `--dry-run` for all mutating commands, not just some
- Response sanitization against prompt injection embedded in API data

### Agent Knowledge Packaging
- Ship `AGENTS.md`, `CONTEXT.md`, or structured skill files
- Encode guardrails: "always use --dry-run", "always use --fields"
- Make skills versioned and discoverable

---

## ShipTypes: Contract-First Design

*Source: [shiptypes.com](https://shiptypes.com/) via agent-to-bricks project*

The principle that all public CLI, plugin, and GUI behavior should be **typed, machine-readable, and discoverable** without relying on prose scraping.

### Core Rules
1. One canonical contract defines payloads, outputs, error shapes, and capability metadata
2. CLI, plugin, GUI, and docs are generated from or validated against that contract
3. Agents can discover capabilities without scraping help text
4. CI fails on contract drift
5. Mutating actions have consistent safety rails and predictable structured responses

### Practical Application
- Maintain a `schema.json` or OpenAPI spec as the single source of truth
- Validate CLI behavior against the schema in CI
- Use typed error codes (not just "Error: something went wrong")
- Safety metadata on each action: `readonly`, `destructive`, `idempotent`, `requires_confirmation`, `supports_dry_run`

---

## AgentDX: MCP Server Linting

*Source: [agentdx/agentdx](https://github.com/agentdx/agentdx) — npm package*

"ESLint for MCP servers" — lints tool descriptions, schemas, and naming that make LLMs pick the wrong tool. 30 rules across 4 categories, producing a 0-100 Lint Score.

```bash
npx agentdx lint           # Lint MCP server tool definitions
npx agentdx lint --format json   # Machine-readable output
npx agentdx init [name]    # Scaffold new MCP server
```

### Rule Categories
1. **Description Quality** (10 rules) — exist, right length, action verbs, avoid vague terms, state purpose/limitations
2. **Schema & Parameters** (11 rules) — valid schemas, typed params, enums > booleans, nesting <= 3, param count <= 10
3. **Naming Conventions** (4 rules) — consistent convention, verb-noun, no duplicates, prefix grouping
4. **Provider Compatibility** (4 rules) — OpenAI tool count limits, name length <= 64, valid chars

Relevant for craft-cli's MCP server mode — lint the tool definitions before shipping.

---

## Checklist: Making craft-cli Agent-First

Based on all sources above, these are the high-priority improvements:

- [ ] JSON output is default in non-TTY contexts
- [ ] All errors return structured JSON with error codes and hints
- [ ] `--dry-run` on all mutating commands (create, update, delete, move, clear)
- [ ] `--fields` flag to limit response fields (context window discipline)
- [ ] `--yes` flag to skip confirmations
- [ ] Stdin JSON input for all mutating commands
- [ ] `craft schema` command for runtime introspection
- [ ] Examples in every `--help` output
- [ ] Predictable resource+verb pattern across all commands
- [ ] AGENTS.md with agent-specific guardrails and patterns
- [ ] Input validation for hallucination patterns (path traversals, encoded chars)
- [ ] Idempotent behavior for create/update operations
- [ ] MCP server surface (stdio JSON-RPC)
- [ ] Headless auth (env var for API key, no interactive prompts)
