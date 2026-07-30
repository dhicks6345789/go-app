package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"go-app/internal/auth"
)

type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type APIHandler struct {
	startTime    time.Time
	isServerMode bool
	items        []Item
	nextID       int
	mu           sync.RWMutex
	openAPISpec  []byte
}

func NewAPIHandler(isServerMode bool, openAPISpec []byte) *APIHandler {
	return &APIHandler{
		startTime:    time.Now(),
		isServerMode: isServerMode,
		items: []Item{
			{
				ID:        1,
				Name:      "Welcome to Go Self-Contained App",
				CreatedBy: "system",
				CreatedAt: time.Now(),
			},
		},
		nextID:      2,
		openAPISpec: openAPISpec,
	}
}

func (h *APIHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *APIHandler) User(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUser(r, h.isServerMode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

func (h *APIHandler) Info(w http.ResponseWriter, r *http.Request) {
	mode := "desktop"
	if h.isServerMode {
		mode = "server"
	}
	uptime := time.Since(h.startTime).Truncate(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"mode":       mode,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"uptime":     uptime,
	})
}

func (h *APIHandler) Items(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.mu.RLock()
		defer h.mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": h.items,
		})

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
			return
		}

		user := auth.GetUser(r, h.isServerMode)

		h.mu.Lock()
		newItem := Item{
			ID:        h.nextID,
			Name:      req.Name,
			CreatedBy: user.Username,
			CreatedAt: time.Now(),
		}
		h.nextID++
		h.items = append(h.items, newItem)
		h.mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newItem)

	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/docs/api/openapi.json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write(h.openAPISpec)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>API Documentation - Go Self-Contained App</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; background: #0f172a; color: #fff; }
    .swagger-ui { filter: invert(88%%) hue-rotate(180deg); }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/docs/api/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
      });
    };
  </script>
</body>
</html>`)
	w.Write([]byte(html))
}
