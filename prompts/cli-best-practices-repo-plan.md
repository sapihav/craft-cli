# Private CLI Best Practices Repo — Plan

A private repository consolidating learnings from building agent-friendly CLIs (craft-cli, agent-to-bricks, and the broader ecosystem).

---

## Proposed Repo Name
`cli-best-practices` (private, under nerveband)

## Purpose
Single source of truth for everything learned about building CLIs that work well with AI agents. Combines personal experience from craft-cli and agent-to-bricks with community best practices.

---

## Proposed Structure

```
cli-best-practices/
├── README.md                     # Overview and philosophy
├── AGENTS.md                     # Meta: how agents should use THIS repo
│
├── principles/
│   ├── agent-dx-cli-scale.md     # 7-axis scoring framework (0-21)
│   ├── building-clis-for-agents.md   # Eric Zakariasson + Justin Poehnelt consolidated
│   ├── ship-types.md             # Contract-first design philosophy
│   ├── clig-dev-summary.md       # Key takeaways from clig.dev (3,573 stars)
│   └── non-interactive-first.md  # Why flags > prompts for agents
│
├── patterns/
│   ├── structured-output.md      # JSON default, NDJSON streaming, error shapes
│   ├── raw-payload-input.md      # Stdin JSON, --json flags, API-schema alignment
│   ├── schema-introspection.md   # Runtime discovery, --help --json, describe commands
│   ├── context-window-discipline.md  # --fields, --limit, pagination
│   ├── input-hardening.md        # Hallucination defense, path traversal, encoding
│   ├── safety-rails.md           # --dry-run, --yes, idempotency, response sanitization
│   ├── predictable-structure.md  # resource+verb, consistent subcommands
│   ├── error-taxonomy.md         # Exit codes, error codes, hints, structured errors
│   └── self-update.md            # Version checking, goreleaser, selfupdate libraries
│
├── prompts/
│   ├── implement.md              # Session prompt for implementation work
│   ├── check.md                  # Session prompt for verification
│   └── release.md                # Session prompt for release pipeline
│
├── scorecards/
│   ├── craft-cli.md              # craft-cli scored: 8/21
│   ├── agent-to-bricks.md        # agent-to-bricks scored: 9/21 (from their own audit)
│   ├── gh-cli.md                 # GitHub CLI scored (reference implementation)
│   └── template.md               # Blank scorecard template
│
├── case-studies/
│   ├── craft-cli.md              # Lessons from building craft-cli
│   ├── agent-to-bricks.md        # Lessons from agent-to-bricks
│   ├── mogcli.md                 # Analysis of jaredpalmer/mogcli (agent-friendly M365)
│   └── n8n-cli.md                # Analysis of Gladium-AI/n8n-cli (agent-friendly n8n)
│
├── tools/
│   ├── goreleaser.md             # goreleaser config patterns
│   ├── cobra-patterns.md         # Go Cobra CLI patterns for agent DX
│   ├── go-selfupdate.md          # Self-update libraries comparison
│   └── mcp-integration.md        # Adding MCP server mode to CLIs
│
└── references/
    ├── repos.md                  # Curated list of relevant repos
    ├── articles.md               # Key articles and posts
    └── ecosystem.md              # State of agent-friendly CLI ecosystem
```

---

## Key Repos to Reference

| Repo | Stars | Relevance |
|------|-------|-----------|
| [HKUDS/CLI-Anything](https://github.com/HKUDS/CLI-Anything) | 13,000+ | Auto-generates CLIs from source code for AI agents |
| [cli-guidelines/cli-guidelines](https://github.com/cli-guidelines/cli-guidelines) | 3,573 | Canonical CLI design reference (clig.dev) |
| [larksuite/cli](https://github.com/larksuite/cli) | — | 200+ commands, 19 AI Agent Skills, MIT. Reference implementation |
| [jaredpalmer/mogcli](https://github.com/jaredpalmer/mogcli) | 197 | Agent-friendly CLI for M365 by Jared Palmer |
| [brwse/earl](https://github.com/brwse/earl) | — | AI-safe CLI with template-driven security (HCL ops) |
| [rhysd/go-github-selfupdate](https://github.com/rhysd/go-github-selfupdate) | 641 | Go + GitHub Releases self-update |
| [minio/selfupdate](https://github.com/minio/selfupdate) | 906 | Go self-update library from MinIO |
| [sanbornm/go-selfupdate](https://github.com/sanbornm/go-selfupdate) | 1,684 | Most-starred Go self-update |
| [sindresorhus/update-notifier](https://github.com/sindresorhus/update-notifier) | 1,795 | Gold standard for update notification UX |
| [agarrharr/awesome-cli-apps](https://github.com/agarrharr/awesome-cli-apps) | 19,168 | Curated CLI app catalog |
| [bradAGI/awesome-cli-coding-agents](https://github.com/bradAGI/awesome-cli-coding-agents) | — | Directory of terminal-native AI coding agents |
| [Gladium-AI/n8n-cli](https://github.com/Gladium-AI/n8n-cli) | 7 | Agent-friendly n8n wrapper |
| [zcaceres/builtwith-api](https://github.com/zcaceres/builtwith-api) | 19 | CLI + MCP dual-interface pattern |

## Key Articles

| Source | Author | Topic |
|--------|--------|-------|
| [Building CLIs for agents](https://x.com/ericzakariasson/status/2036762680401223946) | Eric Zakariasson (Cursor) | 10 practical agent-friendly CLI rules |
| [Rewrite Your CLI for AI Agents](https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/) | Justin Poehnelt | Technical framework behind Agent DX scale |
| [Keep the Terminal Relevant: Patterns for Agent CLIs](https://www.infoq.com/articles/ai-agent-cli/) | InfoQ | 5 design patterns for agent-driven CLIs |
| [Designing a CLI Skills Protocol](https://cliwatch.com/blog/designing-a-cli-skills-protocol) | CLIWatch | `skills` subcommand, 33%→50% GPT pass rate |
| [MCP is dead. Long live the CLI](https://ejholmes.github.io/2026/02/28/mcp-is-dead-long-live-the-cli.html) | Eric Holmes | CLI vs MCP argument, ~400 HN points |
| [MCP vs CLI: Benchmarking Cost & Reliability](https://www.scalekit.com/blog/mcp-vs-cli-use) | Scalekit | CLI 10-32x cheaper, 0% failure rate |
| [Writing CLI Tools Agents Want to Use](https://dev.to/uenyioha/writing-cli-tools-that-ai-agents-actually-want-to-use-39no) | Uche Enyioha | Idempotent/declarative commands pattern |
| [10 Design Principles for Delightful CLIs](https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis) | Atlassian | Classic CLI UX principles |
| [BetterCLI.org](https://bettercli.org/) | BetterCLI | CLI design showcase and reference |
| [shiptypes.com](https://shiptypes.com/) | Boris Tane (Cloudflare) | Contract-first design philosophy |
| [clig.dev](https://clig.dev/) | cli-guidelines | Canonical modern CLI design guide |
| [CLI Design Guidelines](https://www.thoughtworks.com/insights/blog/engineering-effectiveness/elevate-developer-experiences-cli-design-guidelines) | Thoughtworks | 8 guidelines for developer CLI UX |
| [skill.md explained](https://www.gitbook.com/blog/skill-md) | GitBook | How to structure product for AI agents |
| [GitLab glab agent-friendly enhancement](https://gitlab.com/gitlab-org/cli/-/work_items/8177) | GitLab | Blueprint: --agent-info, --help --format json |

## Emerging Standards

| Standard | What | Where |
|----------|------|-------|
| `AI_AGENT` env var | Agent runtime detection | [@vercel/detect-agent](https://www.npmjs.com/package/@vercel/detect-agent) |
| `skills` subcommand | JSON capability discovery | CLIWatch, larksuite/cli |
| SKILL.md format | Structured agent knowledge | [Guide](https://serenitiesai.com/articles/agent-skills-guide-2026) |
| AGENTS.md standard | Agent instructions file | [agentsmd/agents.md](https://github.com/agentsmd/agents.md) |
| `--help --format json` | Machine-readable help | GitLab glab #8177 |

---

## Additional Tools

| Tool | Stars | What |
|------|-------|------|
| [agentdx/agentdx](https://github.com/agentdx/agentdx) | — | "ESLint for MCP servers" — lint tool descriptions, 30 rules, 0-100 score |
| [shadawck/awesome-cli-frameworks](https://github.com/shadawck/awesome-cli-frameworks) | 1,158 | CLI framework catalog across languages |
| [charmbracelet/fang](https://github.com/charmbracelet/fang) | 1,874 | CLI starter kit |

## Gap Identified

**No one has created a public "CLI best practices for AI agents" compilation repository.** The closest is `cli-guidelines/cli-guidelines` (clig.dev) which covers traditional CLI design but predates the agent era. Your repo would be the first to consolidate agent-specific patterns.

---

## Next Steps

1. Create the private repo: `gh repo create nerveband/cli-best-practices --private`
2. Populate with content from craft-cli/prompts/ and agent-to-bricks/prompts/
3. Score 2-3 reference CLIs (gh, craft, agent-to-bricks) for the scorecards/
4. Write up case studies from your own projects
5. Add to CLAUDE.md as a reference for future CLI work
