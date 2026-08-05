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
	"net/url"
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
	basePath := flag.String("base-path", getEnv("BASE_PATH", ""), "URL prefix to serve under when mounted behind a reverse proxy at a sub-path (e.g. /magooify)")
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
	base := normalizeBasePath(*basePath)

	a := newAPI(isServerMode, docsFS)

	handler := buildHandler(a, base)

	user := a.getUser(&http.Request{})

	log.Printf("==================================================")
	log.Printf("Go Self-Contained App Starting")
	log.Printf("Mode      : %s", strings.ToUpper(user.Mode))
	log.Printf("User      : %s (%s)", user.Username, user.AuthType)
	log.Printf("Listening : http://%s", addr)
	if base != "" {
		log.Printf("Base path : %s", base)
	}
	log.Printf("UI        : http://%s%s/", addr, base)
	log.Printf("OpenAPI   : http://%s%s/docs/api", addr, base)
	log.Printf("==================================================")

	// Auto-launch web browser in desktop mode
	if !isServerMode && !*noBrowser {
		targetURL := fmt.Sprintf("http://127.0.0.1:%s%s/", *port, base)
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

// normalizeBasePath cleans a configured base path so it can be used for
// prefix matching. An empty value (or "/") means the app is served from the
// root of the site.
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

// basePathHandler wraps the application handler so it can be served from a
// reverse-proxy sub-path such as /magooify. Requests arriving with that path
// prefix have the prefix stripped before they reach the application routes;
// requests without the prefix (e.g. direct access or a proxy that strips it
// itself) are passed through unchanged.
func basePathHandler(basePath string, next http.Handler) http.Handler {
	basePath = normalizeBasePath(basePath)
	if basePath == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		var stripped string

		switch {
		case p == basePath:
			// Redirect to the trailing-slash form so the browser resolves
			// relative URLs (vendor/, style.css, app.js, api/...) against the
			// base path instead of the site root.
			location := basePath + "/"
			if r.URL.RawQuery != "" {
				location += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, location, http.StatusMovedPermanently)
			return
		case p == basePath+"/":
			stripped = "/"
		case strings.HasPrefix(p, basePath+"/"):
			stripped = strings.TrimPrefix(p, basePath)
		default:
			next.ServeHTTP(w, r)
			return
		}

		r2 := r.Clone(r.Context())
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = stripped
		r2.URL.RawPath = ""
		next.ServeHTTP(w, r2)
	})
}

// buildHandler wires up all HTTP routes for the application, wrapping them
// with support for being served from an optional reverse-proxy base path.
func buildHandler(a *api, basePath string) http.Handler {
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

		// SPA fallback to index.html with the base path injected so relative
		// asset and API URLs resolve under the reverse-proxy sub-path.
		serveIndex(w, r, subFS, baseHrefFor(r, basePath))
	})

	return basePathHandler(basePath, mux)
}

// baseHrefFor returns the <base href> value for the request. No root URL is
// hard-coded: the proxy-supplied X-Forwarded-Prefix header (set by
// Traefik/Pangolin when stripping a sub-path) reflects the external URL this
// specific request came through, so it takes precedence. That lets one app
// instance be served simultaneously from several sub-paths. The configured
// base path is only a fallback for proxies that don't forward the header.
func baseHrefFor(r *http.Request, configured string) string {
	if base := safeForwardedPrefix(r.Header.Get("X-Forwarded-Prefix")); base != "" {
		return base + "/"
	}
	base := normalizeBasePath(configured)
	if base == "" {
		return "/"
	}
	return base + "/"
}

// safeForwardedPrefix validates a proxy-supplied X-Forwarded-Prefix header
// before it is trusted for building URLs. The header is attacker controllable
// on misconfigured proxy chains, so only a safe path-absolute value such as
// "/magooify" is accepted; anything that could redirect asset URLs elsewhere
// (protocol-relative, dot segments, query/fragment characters) is rejected.
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

// serveIndex writes the embedded index.html with a <base> tag pointing at the
// effective base path. This keeps relative URLs (vendor/, style.css, app.js,
// api/...) resolving under the proxy sub-path regardless of whether the proxy
// forwards the request path unchanged or strips it before reaching the app.
func serveIndex(w http.ResponseWriter, r *http.Request, subFS fs.FS, baseHref string) {
	data, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	html := strings.Replace(string(data), "<head>", "<head>\n    <base href=\""+baseHref+"\"/>", 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Never cache the page: the injected <base href> depends on the deployment
	// base path, and a stale cached copy misdirects asset and API URLs.
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(html))
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
