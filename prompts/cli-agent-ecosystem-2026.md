# CLI + AI Agent Ecosystem — 2026 Landscape

Comprehensive survey of articles, repos, benchmarks, and discussions on building CLIs for AI agents.

---

## Landmark Repos

### CLI-Anything (13,000+ stars)
**https://github.com/HKUDS/CLI-Anything**
Analyzes source code of any software and auto-generates a complete, tested CLI for AI agents. From HKU Data Intelligence Lab. MIT license. Viral launch — 13K+ stars in under a week.

### brwse/earl — AI-safe CLI
**https://github.com/brwse/earl**
Template-driven security model. Operations are HCL files committed to repo. LLM only provides parameter values; URL/auth/method are human-written. Blocks SSRF/private IPs, secrets stay in OS keychain.

### larksuite/cli — 200+ commands, 19 AI Agent Skills
**https://github.com/larksuite/cli**
Official Lark/Feishu CLI with built-in agent skills. Covers Messenger, Docs, Base, Sheets, Calendar, Mail, Tasks, Meetings. MIT license. Reference implementation for CLI + Agent Skills pattern.

### bradAGI/awesome-cli-coding-agents
**https://github.com/bradAGI/awesome-cli-coding-agents**
Curated directory of terminal-native AI coding agents: OpenCode, Gemini CLI, OpenHands, Aider, Goose, Claude Code, Codex, etc.

### clifor.ai — CLI Tools for AI Agents Directory
**https://www.clifor.ai/**
Curated directory with install commands, editorial reviews, and head-to-head MCP comparisons.

---

## The MCP vs CLI Debate (Dominant 2026 Discussion)

### The Numbers
- **CLI is 10-32x cheaper** per interaction ([Scalekit benchmark](https://www.scalekit.com/blog/mcp-vs-cli-use))
- **MCP had 28% failure rate** vs CLI's 0% (ConnectTimeout issues)
- **GitHub MCP server ships ~55,000 tokens** of tool definitions; `gh` CLI needs ~0 schema tokens
- At 10K agent interactions/day, cost difference is **$500-$2K/month**

### Consensus
- **CLI for developer workflows** (inner loop) — speed, token efficiency, zero setup
- **MCP for enterprise/multi-tenant** (outer loop) — structured auth, governance
- **Both is the right answer** — choose per integration, not per system

### Key Articles
| Title | Source | URL |
|-------|--------|-----|
| "MCP is dead. Long live the CLI" | Eric Holmes | https://ejholmes.github.io/2026/02/28/mcp-is-dead-long-live-the-cli.html |
| "MCP vs CLI: Benchmarking Cost & Reliability" | Scalekit | https://www.scalekit.com/blog/mcp-vs-cli-use |
| "Your MCP Server Is Eating Your Context Window" | Apideck | https://www.apideck.com/blog/mcp-server-eating-context-window-cli-alternative |
| "MCP vs CLI for AI-native development" | CircleCI | https://circleci.com/blog/mcp-vs-cli/ |
| "Why CLI Tools Are Beating MCP" | Jannik Reinhard | https://jannikreinhard.com/2026/02/22/why-cli-tools-are-beating-mcp-for-ai-agents/ |
| "CLI is the New MCP for AI Agents" | OneUptime | https://oneuptime.com/blog/post/2026-02-03-cli-is-the-new-mcp/view |
| MCP counterpoint: enterprise still needs it | Charles Chen | https://chrlschn.dev/blog/2026/03/mcp-is-dead-long-live-mcp/ |

---

## Agent-Friendly CLI Design Articles

### Must-Read

**InfoQ — "Keep the Terminal Relevant: Patterns for AI Agent Driven CLIs"**
https://www.infoq.com/articles/ai-agent-cli/
Five patterns: non-interactive flags (`--no-prompt`), env vars (`NO_COLOR=true`), clear precedence (flags > env > config), documented exit codes, stable output formats as API contracts. Recommends MCP for dynamic discovery.

**CLIWatch — "Designing a CLI Skills Protocol for AI Agents"**
https://cliwatch.com/blog/designing-a-cli-skills-protocol
Proposes `skills` subcommand returning structured JSON. Claims GPT-5.2 went from 33% → 50% pass rate. Cut tokens from 13K → 8K. One-liner in AGENTS.md (~40 tokens) eliminates blind guessing.

**Dev.to — "Writing CLI Tools That AI Agents Actually Want to Use"**
https://dev.to/uenyioha/writing-cli-tools-that-ai-agents-actually-want-to-use-39no
Key insight: use idempotent/declarative commands (`ensure`, `apply`, `sync`) over imperative (`create`, `delete`). Practical checklist included.

### Reference Guides

| Title | Source | URL |
|-------|--------|-----|
| BetterCLI.org — CLI Design Guide | BetterCLI | https://bettercli.org/ |
| 10 Design Principles for Delightful CLIs | Atlassian | https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis |
| CLI Design Guidelines | Thoughtworks | https://www.thoughtworks.com/insights/blog/engineering-effectiveness/elevate-developer-experiences-cli-design-guidelines |
| CLI UX: Progress Display Patterns | Evil Martians | https://evilmartians.com/chronicles/cli-ux-best-practices-3-patterns-for-improving-progress-displays |
| Designing the Confluent CLI | Confluent | https://www.confluent.io/blog/how-we-designed-and-architected-the-confluent-cli/ |
| Building Developer CLI Tools in 2026 | Dev.to (Chengyi Xu) | https://dev.to/chengyixu/the-complete-guide-to-building-developer-cli-tools-in-2026-a96 |

---

## Emerging Standards

### Agent Detection: `AI_AGENT` env var
**@vercel/detect-agent** (npm): Detects AI agent runtime (Cursor, Claude Code, Devin, Gemini CLI, Codex, Copilot, Replit). Promotes `AI_AGENT` as universal env var.
https://www.npmjs.com/package/@vercel/detect-agent

**AGENTS.md Issue #136**: Proposal for `CI=true` equivalent for agents.
https://github.com/agentsmd/agents.md/issues/136

**Implication for craft-cli**: Detect `AI_AGENT` and auto-switch to JSON output, suppress progress bars, structured errors.

### CLI Skills Protocol
The `skills` subcommand pattern (CLIWatch + larksuite/cli) is emerging as the agent-discovery standard:
```bash
craft skills                    # List available skills as JSON
craft skills --workflow search  # Get search workflow
```

### GitLab glab Agent Blueprint (Issue #8177)
https://gitlab.com/gitlab-org/cli/-/work_items/8177
Practical roadmap: `--agent-info` flag, consistent exit codes, `--help --format json`, structured error JSON, `glab doctor`. Worth studying for craft-cli.

---

## Agent Skills & Discovery

### SKILL.md Format
- YAML frontmatter (name, description, triggers)
- Optional scripts, references, assets
- Supported by 16+ tools
- 100% pass rate when AGENTS.md points to skills vs 53% without
- Guide: https://serenitiesai.com/articles/agent-skills-guide-2026
- GitBook explainer: https://www.gitbook.com/blog/skill-md

### AGENTS.md Best Practice
- Keep it minimal (~40 tokens for the agent discovery hint)
- Point to skills/schema for details
- Don't dump everything — progressive disclosure

---

## Key Themes for craft-cli

1. **Treat `--json` output as a versioned API contract** — breaking changes break all automation
2. **Detect `AI_AGENT` env var** — auto-switch to structured output
3. **Add `craft skills` or `craft schema`** — agent self-discovery without context waste
4. **Idempotent commands** — `ensure`/`apply` patterns where possible
5. **CLI output to stdout, messages to stderr** — clean piping for agents
6. **Progressive disclosure > schema dump** — `--help` → subcommand → `--help` → examples
7. **MCP mode is complementary, not replacement** — offer both surfaces
8. **Study GitLab glab #8177** — most practical agent-friendly CLI blueprint
9. **Ship SKILL.md files** — structured agent knowledge packaging
10. **Benchmark against MCP** — prove the token/cost advantage with numbers
