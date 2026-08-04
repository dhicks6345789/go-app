package main

import (
	"encoding/json"
	"html"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Item is an application resource managed through the API.
type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateItemRequest is the payload accepted when creating a new item.
type CreateItemRequest struct {
	Name string `json:"name" validate:"required" example:"Sample Item"`
}

// UserInfo describes the currently authenticated user.
type UserInfo struct {
	Username string `json:"username" example:"alice"`
	AuthType string `json:"auth_type" example:"local_env"`
	Mode     string `json:"mode" example:"desktop"`
}

// HealthResponse describes the service health status.
type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2026-08-04T12:00:00Z"`
}

// SystemInfoResponse describes system information.
type SystemInfoResponse struct {
	Mode      string `json:"mode" example:"desktop"`
	GoVersion string `json:"go_version" example:"go1.24.4"`
	OS        string `json:"os" example:"linux"`
	Arch      string `json:"arch" example:"amd64"`
	Uptime    string `json:"uptime" example:"12m30s"`
}

// ItemsResponse is the payload returned when listing items.
type ItemsResponse struct {
	Items []Item `json:"items"`
}

// ErrorResponse describes an error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request payload"`
}

// proxyHeaders lists headers set by authenticating reverse proxies
// (Pangolin, Traefik, Cloudflare Tunnel, Authelia, Tailscale, etc.).
var proxyHeaders = []string{
	"X-Forwarded-User",
	"Remote-User",
	"X-User",
	"CF-Access-Authenticated-User-Email",
	"X-Auth-Request-User",
	"Pangolin-User",
}

// api holds the state shared by the API handlers.
type api struct {
	startTime    time.Time
	isServerMode bool
	items        []Item
	nextID       int
	mu           sync.RWMutex
	openAPISpec  []byte
}

// @title Go Self-Contained App API
// @description API documentation for the self-contained Go application framework.
// @version 1.0.0
func newAPI(isServerMode bool, openAPISpec []byte) *api {
	return &api{
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

// getUser resolves the current user from the local environment (desktop mode)
// or from headers set by the authenticating reverse proxy (server mode).
func (a *api) getUser(r *http.Request) UserInfo {
	if a.isServerMode {
		for _, header := range proxyHeaders {
			if val := r.Header.Get(header); val != "" {
				return UserInfo{
					Username: strings.TrimSpace(val),
					AuthType: "proxy_header (" + header + ")",
					Mode:     "server",
				}
			}
		}
		return UserInfo{
			Username: "anonymous",
			AuthType: "none",
			Mode:     "server",
		}
	}

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = os.Getenv("LOGNAME")
	}
	if username == "" {
		username = "localuser"
	}

	return UserInfo{
		Username: username,
		AuthType: "local_env",
		Mode:     "desktop",
	}
}

// health returns the operational status of the service.
// @Summary Health Check
// @Description Returns operational status of the service.
// @Produce json
// @Success 200 {object} HealthResponse "Healthy response"
// @Router /api/v1/health [get]
func (a *api) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// user returns the currently authenticated user.
// @Summary Get Current User
// @Description Returns authenticated user info based on environment or proxy headers.
// @Produce json
// @Success 200 {object} UserInfo
// @Router /api/v1/user [get]
func (a *api) user(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.getUser(r))
}

// info returns system information such as mode, Go version, OS/Arch and uptime.
// @Summary System Info
// @Description Returns system metrics, Go version, OS/Arch, and uptime.
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Router /api/v1/info [get]
func (a *api) info(w http.ResponseWriter, r *http.Request) {
	mode := "desktop"
	if a.isServerMode {
		mode = "server"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"mode":       mode,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"uptime":     time.Since(a.startTime).Truncate(time.Second).String(),
	})
}

// listItems returns all stored items.
// @Summary List Items
// @Description Returns all stored items.
// @Produce json
// @Success 200 {object} ItemsResponse
// @Router /api/v1/items [get]
func (a *api) listItems(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": a.items,
	})
}

// createItem creates a new item and returns it.
// @Summary Create Item
// @Description Creates a new item.
// @Accept json
// @Produce json
// @Param body body CreateItemRequest true "Item to create"
// @Success 201 {object} Item
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/items [post]
func (a *api) createItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request payload"}`))
		return
	}

	user := a.getUser(r)

	a.mu.Lock()
	newItem := Item{
		ID:        a.nextID,
		Name:      req.Name,
		CreatedBy: user.Username,
		CreatedAt: time.Now(),
	}
	a.nextID++
	a.items = append(a.items, newItem)
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

// serveDocs serves the raw OpenAPI specification and an auto-generated
// HTML rendering of it at the /docs/api endpoint.
func (a *api) serveDocs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/docs/api/openapi.json" || r.URL.Path == "/docs/api/swagger.json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write(a.openAPISpec)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(renderDocsHTML(a.openAPISpec))
}

// ---------------------------------------------------------------------------
// OpenAPI documentation rendering
// ---------------------------------------------------------------------------

type openAPISchema struct {
	Type        string                   `json:"type"`
	Format      string                   `json:"format"`
	Ref         string                   `json:"$ref"`
	Description string                   `json:"description"`
	Example     interface{}              `json:"example"`
	Properties  map[string]openAPISchema `json:"properties"`
	Items       *openAPISchema           `json:"items"`
}

type openAPIParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description"`
	Required    bool           `json:"required"`
	Schema      openAPISchema  `json:"schema"`
	Type        string         `json:"type"`
	Format      string         `json:"format"`
	Items       *openAPISchema `json:"items"`
}

type openAPIContent map[string]struct {
	Schema openAPISchema `json:"schema"`
}

type openAPIResponse struct {
	Description string         `json:"description"`
	Schema      openAPISchema  `json:"schema"`
	Content     openAPIContent `json:"content"`
}

type openAPIOperation struct {
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Consumes    []string           `json:"consumes"`
	Produces    []string           `json:"produces"`
	Parameters  []openAPIParameter `json:"parameters"`
	RequestBody struct {
		Description string         `json:"description"`
		Required    bool           `json:"required"`
		Content     openAPIContent `json:"content"`
	} `json:"requestBody"`
	Responses map[string]openAPIResponse `json:"responses"`
}

type openAPIDocument struct {
	OpenAPI string `json:"openapi"`
	Swagger string `json:"swagger"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"info"`
	Servers []struct {
		URL         string `json:"url"`
		Description string `json:"description"`
	} `json:"servers"`
	Host       string                                 `json:"host"`
	BasePath   string                                 `json:"basePath"`
	Paths      map[string]map[string]openAPIOperation `json:"paths"`
	Components struct {
		Schemas map[string]openAPISchema `json:"schemas"`
	} `json:"components"`
	Definitions map[string]openAPISchema `json:"definitions"`
}

var methodOrder = []string{"get", "post", "put", "patch", "delete"}

var methodBadge = map[string]string{
	"get":    "badge text-bg-success",
	"post":   "badge text-bg-primary",
	"put":    "badge text-bg-warning",
	"patch":  "badge text-bg-info",
	"delete": "badge text-bg-danger",
}

func esc(s string) string {
	return html.EscapeString(s)
}

// refName returns the bare schema name for a $ref, accepting both
// OpenAPI 3 "components" and Swagger 2 "definitions" references.
func refName(ref string) string {
	for _, prefix := range []string{"#/components/schemas/", "#/definitions/"} {
		if strings.HasPrefix(ref, prefix) {
			return strings.TrimPrefix(ref, prefix)
		}
	}
	return ref
}

// hasSchema reports whether a schema carries any real information.
func hasSchema(s openAPISchema) bool {
	return s.Ref != "" || s.Type != "" || len(s.Properties) > 0 || s.Items != nil
}

// paramSchema returns the schema describing a parameter, handling both
// Swagger 2 non-body parameters (type at parameter level) and OpenAPI 3
// parameters (schema object).
func paramSchema(p openAPIParameter) openAPISchema {
	if hasSchema(p.Schema) {
		return p.Schema
	}
	return openAPISchema{Type: p.Type, Format: p.Format, Items: p.Items}
}

// schemaInline renders a schema as plain text, dereferencing $refs by name.
func schemaInline(s openAPISchema) string {
	if s.Ref != "" {
		return refName(s.Ref)
	}
	t := s.Type
	if t == "array" && s.Items != nil {
		return "array of " + schemaInline(*s.Items)
	}
	if t == "" {
		t = "object"
	}
	if s.Format != "" {
		t += " (" + s.Format + ")"
	}
	if len(s.Properties) > 0 {
		names := make([]string, 0, len(s.Properties))
		for name := range s.Properties {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, name+": "+schemaInline(s.Properties[name]))
		}
		t += " { " + strings.Join(parts, ", ") + " }"
	}
	return t
}

// schemaHTML renders a schema as HTML, linking $refs to their definitions.
func schemaHTML(s openAPISchema) string {
	if s.Ref != "" {
		name := refName(s.Ref)
		return `<a href="#schema-` + esc(name) + `">` + esc(name) + `</a>`
	}
	return "<code>" + esc(schemaInline(s)) + "</code>"
}

// responseContentHTML renders the content types and schema of a response,
// falling back to Swagger 2 "produces"/"schema" fields when the OpenAPI 3
// "content" object is absent.
func responseContentHTML(op openAPIOperation, resp openAPIResponse) string {
	if len(resp.Content) > 0 {
		return contentHTML(resp.Content)
	}
	if !hasSchema(resp.Schema) {
		return "-"
	}
	if len(op.Produces) > 0 {
		parts := make([]string, 0, len(op.Produces))
		for _, mt := range op.Produces {
			parts = append(parts, "<code>"+esc(mt)+"</code> "+schemaHTML(resp.Schema))
		}
		return strings.Join(parts, "<br/>")
	}
	return "<code>application/json</code> " + schemaHTML(resp.Schema)
}

func contentHTML(c openAPIContent) string {
	if len(c) == 0 {
		return "-"
	}
	types := make([]string, 0, len(c))
	for mediaType := range c {
		types = append(types, mediaType)
	}
	sort.Strings(types)
	parts := make([]string, 0, len(types))
	for _, mediaType := range types {
		parts = append(parts, "<code>"+esc(mediaType)+"</code> "+schemaHTML(c[mediaType].Schema))
	}
	return strings.Join(parts, "<br/>")
}

func renderDocsHTML(spec []byte) []byte {
	var doc openAPIDocument
	if err := json.Unmarshal(spec, &doc); err != nil {
		return []byte("<!DOCTYPE html><html><head><meta charset=\"utf-8\"/><title>API Documentation</title></head><body><p>Failed to parse OpenAPI specification.</p></body></html>")
	}

	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>API Documentation - ` + esc(doc.Info.Title) + `</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #0f172a; color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; }
.topbar { background: #1e293b; border-bottom: 1px solid #334155; padding: 1rem 2rem; display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; }
.topbar h1 { font-size: 1.25rem; font-weight: 600; color: #38bdf8; }
.topbar a { color: #94a3b8; font-size: 0.9rem; text-decoration: none; margin-left: auto; }
.topbar a:hover { color: #38bdf8; }
.container { max-width: 1000px; margin: 0 auto; padding: 2rem 1rem; }
h2 { color: #38bdf8; font-size: 1.5rem; margin: 2rem 0 1rem; border-bottom: 1px solid #334155; padding-bottom: 0.5rem; }
h3 { font-size: 1.1rem; margin-bottom: 0.5rem; }
p { color: #cbd5e1; margin-bottom: 0.5rem; }
.endpoint { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 1.25rem; margin-bottom: 1.25rem; }
.ep-head { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem; }
.ep-path { font-size: 1.05rem; color: #f8fafc; }
.ep-summary { font-weight: 600; color: #f8fafc; margin-bottom: 0.25rem; }
.ep-desc { color: #94a3b8; margin-bottom: 1rem; }
.badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: 6px; font-size: 0.75rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; }
.text-bg-success { background: #10b981; color: #052e16; }
.text-bg-primary { background: #38bdf8; color: #082f49; }
.text-bg-warning { background: #f59e0b; color: #451a03; }
.text-bg-info { background: #22d3ee; color: #083344; }
.text-bg-danger { background: #ef4444; color: #450a0a; }
table { width: 100%; border-collapse: collapse; margin: 0.75rem 0; font-size: 0.9rem; }
th, td { text-align: left; padding: 0.5rem 0.75rem; border-bottom: 1px solid #334155; vertical-align: top; }
th { color: #94a3b8; font-weight: 600; }
code { font-family: 'Consolas', 'Monaco', 'Courier New', monospace; font-size: 0.85rem; color: #a7f3d0; background: #0f172a; padding: 0.15rem 0.4rem; border-radius: 4px; }
a { color: #38bdf8; text-decoration: none; }
a:hover { text-decoration: underline; }
.schema { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 1.25rem; margin-bottom: 1.25rem; }
.subhead { color: #94a3b8; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; margin-top: 1rem; }
</style>
</head>
<body>
<div class="topbar">
  <h1>` + esc(doc.Info.Title) + `</h1>
  <a href="../">Back to Home</a>
</div>
<div class="container">
`)

	version := doc.Info.Version
	if version == "" {
		version = "-"
	}
	specVersion := doc.OpenAPI
	if specVersion == "" {
		specVersion = doc.Swagger
	}
	if specVersion == "" {
		specVersion = "-"
	}
	b.WriteString(`<h2>Overview</h2>`)
	b.WriteString(`<p><strong>Version:</strong> ` + esc(version) + ` &middot; <strong>Spec:</strong> OpenAPI ` + esc(specVersion) + `</p>`)
	if doc.Info.Description != "" {
		b.WriteString(`<p>` + esc(doc.Info.Description) + `</p>`)
	}
	if len(doc.Servers) > 0 {
		b.WriteString(`<p><strong>Servers:</strong> `)
		servers := make([]string, 0, len(doc.Servers))
		for _, srv := range doc.Servers {
			label := srv.URL
			if srv.Description != "" {
				label += " (" + srv.Description + ")"
			}
			servers = append(servers, "<code>"+esc(label)+"</code>")
		}
		b.WriteString(strings.Join(servers, " "))
		b.WriteString(`</p>`)
	} else if doc.Host != "" || doc.BasePath != "" {
		b.WriteString(`<p><strong>Servers:</strong> <code>` + esc(doc.Host+doc.BasePath) + `</code></p>`)
	}
	b.WriteString(`<p><a href="swagger.json" download>Download OpenAPI specification (swagger.json)</a></p>`)

	b.WriteString(`<h2>Endpoints</h2>`)
	paths := make([]string, 0, len(doc.Paths))
	for path := range doc.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		ops := doc.Paths[path]
		methods := make([]string, 0)
		seen := map[string]bool{}
		for _, m := range methodOrder {
			if _, ok := ops[m]; ok {
				methods = append(methods, m)
				seen[m] = true
			}
		}
		for m := range ops {
			if !seen[m] {
				methods = append(methods, m)
			}
		}

		for _, method := range methods {
			op := ops[method]
			badgeClass := methodBadge[method]
			if badgeClass == "" {
				badgeClass = "badge text-bg-primary"
			}

			b.WriteString(`<div class="endpoint">`)
			b.WriteString(`<div class="ep-head"><span class="` + esc(badgeClass) + `">` + strings.ToUpper(method) + `</span><code class="ep-path">` + esc(path) + `</code></div>`)
			if op.Summary != "" {
				b.WriteString(`<p class="ep-summary">` + esc(op.Summary) + `</p>`)
			}
			if op.Description != "" {
				b.WriteString(`<p class="ep-desc">` + esc(op.Description) + `</p>`)
			}

			nonBodyParams := make([]openAPIParameter, 0)
			var bodyParam *openAPIParameter
			for i := range op.Parameters {
				if op.Parameters[i].In == "body" {
					p := op.Parameters[i]
					bodyParam = &p
				} else {
					nonBodyParams = append(nonBodyParams, op.Parameters[i])
				}
			}

			if len(nonBodyParams) > 0 {
				b.WriteString(`<div class="subhead">Parameters</div>`)
				b.WriteString(`<table><tr><th>Name</th><th>In</th><th>Required</th><th>Type</th><th>Description</th></tr>`)
				for _, p := range nonBodyParams {
					required := "no"
					if p.Required {
						required = "yes"
					}
					b.WriteString(`<tr><td><code>` + esc(p.Name) + `</code></td><td>` + esc(p.In) + `</td><td>` + required + `</td><td>` + schemaHTML(paramSchema(p)) + `</td><td>` + esc(p.Description) + `</td></tr>`)
				}
				b.WriteString(`</table>`)
			}

			if len(op.RequestBody.Content) > 0 {
				b.WriteString(`<div class="subhead">Request Body</div>`)
				req := op.RequestBody
				desc := req.Description
				if req.Required {
					if desc != "" {
						desc += " "
					}
					desc += "(required)"
				}
				if desc != "" {
					b.WriteString(`<p>` + esc(desc) + `</p>`)
				}
				b.WriteString(`<p>` + contentHTML(req.Content) + `</p>`)
			} else if bodyParam != nil {
				b.WriteString(`<div class="subhead">Request Body</div>`)
				desc := bodyParam.Description
				if bodyParam.Required {
					if desc != "" {
						desc += " "
					}
					desc += "(required)"
				}
				if desc != "" {
					b.WriteString(`<p>` + esc(desc) + `</p>`)
				}
				mediaType := "application/json"
				if len(op.Consumes) > 0 {
					mediaType = op.Consumes[0]
				}
				b.WriteString(`<p><code>` + esc(mediaType) + `</code> ` + schemaHTML(paramSchema(*bodyParam)) + `</p>`)
			}

			b.WriteString(`<div class="subhead">Responses</div>`)
			codes := make([]string, 0, len(op.Responses))
			for code := range op.Responses {
				codes = append(codes, code)
			}
			sort.Strings(codes)
			b.WriteString(`<table><tr><th>Code</th><th>Description</th><th>Content</th></tr>`)
			for _, code := range codes {
				resp := op.Responses[code]
				b.WriteString(`<tr><td><code>` + esc(code) + `</code></td><td>` + esc(resp.Description) + `</td><td>` + responseContentHTML(op, resp) + `</td></tr>`)
			}
			b.WriteString(`</table>`)
			b.WriteString(`</div>`)
		}
	}

	schemas := doc.Components.Schemas
	if len(schemas) == 0 {
		schemas = doc.Definitions
	}
	if len(schemas) > 0 {
		b.WriteString(`<h2>Schemas</h2>`)
		names := make([]string, 0, len(schemas))
		for name := range schemas {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			s := schemas[name]
			b.WriteString(`<div class="schema"><h3 id="schema-` + esc(name) + `">` + esc(name) + `</h3>`)
			b.WriteString(`<p>Type: <code>` + esc(schemaInline(s)) + `</code></p>`)
			if s.Description != "" {
				b.WriteString(`<p>` + esc(s.Description) + `</p>`)
			}
			if len(s.Properties) > 0 {
				b.WriteString(`<table><tr><th>Property</th><th>Type</th><th>Description</th></tr>`)
				props := make([]string, 0, len(s.Properties))
				for p := range s.Properties {
					props = append(props, p)
				}
				sort.Strings(props)
				for _, p := range props {
					prop := s.Properties[p]
					b.WriteString(`<tr><td><code>` + esc(p) + `</code></td><td>` + schemaHTML(prop) + `</td><td>` + esc(prop.Description) + `</td></tr>`)
				}
				b.WriteString(`</table>`)
			}
			b.WriteString(`</div>`)
		}
	}

	b.WriteString(`</div>`)
	b.WriteString(`</body>`)
	b.WriteString(`</html>`)

	return []byte(b.String())
}
