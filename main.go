package main

// @title           Go Self-Contained App API
// @version         1.0.0
// @description     API documentation for the self-contained Go application framework.
// @host            localhost:8080
// @BasePath        /api/v1

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"go-app/internal/auth"
	"go-app/internal/browser"
	"go-app/internal/handlers"
)

//go:embed ui/dist/*
var uiFS embed.FS

//go:embed docs/swagger.json
var openAPISpec []byte

func main() {
	defaultPort := getEnv("PORT", "8080")
	defaultMode := getEnv("APP_MODE", "desktop")
	defaultHost := getEnv("HOST", "")

	port := flag.String("port", defaultPort, "Port to listen on")
	mode := flag.String("mode", defaultMode, "Operation mode: 'desktop' or 'server'")
	host := flag.String("host", defaultHost, "Host IP to bind to (defaults to 127.0.0.1 for desktop, 0.0.0.0 for server)")
	noBrowser := flag.Bool("no-browser", false, "Disable automatic browser launch in desktop mode")
	flag.Parse()

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

	// API Handlers
	apiHandler := handlers.NewAPIHandler(isServerMode, openAPISpec)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", apiHandler.Health)
	mux.HandleFunc("/api/v1/user", apiHandler.User)
	mux.HandleFunc("/api/v1/info", apiHandler.Info)
	mux.HandleFunc("GET /api/v1/items", apiHandler.ListItems)
	mux.HandleFunc("POST /api/v1/items", apiHandler.CreateItem)

	// Docs handlers
	mux.HandleFunc("/docs/api", apiHandler.ServeDocs)
	mux.HandleFunc("/docs/api/", apiHandler.ServeDocs)

	// Static UI setup with SPA fallback
	subFS, err := fs.Sub(uiFS, "ui/dist")
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
		if path == "" {
			fileServer.ServeHTTP(w, r)
			return
		}

		f, err := subFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback to index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	user := auth.GetUser(&http.Request{}, isServerMode)

	log.Printf("==================================================")
	log.Printf("⚡ Go Self-Contained App Starting")
	log.Printf("Mode      : %s", strings.ToUpper(user.Mode))
	log.Printf("User      : %s (%s)", user.Username, user.AuthType)
	log.Printf("Listening : http://%s", addr)
	log.Printf("OpenAPI   : http://%s/docs/api", addr)
	log.Printf("==================================================")

	// Auto-launch web browser in desktop mode
	if !isServerMode && !*noBrowser {
		targetURL := fmt.Sprintf("http://127.0.0.1:%s", *port)
		go func() {
			log.Printf("Launching browser targeting %s...", targetURL)
			if err := browser.Open(targetURL); err != nil {
				log.Printf("Note: Could not open web browser automatically: %v", err)
			}
		}()
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	if err := http.Serve(listener, mux); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
