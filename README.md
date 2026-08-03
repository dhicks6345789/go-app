# Go App

A project to act as an example framework to produce a self-hostable Go application that contains application logic, an embedded Bootstrap web user interface and interactive OpenAPI documentation into a single executable.

This application itself doesn't do anything much, it just presents a basic user interface showing the current user name. 

## Features

- **Single Executable Deployment**: Uses Go's `embed` package to bundle the frontend React assets, documentation and OpenAPI specifications into a single binary.
- **Offline Operation**: Designed to be able to operate without internet access; all UI libraries and documentation resources are served locally.
- **Dual Operation Modes**:
  - **Desktop Mode**: Ideal for local home desktop use on pretty much any platform (Linux, MacOS, Windows, Raspberry Pi). Running the executable on your desktop machine should give you a localhost-only server and automatically launch your default web browser to display the user interface.
  - **Server Mode**: Suitable for multi-user deployment behind authenticating reverse proxies (Pangolin / Traefik, Cloudflare Tunnel, Authelia, Tailscale). Authenticates users via incoming proxy headers (`X-Forwarded-User`, `Remote-User`, `Pangolin-User`, etc.). As a single, statically linked Go binary with no external dependencies, it can be run inside a very minimal container environment.
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

This will compile the executables for all platforms and generate documentation, including Swaggo's interactive API documentation.

You can copy the generated files directly to somewhere they can be served as a web site, you just need to specify the path you want the files to go to, e.g.:

```
make dist DEST_DIR=~/www/go-app
```

## Using As a Basis For Your Own Projects

The entire purpose of this project is just to act as a basic starting point for a self-contained "app" that is easy for end users to run and use. It should produce executables able to run on your preferred platform. If you just compile and run the basic project you will get a minimal application that just reports back the username, useful to test your build process is working okay.

Extending this project should be a case of cloning the Git repository and adding your own functions to `internal/handlers/handlers.go` and user interface elements to `ui/index.html`. You can, of course, use and extend the AGENTS.md file with a suitable AI coding agent.
