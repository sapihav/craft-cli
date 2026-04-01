# Release Prompt

Copy-paste this into Claude Code after you're done with your changes and ready to release:

---

Run the full pre-release check, release pipeline, and post-release verification. Do these steps:

**Build checks:**
1. Run `go test ./...` for all tests
2. Run `go vet ./...` for static analysis
3. Run `go build -o craft-cli .` to verify the binary builds
4. Verify version in `cmd/root.go` matches the intended release version

**API integration verification:**
5. Test core operations against the live Craft Connect API:
   - `./craft-cli list` — document listing
   - `./craft-cli folders` — folder listing
   - `./craft-cli search "test"` — search functionality
   - `./craft-cli get <doc-id>` — document retrieval
   - `./craft-cli blocks get <doc-id>` — block-level access
   - `./craft-cli collections` — collections listing
   - `./craft-cli tasks` — task listing
6. Verify all output formats: `--format json`, `--format table`, `--format markdown`, `--format compact`

**Dependency check:**
7. Run `go list -m -u all 2>/dev/null | grep '\[' | head -10` to check for Go dependency updates
8. If any critical updates, apply them before releasing

**Docs & content sync:**
9. Check `git log --oneline` since the last release tag to see all changes
10. Update `README.md` for any user-facing changes
11. Update `docs/llm/` for any LLM-facing changes
12. Update `AGENTS.md` if process or architecture changed
13. Confirm the release follows agent-DX principles:
    - Contract-first, machine-readable public surfaces
    - Structured JSON I/O, raw payload input, stable errors
    - No newly introduced structured-string corruption

**Release:**
14. Choose the version bump based on compatibility:
    - patch for internal-only fixes
    - minor for additive public behavior (new commands, flags)
    - major for breaking changes to output format or command structure
15. Update version in `cmd/root.go`
16. Commit all release changes together
17. Tag with `vX.Y.Z`
18. Push main first, then push the tag
19. Run `goreleaser release --clean` or let GitHub Actions handle it
20. Monitor the release until all artifacts are published (Linux, macOS, Windows builds)

**Post-release verification:**
21. Download and install the released binary
22. Run `craft version` to confirm the new version
23. Run `craft info` to verify API connectivity
24. Run `craft list` to verify basic functionality
25. Check `gh release view v<VERSION>` to confirm all assets are present

**If re-releasing (workflow failed):**
- First try: `gh run rerun <run-id> --failed`
- If that doesn't work: delete the release and tag, re-create and push
- After re-release: verify assets don't conflict with previous release
