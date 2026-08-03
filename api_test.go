package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	a := newAPI(false, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	a.health(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", resp["status"])
	}
}

func TestUserEndpointDesktop(t *testing.T) {
	t.Setenv("USER", "tester")
	a := newAPI(false, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/user", nil)
	rr := httptest.NewRecorder()

	a.user(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp UserInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp.Mode != "desktop" {
		t.Errorf("expected mode 'desktop', got %s", resp.Mode)
	}
	if resp.Username != "tester" {
		t.Errorf("expected username 'tester', got %s", resp.Username)
	}
}

func TestUserEndpointServerProxyHeader(t *testing.T) {
	a := newAPI(true, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/user", nil)
	req.Header.Set("Remote-User", "alice_proxy")
	rr := httptest.NewRecorder()

	a.user(rr, req)

	var resp UserInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp.Username != "alice_proxy" {
		t.Errorf("expected username 'alice_proxy', got %s", resp.Username)
	}
	if resp.Mode != "server" {
		t.Errorf("expected mode 'server', got %s", resp.Mode)
	}
}

func TestUserEndpointServerNoHeader(t *testing.T) {
	a := newAPI(true, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/user", nil)
	rr := httptest.NewRecorder()

	a.user(rr, req)

	var resp UserInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp.Username != "anonymous" {
		t.Errorf("expected username 'anonymous', got %s", resp.Username)
	}
}

func TestItemsCRUD(t *testing.T) {
	t.Setenv("USER", "tester")
	a := newAPI(false, []byte("{}"))

	// Test GET items
	reqGet := httptest.NewRequest("GET", "/api/v1/items", nil)
	rrGet := httptest.NewRecorder()
	a.listItems(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Errorf("GET /items returned %d", rrGet.Code)
	}

	// Test POST item
	payload := []byte(`{"name":"New Test Item"}`)
	reqPost := httptest.NewRequest("POST", "/api/v1/items", bytes.NewBuffer(payload))
	rrPost := httptest.NewRecorder()
	a.createItem(rrPost, reqPost)

	if rrPost.Code != http.StatusCreated {
		t.Errorf("POST /items returned %d", rrPost.Code)
	}

	var created Item
	if err := json.Unmarshal(rrPost.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if created.Name != "New Test Item" {
		t.Errorf("expected item name 'New Test Item', got %s", created.Name)
	}
	if created.CreatedBy != "tester" {
		t.Errorf("expected created_by 'tester', got %s", created.CreatedBy)
	}

	// Test POST item with empty name
	reqPostBad := httptest.NewRequest("POST", "/api/v1/items", bytes.NewBuffer([]byte(`{"name":""}`)))
	rrPostBad := httptest.NewRecorder()
	a.createItem(rrPostBad, reqPostBad)
	if rrPostBad.Code != http.StatusBadRequest {
		t.Errorf("POST /items with empty name returned %d", rrPostBad.Code)
	}
}

func TestServeDocs(t *testing.T) {
	a := newAPI(false, []byte(`{"openapi":"3.0.3","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`))
	req := httptest.NewRequest("GET", "/docs/api", nil)
	rr := httptest.NewRecorder()

	a.serveDocs(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("docs handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %q", ct)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("Test API")) {
		t.Errorf("expected docs page to contain API title")
	}
}

func TestServeOpenAPISpec(t *testing.T) {
	spec := []byte(`{"openapi":"3.0.3","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`)
	a := newAPI(false, spec)
	req := httptest.NewRequest("GET", "/docs/api/openapi.json", nil)
	rr := httptest.NewRecorder()

	a.serveDocs(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("spec handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content type, got %q", ct)
	}
	if !bytes.Equal(rr.Body.Bytes(), spec) {
		t.Errorf("expected raw spec to be served unchanged")
	}
}
