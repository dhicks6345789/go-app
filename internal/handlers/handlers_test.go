package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	h := NewAPIHandler(false, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

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
	h := NewAPIHandler(false, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/user", nil)
	rr := httptest.NewRecorder()

	h.User(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["mode"] != "desktop" {
		t.Errorf("expected mode 'desktop', got %s", resp["mode"])
	}
}

func TestUserEndpointServerProxyHeader(t *testing.T) {
	h := NewAPIHandler(true, []byte("{}"))
	req := httptest.NewRequest("GET", "/api/v1/user", nil)
	req.Header.Set("Remote-User", "alice_proxy")
	rr := httptest.NewRecorder()

	h.User(rr, req)

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["username"] != "alice_proxy" {
		t.Errorf("expected username 'alice_proxy', got %s", resp["username"])
	}
}

func TestItemsCRUD(t *testing.T) {
	h := NewAPIHandler(false, []byte("{}"))

	// Test GET items
	reqGet := httptest.NewRequest("GET", "/api/v1/items", nil)
	rrGet := httptest.NewRecorder()
	h.Items(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Errorf("GET /items returned %d", rrGet.Code)
	}

	// Test POST item
	payload := []byte(`{"name":"New Test Item"}`)
	reqPost := httptest.NewRequest("POST", "/api/v1/items", bytes.NewBuffer(payload))
	rrPost := httptest.NewRecorder()
	h.Items(rrPost, reqPost)

	if rrPost.Code != http.StatusCreated {
		t.Errorf("POST /items returned %d", rrPost.Code)
	}
}
