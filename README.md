# go-app

A self-contained, self-hostable Go application that compiles application logic, an embedded React web user interface, and interactive OpenAPI documentation into a single executable.

## Features

- **Single Executable Deployment**: Uses Go's `embed` package to bundle the frontend React assets and OpenAPI specifications into a single binary.
- **Offline & Air-gapped Operation**: Designed to operate without internet access; all UI libraries and documentation resources are served locally.
- **Dual Operation Modes**:
  - **Desktop Mode**: Ideal for local home use (Linux, macOS, Windows, Raspberry Pi). Listens on `127.0.0.1`, automatically launches your default web browser, and identifies the user via local environment variables (`USER`, `USERNAME`, `LOGNAME`).
  - **Server Mode**: Suitable for multi-user deployment behind reverse proxies (Traefik, Pangolin, Cloudflare Tunnel, Authelia). Listens on `0.0.0.0` and authenticates users via incoming proxy headers (`X-Forwarded-User`, `Remote-User`, `Pangolin-User`, etc.).
- **Built-in API Documentation**: Auto-renders Swagger UI for OpenAPI specifications directly at `/docs/api`.

---

## Endpoints

| Endpoint | Description |
| --- | --- |
| `/` | Web UI served from embedded React build (supports SPA client-side routing) |
| `/api/v1/health` | Health check endpoint |
| `/api/v1/user` | Current authenticated user context and authentication mode |
| `/api/v1/info` | Runtime system information (uptime, Go version, OS/Arch) |
| `/api/v1/items` | Example REST API endpoints (`GET`, `POST`) |
| `/docs/api` | Interactive OpenAPI / Swagger UI documentation |

---

## Quick Start

### Prerequisites
- [Go](https://golang.org/) 1.24 or later
- [Node.js](https://nodejs.org/) 20 or later (for compiling UI assets)

### Building from Source

To install dependencies, build the React frontend, and compile the Go executable for your current host platform:

```bash
make build
```

This generates a standalone binary named `go-app`.

### Cross-Platform Compilation

You can compile standalone executables for multiple platforms (Windows x64, macOS Intel/ARM, Linux x64, Raspberry Pi ARM):

```bash
# Build for all supported target platforms into ./dist/
make build-all

# Or build for a specific target platform:
make build-linux-amd64     # Linux x64 -> dist/go-app-linux-amd64
make build-windows-amd64   # Windows x64 -> dist/go-app-windows-amd64.exe
make build-darwin-amd64    # macOS Intel x64 -> dist/go-app-darwin-amd64
make build-darwin-arm64    # macOS Apple Silicon -> dist/go-app-darwin-arm64
make build-rpi-arm64       # Raspberry Pi ARM64 -> dist/go-app-rpi-arm64
make build-rpi-armv7       # Raspberry Pi 32-bit ARMv7 -> dist/go-app-rpi-armv7
```

---

## Usage

### Desktop Mode (Default)
Starts the application bound to `127.0.0.1:8080` and opens the web UI in your default browser:

```bash
make run-desktop
# or
./go-app -mode=desktop -port=8080
```

### Server Mode
Starts the server bound to `0.0.0.0:8080` for hosting behind a reverse proxy:

```bash
make run-server
# or
./go-app -mode=server -port=8080
```

### Command-Line Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-mode` | `desktop` | Operation mode (`desktop` or `server`) |
| `-port` | `8080` | Port for backend server to listen on |
| `-host` | `127.0.0.1` (desktop) / `0.0.0.0` (server) | Host IP address to bind to |
| `-no-browser` | `false` | Disable automatic browser launch in desktop mode |

### Environment Variables

| Variable | Default | Description |
| --- | --- | --- |
| `APP_MODE` | `desktop` | Operation mode (`desktop` or `server`) |
| `PORT` | `8080` | Server listening port |
| `HOST` | System default | Host IP address to bind |

---

## Running with Docker

Build and run the minimal containerized application:

```bash
docker build -t go-app .
docker run -d -p 8080:8080 --name go-app go-app
```

Access the UI at `http://localhost:8080` and API docs at `http://localhost:8080/docs/api`.

---

## Testing

Run backend tests:

```bash
make test
# or
go test ./...
```

