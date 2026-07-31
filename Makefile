.PHONY: all build build-ui build-all build-linux-amd64 build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-rpi-arm64 build-rpi-armv7 index dist run-desktop run-server clean test

OUT_DIR := dist

all: build

build-ui:
	@echo "==> Building React Frontend..."
	@cd ui && npm install && npm run build

build: build-ui
	@echo "==> Compiling Go Executable (Current Platform)..."
	@go build -o go-app main.go
	@echo "==> Build Complete: ./go-app"

build-linux-amd64: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (Linux x64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-linux-amd64 main.go
	@echo "==> Built: $(OUT_DIR)/go-app-linux-amd64"

build-windows-amd64: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (Windows x64)..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-windows-amd64.exe main.go
	@echo "==> Built: $(OUT_DIR)/go-app-windows-amd64.exe"

build-darwin-amd64: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (macOS x64)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-darwin-amd64 main.go
	@echo "==> Built: $(OUT_DIR)/go-app-darwin-amd64"

build-darwin-arm64: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (macOS ARM64)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-darwin-arm64 main.go
	@echo "==> Built: $(OUT_DIR)/go-app-darwin-arm64"

build-rpi-arm64: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (Raspberry Pi / Linux ARM64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-rpi-arm64 main.go
	@echo "==> Built: $(OUT_DIR)/go-app-rpi-arm64"

build-rpi-armv7: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling Go Executable (Raspberry Pi 32-bit ARMv7)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-rpi-armv7 main.go
	@echo "==> Built: $(OUT_DIR)/go-app-rpi-armv7"

build-all: build-ui
	@mkdir -p $(OUT_DIR)
	@echo "==> Compiling for All Target Platforms..."
	@echo " -> Linux x64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-linux-amd64 main.go
	@echo " -> Windows x64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-windows-amd64.exe main.go
	@echo " -> macOS Intel x64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-darwin-amd64 main.go
	@echo " -> macOS Apple Silicon ARM64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-darwin-arm64 main.go
	@echo " -> Raspberry Pi / Linux ARM64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-rpi-arm64 main.go
	@echo " -> Raspberry Pi 32-bit ARMv7..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -o $(OUT_DIR)/go-app-rpi-armv7 main.go
	@echo "==> All builds complete in ./$(OUT_DIR)/"
	@$(MAKE) index

index:
	@echo "==> Generating repository index.html..."
	@python3 scripts/generate-index.py

dist:
	@if [ -z "$(DEST_DIR)" ]; then echo "Error: DEST_DIR not set. Usage: make dist DEST_DIR=/path/to/site"; exit 1; fi
	@$(MAKE) build-all
	@echo "==> Staging Distribution Files to $(DEST_DIR)..."
	@mkdir -p $(DEST_DIR)/docs
	@cp index.html $(DEST_DIR)/
	@cp docs/* $(DEST_DIR)/docs/
	@cp dist/* $(DEST_DIR)/
	@echo "==> Distribution Complete: $(DEST_DIR)"

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
	@rm -rf $(OUT_DIR)
	@rm -rf ui/dist


