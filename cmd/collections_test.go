package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSchemaInput(t *testing.T) {
	t.Run("inline JSON", func(t *testing.T) {
		v, err := readSchemaInput(`{"properties":[{"key":"x"}]}`)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		m, ok := v.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", v)
		}
		if _, ok := m["properties"]; !ok {
			t.Error("missing properties key")
		}
	})

	t.Run("@file reads file contents", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "schema.json")
		if err := os.WriteFile(path, []byte(`{"properties":[]}`), 0o644); err != nil {
			t.Fatal(err)
		}
		v, err := readSchemaInput("@" + path)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if _, ok := v.(map[string]interface{}); !ok {
			t.Fatalf("expected map, got %T", v)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := readSchemaInput("   ")
		if err == nil {
			t.Error("expected error for empty input")
		}
	})

	t.Run("@ with empty filename", func(t *testing.T) {
		_, err := readSchemaInput("@   ")
		if err == nil {
			t.Error("expected error for empty filename")
		}
	})

	t.Run("@file missing", func(t *testing.T) {
		_, err := readSchemaInput("@/nonexistent/path/to/file.json")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("invalid inline JSON", func(t *testing.T) {
		_, err := readSchemaInput(`{"bad":`)
		if err == nil {
			t.Error("expected JSON parse error")
		}
	})
}

func TestCollectionsCreateCmd_DryRunPayload(t *testing.T) {
	// Reset globals between tests.
	t.Cleanup(func() {
		dryRun = false
		outputFormat = ""
		quietMode = false
		collectionDocumentID = ""
		collectionCreateName = ""
		collectionCreateDesc = ""
		collectionCreateIcon = ""
		collectionSchemaInput = ""
	})

	dryRun = true
	outputFormat = "json"
	collectionDocumentID = "doc-1"
	collectionCreateName = "Tasks"
	collectionCreateDesc = "all tasks"
	collectionCreateIcon = "✅"
	collectionSchemaInput = `{"properties":[]}`

	out := captureStdout(t, func() {
		_ = collectionsCreateCmd.RunE(collectionsCreateCmd, []string{})
	})

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("dry-run output not JSON: %v\noutput: %s", err, out)
	}
	target, ok := got["target"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing target: %v", got)
	}
	if target["method"] != "POST" || target["path"] != "/collections" {
		t.Errorf("unexpected method/path: %+v", target)
	}
	if target["document_id"] != "doc-1" || target["name"] != "Tasks" || target["icon"] != "✅" {
		t.Errorf("unexpected target fields: %+v", target)
	}
	if _, ok := target["schema"].(map[string]interface{}); !ok {
		t.Errorf("expected schema map in target, got %T", target["schema"])
	}
}

func TestCollectionsCreateCmd_RejectsBadDocID(t *testing.T) {
	t.Cleanup(func() {
		collectionDocumentID = ""
		collectionCreateName = ""
		dryRun = false
	})

	collectionDocumentID = "../etc/passwd"
	collectionCreateName = "x"
	err := collectionsCreateCmd.RunE(collectionsCreateCmd, []string{})
	if err == nil {
		t.Fatal("expected validation error for bad document ID")
	}
}

func TestCollectionsCreateCmd_RejectsBadSchemaJSON(t *testing.T) {
	t.Cleanup(func() {
		collectionDocumentID = ""
		collectionCreateName = ""
		collectionSchemaInput = ""
		dryRun = false
	})

	collectionDocumentID = "doc-1"
	collectionCreateName = "x"
	collectionSchemaInput = `{"bad":`
	err := collectionsCreateCmd.RunE(collectionsCreateCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "invalid schema JSON") {
		t.Fatalf("expected schema JSON parse error, got %v", err)
	}
}

func TestCollectionsSchemaUpdateCmd_DryRunPayload(t *testing.T) {
	t.Cleanup(func() {
		dryRun = false
		outputFormat = ""
		collectionSchemaInput = ""
	})

	dryRun = true
	outputFormat = "json"
	collectionSchemaInput = `{"properties":[{"key":"k","name":"K","type":"text"}]}`

	out := captureStdout(t, func() {
		_ = collectionsSchemaUpdateCmd.RunE(collectionsSchemaUpdateCmd, []string{"col-1"})
	})

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("dry-run output not JSON: %v\noutput: %s", err, out)
	}
	target, ok := got["target"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing target: %v", got)
	}
	if target["method"] != "PUT" {
		t.Errorf("expected method PUT, got %v", target["method"])
	}
	if target["path"] != "/collections/col-1/schema" {
		t.Errorf("unexpected path: %v", target["path"])
	}
	if target["collection_id"] != "col-1" {
		t.Errorf("unexpected collection_id: %v", target["collection_id"])
	}
	if _, ok := target["schema"].(map[string]interface{}); !ok {
		t.Errorf("expected schema map, got %T", target["schema"])
	}
}

func TestCollectionsSchemaUpdateCmd_RejectsBadID(t *testing.T) {
	t.Cleanup(func() {
		collectionSchemaInput = ""
		dryRun = false
	})

	collectionSchemaInput = `{"properties":[]}`
	err := collectionsSchemaUpdateCmd.RunE(collectionsSchemaUpdateCmd, []string{"../bad"})
	if err == nil {
		t.Fatal("expected validation error for bad collection ID")
	}
}

func TestCollectionsSchemaUpdateCmd_RejectsBadSchemaJSON(t *testing.T) {
	t.Cleanup(func() {
		collectionSchemaInput = ""
		dryRun = false
	})

	collectionSchemaInput = `not json`
	err := collectionsSchemaUpdateCmd.RunE(collectionsSchemaUpdateCmd, []string{"col-1"})
	if err == nil || !strings.Contains(err.Error(), "invalid schema JSON") {
		t.Fatalf("expected schema JSON parse error, got %v", err)
	}
}
