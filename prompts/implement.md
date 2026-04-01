# Implementation Prompt

Copy-paste this into Claude Code when you want a coding session to implement changes and validate the right build/test surface before handing work back:

---

Implement the requested changes end to end, then run the relevant verification and docs updates for every touched surface. Do these steps:

**Environment and guardrails:**
1. Load repo-local env if present and check `go.mod` is current.
2. Do not release unless explicitly asked. For release work, follow `prompts/release.md`.

**Before editing:**
3. Inspect the touched subsystem first (`cmd/`, `internal/api/`, `internal/config/`, `internal/models/`) and identify the smallest meaningful verification set for the change.
4. Treat the repo's core philosophy as part of the acceptance criteria:
   - Keep public surfaces contract-first and machine-readable (ShipTypes principle)
   - Preserve or improve agent-DX CLI qualities: structured JSON I/O, raw payload input, stable errors, pagination/field selection, and safety rails
   - Do not rewrite structured strings with naive regex or delimiter splits when valid JSON/URL content could be corrupted

**Required baseline checks for product code changes:**
5. Run `go test ./...` when anything under `cmd/`, `internal/`, or root changes.
6. Run `go vet ./...` when Go code changes.
7. Run `go build -o craft-cli .` to verify the binary builds.
8. If CLI commands, flags, or output formats changed, manually verify with `./craft-cli <command> --help` and test against the live API endpoint.

**API integration checks:**
9. If the change affects live API behavior, test against the Craft Connect API:
   - Endpoint: `https://connect.craft.do/links/HHRuPxZZTJ6/api/v1`
   - Auth: None required (token disabled for debugging)
   - API Reference: `https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1`
10. For destructive operations (delete, update, move), use `--dry-run` first if available.
11. Any integration test must leave the workspace reusable — clean up created test documents.

**Docs and prompt sync:**
12. Update `README.md` for any user-facing change (new commands, flags, behavior).
13. Update `docs/llm/` for any LLM-facing changes (output format, styling, new capabilities).
14. If you change the expected build, install, or release workflow, update the prompt docs under `prompts/` and align `AGENTS.md` so future AI sessions inherit the same process.

**Reporting:**
15. Report the code changes, the exact verification you ran, any agent-DX regressions you prevented or improved, and any remaining risks.
