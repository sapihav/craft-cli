package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestClient_GetCollections(t *testing.T) {
	t.Run("without filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/collections" {
				t.Errorf("Expected path /collections, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("documentIds") != "" {
				t.Errorf("Expected no documentIds param, got %s", r.URL.Query().Get("documentIds"))
			}

			response := models.CollectionList{
				Items: []models.Collection{
					{ID: "col1", Name: "Tasks DB", ItemCount: 5, DocumentID: "doc1"},
					{ID: "col2", Name: "Projects", ItemCount: 3, DocumentID: "doc2"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollections("")
		if err != nil {
			t.Fatalf("GetCollections() error = %v", err)
		}
		if len(result.Items) != 2 {
			t.Errorf("Expected 2 collections, got %d", len(result.Items))
		}
		if result.Items[0].ID != "col1" {
			t.Errorf("Expected first collection ID 'col1', got %s", result.Items[0].ID)
		}
		if result.Items[0].Name != "Tasks DB" {
			t.Errorf("Expected first collection name 'Tasks DB', got %s", result.Items[0].Name)
		}
		if result.Items[1].ItemCount != 3 {
			t.Errorf("Expected second collection item count 3, got %d", result.Items[1].ItemCount)
		}
	})

	t.Run("with document ID filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/collections" {
				t.Errorf("Expected path /collections, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("documentIds") != "doc1,doc2" {
				t.Errorf("Expected documentIds 'doc1,doc2', got %s", r.URL.Query().Get("documentIds"))
			}

			response := models.CollectionList{
				Items: []models.Collection{
					{ID: "col1", Name: "Tasks DB", ItemCount: 5, DocumentID: "doc1"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollections("doc1,doc2")
		if err != nil {
			t.Fatalf("GetCollections() error = %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 collection, got %d", len(result.Items))
		}
		if result.Items[0].DocumentID != "doc1" {
			t.Errorf("Expected document ID 'doc1', got %s", result.Items[0].DocumentID)
		}
	})
}

func TestClient_GetCollectionSchema(t *testing.T) {
	t.Run("without format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/collections/col1/schema" {
				t.Errorf("Expected path /collections/col1/schema, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("format") != "" {
				t.Errorf("Expected no format param, got %s", r.URL.Query().Get("format"))
			}

			response := models.CollectionSchema{
				Key:  "col1",
				Name: "Tasks DB",
				ContentPropDetails: &models.CollectionPropDetails{
					Key:  "title",
					Name: "Title",
				},
				Properties: []models.CollectionProperty{
					{Key: "status", Name: "Status", Type: "select", Options: []string{"Todo", "Done"}},
					{Key: "priority", Name: "Priority", Type: "number"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollectionSchema("col1", "")
		if err != nil {
			t.Fatalf("GetCollectionSchema() error = %v", err)
		}
		if result.Key != "col1" {
			t.Errorf("Expected schema key 'col1', got %s", result.Key)
		}
		if result.Name != "Tasks DB" {
			t.Errorf("Expected schema name 'Tasks DB', got %s", result.Name)
		}
		if result.ContentPropDetails == nil {
			t.Fatal("Expected content prop details, got nil")
		}
		if result.ContentPropDetails.Key != "title" {
			t.Errorf("Expected content prop key 'title', got %s", result.ContentPropDetails.Key)
		}
		if len(result.Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(result.Properties))
		}
		if result.Properties[0].Type != "select" {
			t.Errorf("Expected first property type 'select', got %s", result.Properties[0].Type)
		}
		if len(result.Properties[0].Options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(result.Properties[0].Options))
		}
	})

	t.Run("with format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/collections/col1/schema" {
				t.Errorf("Expected path /collections/col1/schema, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("format") != "json" {
				t.Errorf("Expected format 'json', got %s", r.URL.Query().Get("format"))
			}

			response := models.CollectionSchema{
				Key:  "col1",
				Name: "Tasks DB",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollectionSchema("col1", "json")
		if err != nil {
			t.Fatalf("GetCollectionSchema() error = %v", err)
		}
		if result.Key != "col1" {
			t.Errorf("Expected schema key 'col1', got %s", result.Key)
		}
	})
}

func TestClient_GetCollectionItems(t *testing.T) {
	t.Run("without maxDepth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/collections/col1/items" {
				t.Errorf("Expected path /collections/col1/items, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("maxDepth") != "" {
				t.Errorf("Expected no maxDepth param, got %s", r.URL.Query().Get("maxDepth"))
			}

			response := models.CollectionItemList{
				Items: []models.CollectionItem{
					{
						ID:    "item1",
						Title: "First Task",
						Properties: map[string]interface{}{
							"status":   "Todo",
							"priority": float64(1),
						},
					},
					{
						ID:    "item2",
						Title: "Second Task",
						Properties: map[string]interface{}{
							"status":   "Done",
							"priority": float64(2),
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollectionItems("col1", 0)
		if err != nil {
			t.Fatalf("GetCollectionItems() error = %v", err)
		}
		if len(result.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result.Items))
		}
		if result.Items[0].ID != "item1" {
			t.Errorf("Expected first item ID 'item1', got %s", result.Items[0].ID)
		}
		if result.Items[0].Title != "First Task" {
			t.Errorf("Expected first item title 'First Task', got %s", result.Items[0].Title)
		}
		if result.Items[0].Properties["status"] != "Todo" {
			t.Errorf("Expected first item status 'Todo', got %v", result.Items[0].Properties["status"])
		}
	})

	t.Run("with maxDepth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/collections/col1/items" {
				t.Errorf("Expected path /collections/col1/items, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("maxDepth") != "3" {
				t.Errorf("Expected maxDepth '3', got %s", r.URL.Query().Get("maxDepth"))
			}

			response := models.CollectionItemList{
				Items: []models.CollectionItem{
					{ID: "item1", Title: "Task with depth"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.GetCollectionItems("col1", 3)
		if err != nil {
			t.Fatalf("GetCollectionItems() error = %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(result.Items))
		}
	})
}

func TestClient_AddCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col1/items" {
			t.Errorf("Expected path /collections/col1/items, got %s", r.URL.Path)
		}

		var body struct {
			Items []struct {
				Title      string                 `json:"title"`
				Properties map[string]interface{} `json:"properties"`
			} `json:"items"`
			AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(body.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(body.Items))
		}
		if body.Items[0].Title != "New Task" {
			t.Errorf("Expected title 'New Task', got %s", body.Items[0].Title)
		}
		if body.Items[0].Properties["status"] != "Todo" {
			t.Errorf("Expected property status 'Todo', got %v", body.Items[0].Properties["status"])
		}
		if !body.AllowNewSelectOptions {
			t.Error("Expected allowNewSelectOptions to be true")
		}

		response := models.CollectionItemList{
			Items: []models.CollectionItem{
				{
					ID:    "item-new",
					Title: "New Task",
					Properties: map[string]interface{}{
						"status": "Todo",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	props := map[string]interface{}{
		"status": "Todo",
	}
	result, err := client.AddCollectionItem("col1", "New Task", props, true)
	if err != nil {
		t.Fatalf("AddCollectionItem() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].ID != "item-new" {
		t.Errorf("Expected item ID 'item-new', got %s", result.Items[0].ID)
	}
	if result.Items[0].Title != "New Task" {
		t.Errorf("Expected item title 'New Task', got %s", result.Items[0].Title)
	}
}

func TestClient_UpdateCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col1/items" {
			t.Errorf("Expected path /collections/col1/items, got %s", r.URL.Path)
		}

		var body struct {
			ItemsToUpdate []struct {
				ID         string                 `json:"id"`
				Properties map[string]interface{} `json:"properties"`
			} `json:"itemsToUpdate"`
			AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(body.ItemsToUpdate) != 1 {
			t.Errorf("Expected 1 item to update, got %d", len(body.ItemsToUpdate))
		}
		if body.ItemsToUpdate[0].ID != "item1" {
			t.Errorf("Expected item ID 'item1', got %s", body.ItemsToUpdate[0].ID)
		}
		if body.ItemsToUpdate[0].Properties["status"] != "Done" {
			t.Errorf("Expected property status 'Done', got %v", body.ItemsToUpdate[0].Properties["status"])
		}
		if body.AllowNewSelectOptions {
			t.Error("Expected allowNewSelectOptions to be false")
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	props := map[string]interface{}{
		"status": "Done",
	}
	err := client.UpdateCollectionItem("col1", "item1", props, false)
	if err != nil {
		t.Fatalf("UpdateCollectionItem() error = %v", err)
	}
}

func TestClient_CreateCollection(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/collections" {
				t.Errorf("Expected path /collections, got %s", r.URL.Path)
			}
			var body struct {
				Collections []models.CreateCollectionRequest `json:"collections"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(body.Collections) != 1 {
				t.Fatalf("expected 1 collection, got %d", len(body.Collections))
			}
			c := body.Collections[0]
			if c.DocumentID != "doc1" || c.Name != "Tasks" || c.Icon != "✅" || c.Description != "all the things" {
				t.Errorf("unexpected request body: %+v", c)
			}
			schemaMap, ok := c.Schema.(map[string]interface{})
			if !ok {
				t.Fatalf("expected schema map, got %T", c.Schema)
			}
			if _, ok := schemaMap["properties"]; !ok {
				t.Errorf("expected schema.properties present")
			}

			resp := map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": "col-new", "name": "Tasks", "documentId": "doc1", "itemCount": 0},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		req := &models.CreateCollectionRequest{
			DocumentID:  "doc1",
			Name:        "Tasks",
			Description: "all the things",
			Icon:        "✅",
			Schema:      map[string]interface{}{"properties": []interface{}{}},
		}
		col, err := client.CreateCollection(req)
		if err != nil {
			t.Fatalf("CreateCollection() error = %v", err)
		}
		if col.ID != "col-new" {
			t.Errorf("expected id 'col-new', got %q", col.ID)
		}
		if col.DocumentID != "doc1" {
			t.Errorf("expected documentId 'doc1', got %q", col.DocumentID)
		}
	})

	t.Run("response missing items falls back to request name+doc", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server returns items[0] with only an ID; client should fill name/docID from request.
			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{{"id": "col-new"}},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		col, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1", Name: "X"})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if col.Name != "X" || col.DocumentID != "doc1" {
			t.Errorf("expected fallback name/doc, got %+v", col)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		client := newTestClient("http://example.test")
		if _, err := client.CreateCollection(nil); err == nil {
			t.Error("expected error for nil request")
		}
	})

	t.Run("missing documentId", func(t *testing.T) {
		client := newTestClient("http://example.test")
		_, err := client.CreateCollection(&models.CreateCollectionRequest{Name: "x"})
		if err == nil {
			t.Error("expected error for missing documentId")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		client := newTestClient("http://example.test")
		_, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1"})
		if err == nil {
			t.Error("expected error for missing name")
		}
	})

	t.Run("API 4xx error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"bad request","message":"invalid schema"}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1", Name: "x"})
		if err == nil {
			t.Fatal("expected API error")
		}
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("expected status 400, got %d", apiErr.StatusCode)
		}
	})

	t.Run("invalid response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1", Name: "x"})
		if err == nil {
			t.Fatal("expected JSON parse error")
		}
	})

	t.Run("empty items in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1", Name: "x"})
		if err == nil {
			t.Fatal("expected error for empty items")
		}
	})

	t.Run("transport error", func(t *testing.T) {
		// Closed server URL → connection refused.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateCollection(&models.CreateCollectionRequest{DocumentID: "doc1", Name: "x"})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func TestClient_UpdateCollectionSchema(t *testing.T) {
	t.Run("echoes updated schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/collections/col1/schema" {
				t.Errorf("Expected path /collections/col1/schema, got %s", r.URL.Path)
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, ok := body["properties"]; !ok {
				t.Errorf("expected properties in body, got %v", body)
			}

			json.NewEncoder(w).Encode(models.CollectionSchema{
				Key:  "col1",
				Name: "Tasks",
				Properties: []models.CollectionProperty{
					{Key: "status", Name: "Status", Type: "select"},
				},
			})
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		schema := map[string]interface{}{
			"properties": []interface{}{
				map[string]interface{}{"key": "status", "name": "Status", "type": "select"},
			},
		}
		result, err := client.UpdateCollectionSchema("col1", schema)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if result == nil || result.Key != "col1" || len(result.Properties) != 1 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("empty body returns nil schema, nil error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.UpdateCollectionSchema("col1", map[string]interface{}{"properties": []interface{}{}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil schema, got %+v", result)
		}
	})

	t.Run("non-schema 2xx body is treated as success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Some Craft endpoints return a non-JSON body (e.g. plain "ok").
			// This must not break the client.
			w.Write([]byte("ok"))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.UpdateCollectionSchema("col1", map[string]interface{}{"properties": []interface{}{}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil schema for non-JSON body, got %+v", result)
		}
	})

	t.Run("missing collection ID", func(t *testing.T) {
		client := newTestClient("http://example.test")
		if _, err := client.UpdateCollectionSchema("", map[string]interface{}{"a": 1}); err == nil {
			t.Error("expected error for empty collection ID")
		}
	})

	t.Run("nil schema", func(t *testing.T) {
		client := newTestClient("http://example.test")
		if _, err := client.UpdateCollectionSchema("col1", nil); err == nil {
			t.Error("expected error for nil schema")
		}
	})

	t.Run("API 4xx error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.UpdateCollectionSchema("col1", map[string]interface{}{"properties": []interface{}{}})
		if err == nil {
			t.Fatal("expected error")
		}
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.StatusCode != 404 {
			t.Errorf("expected 404, got %d", apiErr.StatusCode)
		}
	})

	t.Run("transport error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close()

		client := newTestClient(server.URL)
		_, err := client.UpdateCollectionSchema("col1", map[string]interface{}{"x": 1})
		if err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func TestClient_DeleteCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/collections/col1/items" {
			t.Errorf("Expected path /collections/col1/items, got %s", r.URL.Path)
		}

		var body struct {
			IDsToDelete []string `json:"idsToDelete"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(body.IDsToDelete) != 1 {
			t.Errorf("Expected 1 ID to delete, got %d", len(body.IDsToDelete))
		}
		if body.IDsToDelete[0] != "item1" {
			t.Errorf("Expected ID 'item1', got %s", body.IDsToDelete[0])
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	err := client.DeleteCollectionItem("col1", "item1")
	if err != nil {
		t.Fatalf("DeleteCollectionItem() error = %v", err)
	}
}
