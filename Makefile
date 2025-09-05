BIN_NAME=github-backup
BUILD_DIR=./build
MODULE := $(shell go list -m)
BUILD=$(shell git rev-parse --short HEAD)@$(shell date +%s)
CURRENT_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH := $(shell uname -m | tr '[:upper:]' '[:lower:]')
LD_FLAGS=-ldflags "-X main.BuildVersion=$(BUILD)"
GO_BUILD=CGO_ENABLED=0 go build $(LD_FLAGS)

.PHONY: build
build:
	$(GO_BUILD) -o $(BUILD_DIR)/ ./...

.PHONY: run
run:
	go run $(LD_FLAGS) .

.PHONY: install
install:
	go install $(LD_FLAGS) .

.PHONY: buildLinuxX86
buildLinuxX86:
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_linux_x86/ ./...

.PHONY: buildLinuxARM64
buildLinuxARM64:
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_linux_arm64/ ./...

.PHONY: buildWindowsX86
buildWindowsX86:
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_windows_x86/ ./...

.PHONY: buildWindowsARM64
buildWindowsARM64:
	GOOS=windows GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_windows_arm64/ ./...

.PHONY: buildDarwinX86
buildDarwinX86:
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_darwin_x86/ ./...

.PHONY: buildDarwinARM64
buildDarwinARM64:
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BIN_NAME)_darwin_arm64/ ./...

.PHONY: buildAll
buildAll: buildLinuxX86 buildLinuxARM64 buildWindowsX86 buildWindowsARM64 buildDarwinX86 buildDarwinARM64

.PHONY: compressAll
compressAll: buildAll
	@cd $(BUILD_DIR) && \
	for dir in */; do \
		base=$${dir%/}; \
		tar -czvf $${base}.tar.gz $${base}; \
	done

.PHONY: buildImage
buildImage:
	docker buildx build --platform=linux/amd64,linux/arm64 -t ghcr.io/tbxark/github-backup:latest . --push --provenance=false

.PHONY: lint
lint:
	go fmt ./...
	go vet ./...
	go get ./...
	go test ./...
	go mod tidy
	golangci-lint fmt --no-config --enable gofmt,goimports
	golangci-lint run --no-config --fix
	nilaway -include-pkgs="$(MODULE)" ./...