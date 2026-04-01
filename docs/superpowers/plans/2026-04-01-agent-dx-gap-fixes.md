# Agent DX Gap Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 9 of 12 gaps identified by the Agent CLI Audit (38/50) and Agent DX Scale (10/21), pushing craft-cli to ~45/50 and 15/21 (Agent-ready). Three gaps (--count, idempotent create, partial failure reporting) are deferred to a follow-up plan.

**Architecture:** Changes are scoped to `cmd/` (command behavior), `internal/api/` (input validation), and docs (`AGENTS.md`). No new dependencies. Each task is independent and can be committed separately.

**Tech Stack:** Go 1.22+, Cobra CLI framework, existing test infrastructure.

---

## File map

| File | Changes |
|------|---------|
| `cmd/root.go` | Add `--yes` flag, add `--limit` flag, add input validation helper, enhance `handleError` with hints for all codes |
| `cmd/validate.go` | NEW: input hardening functions (path traversal, control chars, query params) |
| `cmd/validate_test.go` | NEW: tests for input validation |
| `cmd/delete.go` | Structured dry-run output |
| `cmd/create.go` | Structured dry-run output |
| `cmd/move.go` | Structured dry-run output |
| `cmd/clear.go` | Structured dry-run output |
| `cmd/blocks.go` | Structured dry-run output, input validation on IDs |
| `cmd/whiteboards.go` | Structured dry-run output, input validation on IDs |
| `cmd/tasks.go` | Structured dry-run output |
| `cmd/folders.go` | Structured dry-run output |
| `cmd/collections.go` | Structured dry-run output |
| `cmd/comments.go` | Structured dry-run output |
| `cmd/list.go` | Add `--limit` flag |
| `cmd/get.go` | Add `--max-depth` flag |
| `cmd/search.go` | Add `--limit` flag |
| `AGENTS.md` | Add guardrails, pitfalls, exit code table |

---

### Task 1: Structured dry-run output

Audit gap: 4.3 "Dry-run output doesn't describe what would happen"
Agent DX gap: Safety Rails 2/3 (dry-run output is prose, not JSON)

Every dry-run currently prints `[dry-run] Would delete...` as prose. When `--format json` or `--json-errors` is active (or no TTY), dry-run should return structured JSON instead.

**Files:**
- Modify: `cmd/root.go` (add `dryRunJSON` helper)
- Modify: `cmd/delete.go:27-41`
- Modify: `cmd/create.go:82-96`
- Modify: `cmd/move.go:33-37`
- Modify: `cmd/clear.go`
- Modify: `cmd/blocks.go` (all dry-run blocks)
- Modify: `cmd/whiteboards.go` (all dry-run blocks)
- Modify: `cmd/tasks.go` (all dry-run blocks)
- Modify: `cmd/folders.go` (all dry-run blocks)
- Modify: `cmd/collections.go` (all dry-run blocks)
- Modify: `cmd/comments.go` (all dry-run blocks)

- [ ] **Step 1: Add dryRunOutput helper to root.go**

```go
// dryRunOutput prints structured dry-run info.
// If JSON mode, outputs JSON to stdout. Otherwise prints human prose.
func dryRunOutput(action string, target map[string]interface{}) error {
    if getOutputFormat() == "json" || jsonErrors {
        result := map[string]interface{}{
            "dry_run": true,
            "action":  action,
            "target":  target,
        }
        return outputJSON(result)
    }
    // Fall back to prose
    fmt.Printf("[dry-run] Would %s", action)
    if id, ok := target["id"]; ok {
        fmt.Printf(" %v", id)
    }
    if title, ok := target["title"]; ok {
        fmt.Printf(" (%v)", title)
    }
    fmt.Println()
    return nil
}
```

- [ ] **Step 2: Refactor delete.go dry-run**

Replace lines 27-42 in delete.go:
```go
if isDryRun() {
    client, err := getAPIClient()
    if err != nil {
        return err
    }
    doc, err := client.GetDocument(docID)
    if err != nil {
        return fmt.Errorf("document not found: %s", docID)
    }
    return dryRunOutput("delete", map[string]interface{}{
        "id":    doc.ID,
        "title": doc.Title,
        "reversible": true,
    })
}
```

- [ ] **Step 3: Refactor create.go dry-run**

Replace lines 82-96 in create.go with `dryRunOutput("create", ...)`.

- [ ] **Step 4: Refactor all other commands' dry-run blocks**

Apply same pattern to: move.go, clear.go, blocks.go, whiteboards.go, tasks.go, folders.go, collections.go, comments.go. Each command's dry-run should call `dryRunOutput(action, target)`.

- [ ] **Step 5: Build and test**

Run: `go build -o craft-cli . && go test ./...`
Then: `./craft-cli delete abc123 --dry-run --format json`
Expected: `{"action":"delete","dry_run":true,"target":{"id":"...","title":"...","reversible":true}}`

- [ ] **Step 6: Commit**

```
git add cmd/root.go cmd/delete.go cmd/create.go cmd/move.go cmd/clear.go cmd/blocks.go cmd/whiteboards.go cmd/tasks.go cmd/folders.go cmd/collections.go cmd/comments.go
git commit -m "feat: structured JSON dry-run output on all mutating commands"
```

---

### Task 2: Input validation (hardening against hallucinations)

Audit gap: None (not tested directly)
Agent DX gap: Input Hardening 0/3 -> 2/3

**Files:**
- Create: `cmd/validate.go`
- Create: `cmd/validate_test.go`
- Modify: `cmd/get.go` (add validation call)
- Modify: `cmd/delete.go` (add validation call)
- Modify: `cmd/blocks.go` (add validation call)
- Modify: `cmd/whiteboards.go` (add validation call)

- [ ] **Step 1: Write failing tests**

```go
// cmd/validate_test.go
package cmd

import "testing"

func TestValidateID(t *testing.T) {
    tests := []struct {
        input string
        valid bool
    }{
        {"abc123", true},
        {"9773E4A5-A5B0-4817-833B-FE11C4A57679", true},
        {"../../../etc/passwd", false},
        {"doc123?admin=true", false},
        {"doc123#fragment", false},
        {"..%2f..%2fetc%2fpasswd", false},
        {"doc\x00id", false},
        {"doc\nid", false},
        {"", false},
        {"a", true},
    }
    for _, tt := range tests {
        err := validateResourceID(tt.input, "test")
        if tt.valid && err != nil {
            t.Errorf("validateResourceID(%q) = error %v, want nil", tt.input, err)
        }
        if !tt.valid && err == nil {
            t.Errorf("validateResourceID(%q) = nil, want error", tt.input)
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/ -run TestValidateID -v`
Expected: FAIL (function not defined)

- [ ] **Step 3: Implement validate.go**

```go
// cmd/validate.go
package cmd

import (
    "fmt"
    "strings"
    "unicode"
)

// validateResourceID checks an ID for agent hallucination patterns.
func validateResourceID(id, label string) error {
    if id == "" {
        return fmt.Errorf("%s cannot be empty", label)
    }
    // Path traversal
    if strings.Contains(id, "..") {
        return fmt.Errorf("invalid %s: path traversal detected", label)
    }
    // Percent-encoded traversal
    lower := strings.ToLower(id)
    if strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") {
        return fmt.Errorf("invalid %s: encoded path characters detected", label)
    }
    // Embedded query params
    if strings.ContainsAny(id, "?#&=") {
        return fmt.Errorf("invalid %s: query parameters not allowed in IDs", label)
    }
    // Control characters
    for _, r := range id {
        if unicode.IsControl(r) {
            return fmt.Errorf("invalid %s: control characters not allowed", label)
        }
    }
    return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/ -run TestValidateID -v`
Expected: PASS

- [ ] **Step 5: Add validation calls to commands**

In get.go, delete.go, blocks.go, whiteboards.go: add `validateResourceID(id, "document-id")` before API calls.

- [ ] **Step 6: Build and test**

Run: `go build -o craft-cli . && go test ./...`
Then: `./craft-cli get "../../../etc/passwd" 2>&1`
Expected: `Error: invalid document-id: path traversal detected`

- [ ] **Step 7: Commit**

```
git add cmd/validate.go cmd/validate_test.go cmd/get.go cmd/delete.go cmd/blocks.go cmd/whiteboards.go
git commit -m "feat: input hardening against agent hallucination patterns"
```

---

### Task 3: Add --limit flag for context window discipline

Audit gaps: 6.2 "No --limit flag", 6.4 "No --count"
Agent DX gap: Context Window Discipline 1/3 -> 2/3

**Files:**
- Modify: `cmd/list.go` (add --limit flag)
- Modify: `cmd/search.go` (add --limit flag)
- Modify: `cmd/output.go` (add limit helper)

- [ ] **Step 1: Add --limit to list command**

In list.go, add flag:
```go
var listLimit int
// in init():
listCmd.Flags().IntVar(&listLimit, "limit", 0, "Maximum number of documents to return (0 = all)")
```

In RunE, after getting results, add:
```go
if listLimit > 0 && len(result.Items) > listLimit {
    result.Items = result.Items[:listLimit]
    // Keep result.Total as original count so consumers know truncation happened
}
```

- [ ] **Step 2: Add --limit to search command**

Same pattern in search.go.

- [ ] **Step 3: Add --max-depth to get command**

In get.go, add flag:
```go
var getMaxDepth int
// in init():
getCmd.Flags().IntVar(&getMaxDepth, "max-depth", -1, "Maximum block nesting depth (-1 = all)")
```

The get command uses two code paths:
- `client.GetDocumentBlocks(docID)` for structured/craft/rich formats (returns `BlocksResponse`)
- `client.GetDocument(docID)` for legacy formats (returns `Document`)

For --max-depth, modify the API call in `GetDocumentBlocks` to pass the depth parameter. Add a new method `GetDocumentBlocksWithDepth(id string, maxDepth int)` in `internal/api/client.go`:

```go
func (c *Client) GetDocumentBlocksWithDepth(id string, maxDepth int) (models.BlocksResponse, error) {
    params := url.Values{}
    params.Set("id", id)
    if maxDepth != -1 {
        params.Set("maxDepth", strconv.Itoa(maxDepth))
    }
    path := "/blocks?" + params.Encode()
    data, err := c.doRequest("GET", path, nil)
    if err != nil {
        return models.BlocksResponse{}, err
    }
    var blocksResp models.BlocksResponse
    if err := json.Unmarshal(data, &blocksResp); err != nil {
        return models.BlocksResponse{}, fmt.Errorf("invalid response from API: %w", err)
    }
    return blocksResp, nil
}
```

Then in get.go's RunE, replace `client.GetDocumentBlocks(docID)` with `client.GetDocumentBlocksWithDepth(docID, getMaxDepth)`. For legacy format path, use `GetBlockWithOptions(docID, getMaxDepth, false)` and adapt the response.

- [ ] **Step 4: Build and test**

Run: `go build -o craft-cli . && ./craft-cli list --limit 3 --quiet | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['items']))"`
Expected: `3`

- [ ] **Step 5: Commit**

```
git add cmd/list.go cmd/search.go cmd/get.go
git commit -m "feat: add --limit and --max-depth for context window discipline"
```

---

### Task 4: Enhance error handling with hints

Audit gaps: 5.1 "Errors not actionable", 5.4 "No hint field"
Agent DX gap: improves overall error experience

**Files:**
- Modify: `cmd/root.go:194-220` (enhance handleError)
- Modify: `cmd/root.go:265-278` (expand errorHint)

- [ ] **Step 1: Expand errorHint to cover all codes**

```go
func errorHint(code string) string {
    switch code {
    case "PAYLOAD_TOO_LARGE":
        return "Reduce payload size or use chunking (craft update --chunk-bytes 20000)."
    case "PERMISSION_DENIED":
        return "Check link permissions in Craft. Use 'craft info --test-permissions'."
    case "AUTH_ERROR":
        return "Check API key. Use --api-key flag or 'craft config add'."
    case "CONFIG_ERROR":
        return "Run 'craft config list' or 'craft setup' to reconfigure."
    case "NOT_FOUND":
        return "Check the ID is correct. Use 'craft list --id-only' to find valid IDs."
    case "RATE_LIMIT":
        return "Wait and retry. The API limits request frequency."
    case "API_ERROR":
        return "Server error. Retry in a few seconds. If persistent, check Craft status."
    case "USER_ERROR":
        return "Check command usage with --help."
    default:
        return ""
    }
}
```

- [ ] **Step 2: Always include hint in JSON error output**

In handleError, ensure the hint is always present when using --json-errors (already done, just need to cover all codes from step 1).

- [ ] **Step 3: Add hint to non-JSON errors too**

```go
// In handleError, after printing the error:
if !jsonErrors {
    if hint := errorHint(categorizeError(err)); hint != "" {
        fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
    }
}
```

- [ ] **Step 4: Build and test**

Run: `go build -o craft-cli . && ./craft-cli get nonexistent --json-errors 2>&1`
Expected: JSON with `"hint": "Check the ID is correct..."`

- [ ] **Step 5: Commit**

```
git add cmd/root.go
git commit -m "feat: actionable error hints for all error codes"
```

---

### Task 5: Add --yes flag

Audit gap: 4.4 "No --yes flag"

**Files:**
- Modify: `cmd/root.go` (add global --yes flag)

- [ ] **Step 1: Add flag declaration**

```go
var yesFlag bool
// in init():
rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
```

- [ ] **Step 2: Add helper**

```go
func isYes() bool {
    return yesFlag
}
```

- [ ] **Step 3: Build and test**

Run: `go build -o craft-cli . && ./craft-cli --help | grep -q "yes"`
Expected: flag appears in help.

Note: No commands currently have interactive confirmations (the CLI is already non-interactive by design), but the flag establishes the convention for future commands and signals to agents that the CLI is confirmation-free.

- [ ] **Step 4: Commit**

```
git add cmd/root.go
git commit -m "feat: add --yes/-y flag for confirmation skip convention"
```

---

### Task 6: Enhance AGENTS.md with guardrails and pitfalls

Audit gaps: 8.2 "No guardrails", 8.4 "No pitfalls"

**Files:**
- Modify: `AGENTS.md`

- [ ] **Step 1: Rewrite AGENTS.md**

```markdown
# Agents

## Quick start

craft-cli is a Go binary for managing Craft.do documents. JSON output by default.

## Auth

No interactive login. Set credentials via:
- `craft config add <name> <url>` then `craft config use <name>`
- Or per-command: `--api-url URL --api-key KEY`

## Agent guardrails

- Always use `--dry-run` before delete, move, or clear
- Use `--quiet` to suppress status messages (cleaner for parsing)
- Use `--id-only` or `--output-only <field>` to reduce output tokens
- Use `--limit N` to cap list/search results
- Use `--json-errors` for machine-readable errors with hints
- Use `craft schema` to discover commands programmatically

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad input, missing flag) |
| 2 | API error (network, server, rate limit) |
| 3 | Config error (no profile, bad config) |

## Common pitfalls

- `craft list` returns ALL documents (300+). Use `--limit` or `--folder` to filter.
- `craft get` returns full markdown. For large docs, use `--max-depth 1` to limit.
- `craft delete` is soft-delete (moves to trash). Restore with `craft move ID --location unsorted`.
- Whiteboards require element IDs on add. The API rejects elements without an `id` field.
- The search API caps at 20 results. Use `--limit` for fewer, but you can't get more than 20.

## Debugging

- **Endpoint:** `https://connect.craft.do/links/HHRuPxZZTJ6/api/v1`
- **Auth:** None required (token disabled for this test endpoint)
- **API docs:** `https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1`
```

- [ ] **Step 2: Commit**

```
git add AGENTS.md
git commit -m "docs: add agent guardrails, exit codes, and pitfalls to AGENTS.md"
```

---

### Task 7: Add retry guidance to error messages

Audit gaps: 9.3 "No retry guidance"

**Files:**
- Modify: `cmd/root.go` (extend errorHint with retry info)

- [ ] **Step 1: Add retryable flag to error hints**

Update errorHint to append "(retryable)" where appropriate:
```go
case "RATE_LIMIT":
    return "Wait and retry. The API limits request frequency. (retryable)"
case "API_ERROR":
    return "Server error. Retry in a few seconds. (retryable)"
```

For non-retryable errors:
```go
case "AUTH_ERROR":
    return "Check API key. Use --api-key flag or 'craft config add'. (not retryable)"
case "NOT_FOUND":
    return "Check the ID is correct. Use 'craft list --id-only' to find valid IDs. (not retryable)"
```

- [ ] **Step 2: Build and test**

Run: `go build -o craft-cli .`

- [ ] **Step 3: Commit**

```
git add cmd/root.go
git commit -m "feat: add retry guidance to error hints"
```

---

## Score projections

### Agent CLI Audit (50-point scale)

| Category | Before | After | Gained |
|----------|--------|-------|--------|
| Discoverability | 7/7 | 7/7 | 0 |
| Structured output | 6/6 | 6/6 | 0 |
| Input flexibility | 5/5 | 5/5 | 0 |
| Safety rails | 3/6 | 4/6 | +1 (dry-run output structured; --yes is convention-only) |
| Error handling | 3/5 | 5/5 | +2 (actionable errors, hints) |
| Context discipline | 2/5 | 3/5 | +1 (--limit, --max-depth; no --count) |
| Predictability | 4/4 | 4/4 | 0 |
| Agent knowledge | 3/5 | 5/5 | +2 (guardrails, pitfalls) |
| Resilience | 2/4 | 3/4 | +1 (retry guidance) |
| Distribution | 3/3 | 3/3 | 0 |
| **Total** | **38/50** | **45/50** | **+7** |

### Agent DX Scale (21-point scale)

| Axis | Before | After | Notes |
|------|--------|-------|-------|
| Machine-Readable Output | 2 | 2 | (unchanged, NDJSON for 3) |
| Raw Payload Input | 1 | 1 | (unchanged, needs JSON on all mutating for 2+) |
| Schema Introspection | 2 | 2 | (unchanged, live schemas for 3) |
| Context Window Discipline | 1 | 2 | --limit, --max-depth |
| Input Hardening | 0 | 2 | path traversal, control chars, query params |
| Safety Rails | 2 | 2 | structured dry-run improves quality but still 2/3 without response sanitization |
| Agent Knowledge | 2 | 3 | guardrails + pitfalls in AGENTS.md |
| **Total** | **10/21** | **14/21** | **+4 (Agent-ready at 11+)** |

### Remaining gaps for future work (not in this plan)

| Gap | Points | Effort |
|-----|--------|--------|
| NDJSON streaming | +1 (output 2->3) | Medium |
| Raw JSON on create/tasks/folders | +2 (input 1->3) | Medium |
| --count flag | +1 (context) | Low |
| Idempotent create | +1 (safety) | Medium |
| Partial failure reporting | +1 (resilience) | Medium |
