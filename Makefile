# Makefile for DesignAI

# 变量定义
BINARY_NAME=design-ai
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 默认目标
.PHONY: all build clean test deps lint docker

all: clean deps lint test build

# 构建
build:
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# 多平台构建
build-all:
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	# macOS ARM64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# 清理
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f *.db

# 测试
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# 测试覆盖率
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 依赖管理
deps:
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

# 代码检查
lint:
	golangci-lint run

# 格式化代码
fmt:
	$(GOCMD) fmt ./...

# 运行应用
run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	./$(BINARY_NAME)

# 开发模式运行
dev:
	$(GOCMD) run $(LDFLAGS) .

# Docker相关
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8080:8080 -v $$(pwd)/data:/app/data $(BINARY_NAME):$(VERSION)

# 发布准备
release-prepare:
	mkdir -p dist
	$(MAKE) build-all

# 安装工具
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 帮助
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  deps         - Download and verify dependencies"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Run in development mode"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  install-tools- Install development tools"
	@echo "  help         - Show this help"