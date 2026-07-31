# Go App

A project to act as an example framework to produce a self-hostable Go application that contains application logic, an embedded React web user interface and interactive OpenAPI documentation into a single executable.

This application itself doesn't do anything much, it just presents a basic user interface showing the current user name. 

## Features

- **Single Executable Deployment**: Uses Go's `embed` package to bundle the frontend React assets, documentation and OpenAPI specifications into a single binary.
- **Offline Operation**: Designed to be able to operate without internet access; all UI libraries and documentation resources are served locally.
- **Dual Operation Modes**:
  - **Desktop Mode**: Ideal for local home desktop use on pretty much any platform (Linux, MacOS, Windows, Raspberry Pi). Running the executable on your desktop machine should give you a localhost-only server and automatically launch your default web browser to display the user interface.
  - **Server Mode**: Suitable for multi-user deployment behind authenticating reverse proxies (Pangolin / Traefik, Cloudflare Tunnel, Authelia). Authenticates users via incoming proxy headers (`X-Forwarded-User`, `Remote-User`, `Pangolin-User`, etc.). As a single, statically linked Go binary with no external dependencies, it can be run inside a very minimal container environment.
- **Built for use by humans and AI agents**: With built-in documentation and Swagger UI - you should be able to point an AI agent at the API documentation and have it start using it right away.

---

## Running

### Desktop

Simply download and run the executable for your platform from the [project homepage](https://sansay.co.uk/go-app/).

### Server

You can run Go App behind an authenticating proxy server - the proxy server authenticates the user and passes the username to the application via a simple HTTP header. For instance, if you were using Pangolin as your authenticating server, you would add a basic container to hold the Go App executable:

```
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/go-app /app/go-app
```
And to docker compose, add:

```
go-app -mode=server -port=8080
```

## Building

Clone the repository:

```
git clone https://github.com/dhicks6345789/go-app.git
```

And run build-all:

```
cd go-app
make build-all
```

This will compile the executables for all platforms and generate documentation, including Swaggo's interactive API documentation. You can copy the generated files directly to somewhere they can be served as a web site:

```
mkdir -p ~/www/go-app
mkdir -p ~/www/go-app/docs
cp index.html ~/www/go-app
cp docs/* ~/www/go-app/docs
cp dist/* ~/www/go-app
```

## Using As a Basis For Your Own Projects

Add you own functions to internal/handlers/handlers.go.

Add more user interface to ui/index.html.
