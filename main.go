package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

//go:embed all:ui
var uiFS embed.FS

//go:embed all:docs
var docsFS embed.FS

// registerMimeTypes pins deterministic Content-Type headers for embedded
// assets so the UI and docs work even on hosts without a system MIME database.
func registerMimeTypes() {
	mime.AddExtensionType(".html", "text/html; charset=utf-8")
	mime.AddExtensionType(".htm", "text/html; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".mjs", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".map", "application/json")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".webp", "image/webp")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")
	mime.AddExtensionType(".eot", "application/vnd.ms-fontobject")
	mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
	mime.AddExtensionType(".xml", "text/xml; charset=utf-8")
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".pdf", "application/pdf")
}

func main() {
	registerMimeTypes()

	defaultPort := getEnv("PORT", "8080")
	defaultMode := getEnv("APP_MODE", "desktop")
	defaultHost := getEnv("HOST", "")

	port := flag.String("port", defaultPort, "Port to listen on")
	mode := flag.String("mode", defaultMode, "Operation mode: 'desktop' or 'server'")
	host := flag.String("host", defaultHost, "Host IP to bind to (defaults to 127.0.0.1 for desktop, 0.0.0.0 for server)")
	noBrowser := flag.Bool("no-browser", false, "Disable automatic browser launch in desktop mode")
	genDocs := flag.String("gen-docs", "", "Write the Swagger UI documentation page to the given path and exit")
	flag.Parse()

	// Offline documentation generation (used by the build script to produce docs/api.html).
	if *genDocs != "" {
		if err := os.WriteFile(*genDocs, swaggerUIHTML(), 0o644); err != nil {
			log.Fatalf("Failed to write docs to %s: %v", *genDocs, err)
		}
		log.Printf("Wrote API documentation to %s", *genDocs)
		return
	}

	isServerMode := strings.ToLower(*mode) == "server"

	bindHost := *host
	if bindHost == "" {
		if isServerMode {
			bindHost = "0.0.0.0"
		} else {
			bindHost = "127.0.0.1"
		}
	}

	addr := net.JoinHostPort(bindHost, *port)

	a := newAPI(isServerMode, docsFS)

	handler := buildHandler(a)

	user := a.getUser(&http.Request{})

	log.Printf("==================================================")
	log.Printf("Go Self-Contained App Starting")
	log.Printf("Mode      : %s", strings.ToUpper(user.Mode))
	log.Printf("User      : %s (%s)", user.Username, user.AuthType)
	log.Printf("Listening : http://%s", addr)
	log.Printf("UI        : http://%s/", addr)
	log.Printf("OpenAPI   : http://%s/docs/api", addr)
	log.Printf("==================================================")

	// Auto-launch web browser in desktop mode
	if !isServerMode && !*noBrowser {
		targetURL := fmt.Sprintf("http://127.0.0.1:%s/", *port)
		go func() {
			log.Printf("Launching browser targeting %s...", targetURL)
			if err := openBrowser(targetURL); err != nil {
				log.Printf("Note: Could not open web browser automatically: %v", err)
			}
		}()
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	if err := http.Serve(listener, handler); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

// buildHandler wires up all HTTP routes for the application. The app is
// served from the site root; when mounted behind a reverse proxy at a
// sub-path such as /magooify, the proxy is expected to strip that prefix
// (reporting it via X-Forwarded-Prefix) before requests reach the app.
func buildHandler(a *api) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", a.health)
	mux.HandleFunc("/api/v1/user", a.user)
	mux.HandleFunc("/api/v1/info", a.info)
	mux.HandleFunc("GET /api/v1/items", a.listItems)
	mux.HandleFunc("POST /api/v1/items", a.createItem)

	mux.HandleFunc("/docs/api", a.serveDocs)
	mux.HandleFunc("/docs/", a.serveDocs)

	subFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatalf("Failed to initialize embedded UI filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/docs") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" && path != "index.html" {
			f, err := subFS.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// A stripping proxy forwards a bare prefix such as "/magooify" as an
		// empty path. Redirect to the trailing-slash form so the browser loads
		// the index document from a URL ending in "/". Only this case is
		// redirected: ordinary resource requests like /magooify/app.js are
		// served as-is, never bounced to a trailing slash.
		if r.URL.Path == "" {
			if loc, ok := indexRedirectLocation(r); ok {
				http.Redirect(w, r, loc, http.StatusMovedPermanently)
				return
			}
		}

		// SPA fallback to index.html. The page uses plain relative links, so
		// resources resolve against the document URL automatically, whether the
		// app is served from the site root or a reverse-proxy sub-path.
		serveIndex(w, r, subFS)
	})

	return mux
}

// indexRedirectLocation returns the Location header for a 301 that adds a
// trailing slash when a stripping proxy (Traefik/Pangolin) forwards a bare
// prefix such as "/magooify" as an empty path. ok is false when the request
// does not need a redirect. The prefix the proxy reports via
// X-Forwarded-Prefix is re-attached so the redirect points at the external
// URL the browser sees; otherwise it falls back to the site root.
func indexRedirectLocation(r *http.Request) (string, bool) {
	if r.URL.Path != "" {
		return "", false
	}

	prefix := safeForwardedPrefix(r.Header.Get("X-Forwarded-Prefix"))
	if prefix == "" {
		return "/", true
	}
	return prefix + "/", true
}

// safeForwardedPrefix validates a proxy-supplied X-Forwarded-Prefix header
// before it is trusted for building a redirect URL. The header is attacker
// controllable on misconfigured proxy chains, so only a safe path-absolute
// value such as "/magooify" is accepted; anything that could turn the redirect
// into an open redirect (protocol-relative, dot segments, query/fragment
// characters) is rejected.
func safeForwardedPrefix(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || !strings.HasPrefix(val, "/") || strings.HasPrefix(val, "//") {
		return ""
	}
	if strings.ContainsAny(val, "\\?#\r\n") || strings.Contains(val, "..") {
		return ""
	}
	return normalizeBasePath(val)
}

// normalizeBasePath cleans a reported reverse-proxy prefix so it can be used
// in a redirect URL. An empty value (or "/") means the site root.
func normalizeBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimSuffix(p, "/")
}

// serveIndex writes the embedded index.html. The page uses only plain
// relative links, so no <base> tag is needed: resources resolve against the
// document URL, which keeps them working under any reverse-proxy sub-path as
// long as the request URL carries a trailing slash.
func serveIndex(w http.ResponseWriter, r *http.Request, subFS fs.FS) {
	data, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Never cache the page so a stale copy can't linger after redeploys.
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// openBrowser launches the system's default web browser at the given URL.
func openBrowser(url string) error {
	// Give the server a moment to start listening.
	time.Sleep(100 * time.Millisecond)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
