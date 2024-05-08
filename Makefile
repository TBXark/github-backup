BIN_NAME=github-backup
BUILD=$(shell git rev-parse --short HEAD)@$(shell date +%s)
CURRENT_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH := $(shell uname -m | tr '[:upper:]' '[:lower:]')
GO_BUILD=CGO_ENABLED=0 go build -ldflags "-X main.BuildVersion=$(BUILD)"

.PHONY: build
build:
	$(GO_BUILD) -o ./build/$(CURRENT_OS)_$(CURRENT_ARCH)/ ./...

.PHONY: run
run:
	go run main.go

.PHONY: buildLinuxX86
buildLinuxX86:
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o ./build/linux_x86/ ./...

.PHONY: buildWindowsX86
buildWindowsX86:
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o ./build/windows_x86/ ./...

.PHONY: buildAll
buildAll: buildLinuxX86 buildWindowsX86 build