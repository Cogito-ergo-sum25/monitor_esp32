.PHONY: all build build-linux build-windows build-mac clean run-mock

APP_NAME = desk-monitor
BIN_DIR = bin

all: build-linux build-windows

build: build-linux

build-linux:
	@echo "==> Compilando binario estático para Linux (amd64)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-linux .
	@echo "Listo: $(BIN_DIR)/$(APP_NAME)-linux"

build-windows:
	@echo "==> Compilando ejecutable para Windows (amd64)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-windows.exe .
	@echo "Listo: $(BIN_DIR)/$(APP_NAME)-windows.exe"

build-mac:
	@echo "==> Compilando binario para macOS (arm64 Apple Silicon)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 .
	@echo "Listo: $(BIN_DIR)/$(APP_NAME)-darwin-arm64"

run-mock:
	go run . --mock

clean:
	rm -rf $(BIN_DIR)

