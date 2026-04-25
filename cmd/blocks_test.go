package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBlocksRevertCmd_RequiresArg(t *testing.T) {
	// cobra.ExactArgs(1) is enforced via Args; verify directly.
	if err := blocksRevertCmd.Args(blocksRevertCmd, []string{}); err == nil {
		t.Fatal("expected error for missing block-id arg")
	}
	if err := blocksRevertCmd.Args(blocksRevertCmd, []string{"a", "b"}); err == nil {
		t.Fatal("expected error for too many args")
	}
	if err := blocksRevertCmd.Args(blocksRevertCmd, []string{"valid-id"}); err != nil {
		t.Fatalf("expected no error for one arg, got %v", err)
	}
}

func TestBlocksRevertCmd_RejectsBadID(t *testing.T) {
	t.Cleanup(func() { dryRun = false })

	cases := []string{"../etc/passwd", "id?with=query", "id/with/slash", ""}
	for _, bad := range cases {
		err := blocksRevertCmd.RunE(blocksRevertCmd, []string{bad})
		if err == nil {
			t.Errorf("expected validation error for %q", bad)
		}
	}
}

func TestBlocksRevertCmd_DryRunJSON(t *testing.T) {
	t.Cleanup(func() {
		dryRun = false
		outputFormat = ""
		quietMode = false
	})

	dryRun = true
	outputFormat = "json"

	out := captureStdout(t, func() {
		_ = blocksRevertCmd.RunE(blocksRevertCmd, []string{"block-1"})
	})

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("dry-run output not JSON: %v\noutput: %s", err, out)
	}
	if got["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", got["dry_run"])
	}
	if got["action"] != "revert block" {
		t.Errorf("expected action 'revert block', got %v", got["action"])
	}
	target, ok := got["target"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing target: %v", got)
	}
	if target["id"] != "block-1" {
		t.Errorf("expected id 'block-1', got %v", target["id"])
	}
	if target["method"] != "POST" {
		t.Errorf("expected method POST, got %v", target["method"])
	}
	if target["path"] != "/blocks/block-1/revert" {
		t.Errorf("expected path /blocks/block-1/revert, got %v", target["path"])
	}
}

func TestBlocksRevertCmd_DryRunText(t *testing.T) {
	t.Cleanup(func() {
		dryRun = false
		outputFormat = ""
	})

	dryRun = true
	outputFormat = "table" // any non-json, non-empty value forces text dry-run

	out := captureStdout(t, func() {
		_ = blocksRevertCmd.RunE(blocksRevertCmd, []string{"block-7"})
	})

	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix, got %q", out)
	}
	if !strings.Contains(out, "revert block") {
		t.Errorf("expected 'revert block' in output, got %q", out)
	}
	if !strings.Contains(out, "block-7") {
		t.Errorf("expected id in output, got %q", out)
	}
}
