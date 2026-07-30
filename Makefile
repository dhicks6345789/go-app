.PHONY: all build build-ui run-desktop run-server clean test

all: build

build-ui:
	@echo "==> Building React Frontend..."
	@cd ui && npm install && npm run build

build: build-ui
	@echo "==> Compiling Go Executable..."
	@go build -o go-app main.go
	@echo "==> Build Complete: ./go-app"

run-desktop: build
	@echo "==> Starting in Desktop Mode..."
	@./go-app -mode=desktop -port=8080

run-server: build
	@echo "==> Starting in Server Mode..."
	@./go-app -mode=server -port=8080

test:
	@echo "==> Running Tests..."
	@go test ./...

clean:
	@echo "==> Cleaning Build Artifacts..."
	@rm -f go-app
	@rm -rf ui/dist
