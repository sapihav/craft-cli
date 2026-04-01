package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com")

	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %v, want https://api.example.com", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestClient_GetDocuments(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents" {
			t.Errorf("Expected path /documents, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		response := models.DocumentList{
			Items: []models.Document{
				{
					ID:        "doc1",
					Title:     "Test Document",
					CreatedAt: time.Now(),
					LastModifiedAt: time.Now(),
				},
			},
			Total: 1,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	docs, err := client.GetDocuments()

	if err != nil {
		t.Fatalf("GetDocuments() error = %v", err)
	}

	if len(docs.Items) != 1 {
		t.Errorf("Expected 1 document, got %d", len(docs.Items))
	}

	if docs.Items[0].Title != "Test Document" {
		t.Errorf("Document title = %v, want Test Document", docs.Items[0].Title)
	}
}

func TestClient_GetDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blocks" {
			t.Errorf("Expected path /blocks, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("id") != "doc1" {
			t.Errorf("Expected query param id=doc1, got %s", r.URL.Query().Get("id"))
		}

		// Return BlocksResponse format (matching the real Craft API)
		response := models.BlocksResponse{
			ID:        "doc1",
			Type:      "page",
			TextStyle: "page",
			Markdown:  "Test Document",
			Content: []models.Block{
				{
					ID:       "block1",
					Type:     "text",
					Markdown: "Test content paragraph 1",
				},
				{
					ID:       "block2",
					Type:     "text",
					Markdown: "Test content paragraph 2",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	doc, err := client.GetDocument("doc1")

	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}

	if doc.ID != "doc1" {
		t.Errorf("Document ID = %v, want doc1", doc.ID)
	}

	if doc.Title != "Test Document" {
		t.Errorf("Document title = %v, want Test Document", doc.Title)
	}

	// Verify markdown is combined from all blocks
	if doc.Markdown == "" {
		t.Error("Document markdown should not be empty")
	}
}

func TestClient_SearchDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/search" {
			t.Errorf("Expected path /documents/search, got %s", r.URL.Path)
		}

		// Craft API uses 'include' parameter
		include := r.URL.Query().Get("include")
		if include != "test query" {
			t.Errorf("Expected include 'test query', got %s", include)
		}

		response := models.SearchResult{
			Items: []models.SearchItem{
				{
					DocumentID: "doc1",
					Markdown:   "Test **query** result",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	results, err := client.SearchDocuments("test query")

	if err != nil {
		t.Fatalf("SearchDocuments() error = %v", err)
	}

	if len(results.Items) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results.Items))
	}

	if results.Items[0].DocumentID != "doc1" {
		t.Errorf("Expected DocumentID 'doc1', got %s", results.Items[0].DocumentID)
	}
}

func TestClient_CreateDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// First call: POST /documents to create the document
		if r.URL.Path == "/documents" {
			// Craft API expects {"documents": [...]} wrapper
			var wrapper struct {
				Documents []models.CreateDocumentRequest `json:"documents"`
			}
			if err := json.NewDecoder(r.Body).Decode(&wrapper); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if len(wrapper.Documents) != 1 {
				t.Errorf("Expected 1 document, got %d", len(wrapper.Documents))
			}

			if wrapper.Documents[0].Title != "New Document" {
				t.Errorf("Expected title 'New Document', got %s", wrapper.Documents[0].Title)
			}

			// Return the Craft API response format
			response := struct {
				Items []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"items"`
			}{
				Items: []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				}{
					{ID: "doc1", Title: wrapper.Documents[0].Title},
				},
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		// Second call: POST /blocks to insert markdown content
		if r.URL.Path == "/blocks" {
			var req struct {
				Markdown string `json:"markdown"`
				Position struct {
					PageID   string `json:"pageId"`
					Position string `json:"position"`
				} `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if req.Position.PageID != "doc1" {
				t.Errorf("Expected pageId 'doc1', got %s", req.Position.PageID)
			}

			response := struct {
				Items []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Markdown string `json:"markdown"`
				} `json:"items"`
			}{
				Items: []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Markdown string `json:"markdown"`
				}{
					{ID: "block1", Type: "text", Markdown: req.Markdown},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		t.Errorf("Unexpected path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &models.CreateDocumentRequest{
		Title:    "New Document",
		Markdown: "Test content",
	}

	doc, err := client.CreateDocument(req)

	if err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}

	if doc.Title != "New Document" {
		t.Errorf("Document title = %v, want New Document", doc.Title)
	}
}

func TestClient_UpdateDocument(t *testing.T) {
	t.Run("title only update succeeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}
			var body struct {
				Blocks []struct {
					ID       string `json:"id"`
					Markdown string `json:"markdown"`
				} `json:"blocks"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.Blocks) != 1 {
				t.Fatalf("Expected 1 block, got %d", len(body.Blocks))
			}
			if body.Blocks[0].ID != "doc1" {
				t.Errorf("Expected id 'doc1', got %s", body.Blocks[0].ID)
			}
			if body.Blocks[0].Markdown != "Updated Document" {
				t.Errorf("Expected markdown 'Updated Document', got %s", body.Blocks[0].Markdown)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		req := &models.UpdateDocumentRequest{Title: "Updated Document"}
		_, err := client.UpdateDocument("doc1", req)
		if err != nil {
			t.Fatalf("UpdateDocument() error = %v", err)
		}
	})

	t.Run("markdown content update succeeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var req struct {
				Markdown string `json:"markdown"`
				Position struct {
					PageID   string `json:"pageId"`
					Position string `json:"position"`
				} `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if req.Position.PageID != "doc1" {
				t.Errorf("Expected pageId 'doc1', got %s", req.Position.PageID)
			}

			if req.Position.Position != "end" {
				t.Errorf("Expected position 'end', got %s", req.Position.Position)
			}

			response := struct {
				Items []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Markdown string `json:"markdown"`
				} `json:"items"`
			}{
				Items: []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Markdown string `json:"markdown"`
				}{
					{ID: "block1", Type: "text", Markdown: "Updated content"},
				},
			}

			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		req := &models.UpdateDocumentRequest{
			Markdown: "Updated content",
		}

		doc, err := client.UpdateDocument("doc1", req)

		if err != nil {
			t.Fatalf("UpdateDocument() error = %v", err)
		}

		if doc.ID != "doc1" {
			t.Errorf("Document ID = %v, want doc1", doc.ID)
		}
	})
}

func TestClient_DeleteDocument(t *testing.T) {
	t.Run("soft-deletes documents", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			if r.URL.Path != "/documents" {
				t.Errorf("Expected path /documents, got %s", r.URL.Path)
			}
			var body struct {
				DocumentIDs []string `json:"documentIds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.DocumentIDs) != 1 || body.DocumentIDs[0] != "doc1" {
				t.Errorf("Expected documentIds ['doc1'], got %#v", body.DocumentIDs)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"doc1"}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.DeleteDocument("doc1")
		if err != nil {
			t.Fatalf("DeleteDocument() error = %v", err)
		}
	})
}

func TestClient_ClearDocumentContent(t *testing.T) {
	t.Run("deletes content blocks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First call: GET /blocks to fetch document structure
			if r.Method == "GET" && r.URL.Path == "/blocks" {
				response := models.BlocksResponse{
					ID:        "doc1",
					Type:      "page",
					TextStyle: "page",
					Markdown:  "Test Document",
					Content: []models.Block{
						{ID: "block1", Type: "text", Markdown: "Content block 1"},
						{ID: "block2", Type: "text", Markdown: "Content block 2"},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			// Second call: DELETE /blocks with block IDs
			if r.Method == "DELETE" && r.URL.Path == "/blocks" {
				var req struct {
					BlockIDs []string `json:"blockIds"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				if len(req.BlockIDs) != 2 {
					t.Errorf("Expected 2 block IDs, got %d", len(req.BlockIDs))
				}

				response := struct {
					Items []struct {
						ID string `json:"id"`
					} `json:"items"`
				}{
					Items: []struct {
						ID string `json:"id"`
					}{
						{ID: "block1"},
						{ID: "block2"},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		count, err := client.ClearDocumentContent("doc1")
		if err != nil {
			t.Fatalf("ClearDocumentContent() error = %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 deleted blocks, got %d", count)
		}
	})

	t.Run("fails for empty document", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := models.BlocksResponse{ID: "doc1", Type: "page", Markdown: "Empty Document", Content: []models.Block{}}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		count, err := client.ClearDocumentContent("doc1")
		if err != nil {
			t.Fatalf("Expected no error for empty document, got: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected 0 deleted blocks, got %d", count)
		}
	})
}

func TestClient_AddBlocksJSON(t *testing.T) {
	t.Run("sends blocks array with position", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			// Verify blocks array
			blocks, ok := body["blocks"].([]interface{})
			if !ok {
				t.Fatal("Expected 'blocks' array in request body")
			}
			if len(blocks) != 2 {
				t.Errorf("Expected 2 blocks, got %d", len(blocks))
			}

			// Verify first block has styling
			b1 := blocks[0].(map[string]interface{})
			if b1["type"] != "text" {
				t.Errorf("Expected type 'text', got %v", b1["type"])
			}
			if b1["textStyle"] != "h1" {
				t.Errorf("Expected textStyle 'h1', got %v", b1["textStyle"])
			}
			if b1["color"] != "#ef052a" {
				t.Errorf("Expected color '#ef052a', got %v", b1["color"])
			}

			// Verify second block is a line
			b2 := blocks[1].(map[string]interface{})
			if b2["type"] != "line" {
				t.Errorf("Expected type 'line', got %v", b2["type"])
			}
			if b2["lineStyle"] != "strong" {
				t.Errorf("Expected lineStyle 'strong', got %v", b2["lineStyle"])
			}

			// Verify position
			pos, ok := body["position"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected 'position' map in request body")
			}
			if pos["pageId"] != "page1" {
				t.Errorf("Expected pageId 'page1', got %v", pos["pageId"])
			}
			if pos["position"] != "end" {
				t.Errorf("Expected position 'end', got %v", pos["position"])
			}

			response := struct {
				Items []models.Block `json:"items"`
			}{
				Items: []models.Block{
					{ID: "block1", Type: "text", TextStyle: "h1", Color: "#ef052a"},
					{ID: "block2", Type: "line", LineStyle: "strong"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"type": "text", "textStyle": "h1", "color": "#ef052a", "markdown": "# Title"},
			{"type": "line", "lineStyle": "strong"},
		}
		position := map[string]interface{}{
			"pageId":   "page1",
			"position": "end",
		}

		result, err := client.AddBlocksJSON(blocks, position)
		if err != nil {
			t.Fatalf("AddBlocksJSON() error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 blocks returned, got %d", len(result))
		}
		if result[0].ID != "block1" {
			t.Errorf("Expected first block ID 'block1', got %s", result[0].ID)
		}
	})

	t.Run("sends single block with sibling position", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			pos := body["position"].(map[string]interface{})
			if pos["siblingId"] != "sib1" {
				t.Errorf("Expected siblingId 'sib1', got %v", pos["siblingId"])
			}
			if pos["position"] != "before" {
				t.Errorf("Expected position 'before', got %v", pos["position"])
			}

			response := struct {
				Items []models.Block `json:"items"`
			}{Items: []models.Block{{ID: "new1", Type: "text"}}}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{{"type": "text", "markdown": "Hello"}}
		position := map[string]interface{}{"siblingId": "sib1", "position": "before"}

		result, err := client.AddBlocksJSON(blocks, position)
		if err != nil {
			t.Fatalf("AddBlocksJSON() error = %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 block, got %d", len(result))
		}
	})
}

func TestClient_UpdateBlocksJSON(t *testing.T) {
	t.Run("sends blocks array with styling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			blocks, ok := body["blocks"].([]interface{})
			if !ok {
				t.Fatal("Expected 'blocks' array in request body")
			}
			if len(blocks) != 1 {
				t.Errorf("Expected 1 block, got %d", len(blocks))
			}

			b := blocks[0].(map[string]interface{})
			if b["id"] != "block1" {
				t.Errorf("Expected id 'block1', got %v", b["id"])
			}
			if b["color"] != "#0400ff" {
				t.Errorf("Expected color '#0400ff', got %v", b["color"])
			}
			if b["font"] != "serif" {
				t.Errorf("Expected font 'serif', got %v", b["font"])
			}
			if b["textStyle"] != "h2" {
				t.Errorf("Expected textStyle 'h2', got %v", b["textStyle"])
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"id": "block1", "color": "#0400ff", "font": "serif", "textStyle": "h2"},
		}

		err := client.UpdateBlocksJSON(blocks)
		if err != nil {
			t.Fatalf("UpdateBlocksJSON() error = %v", err)
		}
	})

	t.Run("sends multiple blocks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			blocks := body["blocks"].([]interface{})
			if len(blocks) != 2 {
				t.Errorf("Expected 2 blocks, got %d", len(blocks))
			}

			b1 := blocks[0].(map[string]interface{})
			b2 := blocks[1].(map[string]interface{})
			if b1["id"] != "a" {
				t.Errorf("Expected id 'a', got %v", b1["id"])
			}
			if b2["id"] != "b" {
				t.Errorf("Expected id 'b', got %v", b2["id"])
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"id": "a", "markdown": "Updated A"},
			{"id": "b", "color": "#ff0000"},
		}

		err := client.UpdateBlocksJSON(blocks)
		if err != nil {
			t.Fatalf("UpdateBlocksJSON() error = %v", err)
		}
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{"Unauthorized", 401, "authentication required"},
		{"Forbidden", 403, "permission denied"},
		{"NotFound", 404, "resource not found"},
		{"RateLimit", 429, "rate limit exceeded"},
		{"ServerError", 500, "Craft API error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				response := models.ErrorResponse{
					Error:   "error",
					Message: "test error message",
					Code:    tt.statusCode,
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			_, err := client.GetDocuments()

			if err == nil {
				t.Error("Expected error, got nil")
			}

			if err != nil && len(tt.wantErr) > 0 {
				errStr := err.Error()
				if len(errStr) < len(tt.wantErr) || errStr[:len(tt.wantErr)] != tt.wantErr {
					t.Errorf("Error = %v, want to contain %v", err, tt.wantErr)
				}
			}
		})
	}
}
