package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

// ========== GetBlockImage ==========

func TestGetBlockImage_HappyPathBinary(t *testing.T) {
	want := []byte("\x89PNG\r\n\x1a\n-fake-png-bytes-")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/blocks/img-1/image" {
			t.Errorf("Expected /blocks/img-1/image, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "image/png" {
			t.Errorf("Expected Accept=image/png, got %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(want)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	got, ct, err := client.GetBlockImage("img-1", "png")
	if err != nil {
		t.Fatalf("GetBlockImage error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body mismatch: got %q want %q", got, want)
	}
	if ct != "image/png" {
		t.Errorf("Expected content-type image/png, got %q", ct)
	}
}

func TestGetBlockImage_DefaultAcceptStar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "image/*" {
			t.Errorf("Expected default Accept=image/*, got %q", got)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("jpg-bytes"))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, ct, err := client.GetBlockImage("img-2", "")
	if err != nil {
		t.Fatalf("GetBlockImage error: %v", err)
	}
	if ct != "image/jpeg" {
		t.Errorf("Expected image/jpeg, got %q", ct)
	}
}

func TestGetBlockImage_FollowsJSONRedirect(t *testing.T) {
	finalBytes := []byte("redirected-bytes")

	// Asset server (mock CDN)
	asset := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/webp")
		_, _ = w.Write(finalBytes)
	}))
	defer asset.Close()

	// API returns JSON envelope pointing to asset URL
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"assetUrl": asset.URL})
	}))
	defer api.Close()

	client := newTestClient(api.URL)
	got, ct, err := client.GetBlockImage("img-3", "webp")
	if err != nil {
		t.Fatalf("GetBlockImage error: %v", err)
	}
	if string(got) != string(finalBytes) {
		t.Errorf("body mismatch: got %q want %q", got, finalBytes)
	}
	if ct != "image/webp" {
		t.Errorf("Expected image/webp, got %q", ct)
	}
}

func TestGetBlockImage_FollowsURLField(t *testing.T) {
	asset := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("via-url-field"))
	}))
	defer asset.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"url": asset.URL})
	}))
	defer api.Close()

	client := newTestClient(api.URL)
	got, _, err := client.GetBlockImage("img-4", "")
	if err != nil {
		t.Fatalf("GetBlockImage error: %v", err)
	}
	if string(got) != "via-url-field" {
		t.Errorf("body mismatch, got %q", got)
	}
}

func TestGetBlockImage_RejectsEmptyID(t *testing.T) {
	client := newTestClient("https://example.invalid")
	_, _, err := client.GetBlockImage("", "png")
	if err == nil {
		t.Fatal("expected error for empty block ID")
	}
}

func TestGetBlockImage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(models.ErrorResponse{Message: "block not found", Code: 404})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("missing", "png")
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("Expected 'resource not found', got %v", err)
	}
}

func TestGetBlockImage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(models.ErrorResponse{Message: "boom", Code: 500})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("img-5", "")
	if err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestGetBlockImage_MalformedRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"assetUrl":"ftp://insecure.example/x.png"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("img-6", "")
	if err == nil {
		t.Fatal("expected error for scheme-mismatch redirect")
	}
	if !strings.Contains(err.Error(), "scheme mismatch") {
		t.Errorf("expected scheme-mismatch error, got %v", err)
	}
}

func TestGetBlockImage_RedirectInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not-json`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("img-bad-json", "")
	if err == nil {
		t.Fatal("expected error for invalid JSON envelope")
	}
}

func TestGetBlockImage_RedirectMissingURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":"bar"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("img-7", "")
	if err == nil {
		t.Fatal("expected error for envelope missing url field")
	}
}

func TestGetBlockImage_BodyTooLarge(t *testing.T) {
	huge := make([]byte, maxResponseBytes+10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(huge)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, _, err := client.GetBlockImage("img-8", "")
	if err == nil {
		t.Fatal("expected size-limit error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("expected size-limit message, got %v", err)
	}
}

func TestGetBlockImage_TransportError(t *testing.T) {
	// Point at a closed server so the request fails at transport level.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	client := newTestClient(srv.URL)
	_, _, err := client.GetBlockImage("img-9", "")
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestGetBlockImage_SendsAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret-key" {
			t.Errorf("Expected Bearer secret-key, got %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "secret-key", httpClient: newHTTPClient()}
	_, _, err := client.GetBlockImage("img-auth", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ========== GetTask ==========

func TestGetTask_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/tasks/task-1" {
			t.Errorf("Expected /tasks/task-1, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.Task{
			ID:           "task-1",
			BlockID:      "block-1",
			DocumentID:   "doc-1",
			Markdown:     "Buy milk",
			State:        "todo",
			ScheduleDate: "2026-05-01",
		})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	got, err := client.GetTask("task-1")
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if got.ID != "task-1" {
		t.Errorf("Expected task-1, got %s", got.ID)
	}
	if got.Markdown != "Buy milk" {
		t.Errorf("Expected 'Buy milk', got %s", got.Markdown)
	}
	if got.ScheduleDate != "2026-05-01" {
		t.Errorf("Expected schedule 2026-05-01, got %s", got.ScheduleDate)
	}
}

func TestGetTask_RejectsEmptyID(t *testing.T) {
	client := newTestClient("https://example.invalid")
	_, err := client.GetTask("")
	if err == nil {
		t.Fatal("expected error for empty task ID")
	}
}

func TestGetTask_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(models.ErrorResponse{Message: "task not found", Code: 404})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetTask("missing")
	if err == nil {
		t.Fatal("expected 404 error")
	}
	if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("Expected 'resource not found', got %v", err)
	}
}

func TestGetTask_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetTask("task-x")
	if err == nil {
		t.Fatal("expected JSON-decode error")
	}
}

func TestGetTask_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, `{"message":"boom"}`)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetTask("task-x")
	if err == nil {
		t.Fatal("expected 500 error")
	}
}
