# 项目信息
APP_NAME := my-app
VERSION := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 目录结构
BUILD_DIR := build
SRC_DIR := ./cmd/app
MAIN_FILE := $(SRC_DIR)/main.go
DIST_DIR := dist

# 输出文件
LINUX_AMD64_BIN := $(BUILD_DIR)/$(APP_NAME)-linux-amd64
LINUX_ARM64_BIN := $(BUILD_DIR)/$(APP_NAME)-linux-arm64
MAC_AMD64_BIN := $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
MAC_ARM64_BIN := $(BUILD_DIR)/$(APP_NAME)-darwin-arm64

# Go 工具链
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := go test
TIMEOUT := 15s

# 编译标记
LDFLAGS := -ldflags="-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.CommitSHA=$(COMMIT_SHA)'"
GOBUILD := CGO_ENABLED=0 $(GO) build $(LDFLAGS)

# 颜色输出
BLUE := \033[34m
GREEN := \033[32m
RED := \033[31m
YELLOW := \033[33m
RESET := \033[0m

# 默认目标
.PHONY: all
all: clean build test lint

# 构建所有平台
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-mac-amd64 build-mac-arm64

# 各平台构建目标
.PHONY: build-linux-amd64
build-linux-amd64: wire
	@printf "$(BLUE)>> Building for Linux AMD64...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(LINUX_AMD64_BIN) cmd/app/main.go cmd/app/wire_gen.go

.PHONY: build-linux-arm64
build-linux-arm64: wire
	@printf "$(BLUE)>> Building for Linux ARM64...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(LINUX_ARM64_BIN) cmd/app/main.go cmd/app/wire_gen.go

.PHONY: build-mac-amd64
build-mac-amd64: wire
	@printf "$(BLUE)>> Building for macOS AMD64...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(MAC_AMD64_BIN) cmd/app/main.go cmd/app/wire_gen.go

.PHONY: build-mac-arm64
build-mac-arm64: wire
	@printf "$(BLUE)>> Building for macOS ARM64...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(MAC_ARM64_BIN) cmd/app/main.go cmd/app/wire_gen.go

# 构建当前平台
.PHONY: build
build: wire
	@printf "$(BLUE)>> Building for current platform...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) cmd/app/main.go cmd/app/wire_gen.go

# 生成 wire
.PHONY: wire
wire:
	@printf "$(BLUE)>> Generating wire_gen.go...$(RESET)\n"
	@cd cmd/app && wire

# 开发模式运行（支持热重载）
.PHONY: dev
dev: wire
	@printf "$(GREEN)>> Running in development mode...$(RESET)\n"
	@which air > /dev/null || go install github.com/air-verse/air@latest
	@mkdir -p tmp
	air

# 运行应用
.PHONY: run
run: wire
	@printf "$(GREEN)>> Running application...$(RESET)\n"
	@$(GO) run cmd/app/main.go cmd/app/wire_gen.go

# 代码质量检查
.PHONY: lint
lint:
	@printf "$(BLUE)>> Running linter...$(RESET)\n"
	@which $(GOLINT) > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
	$(GOLINT) run ./... --timeout=$(TIMEOUT)

# 格式化代码
.PHONY: fmt
fmt:
	@printf "$(BLUE)>> Formatting code...$(RESET)\n"
	$(GOFMT) -s -w .

# 测试
.PHONY: test
test:
	@printf "$(BLUE)>> Running tests...$(RESET)\n"
	$(GOTEST) -v -race -cover ./...

# 生成测试覆盖率报告
.PHONY: coverage
coverage:
	@printf "$(BLUE)>> Generating coverage report...$(RESET)\n"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)Coverage report generated: coverage.html$(RESET)\n"

# 依赖管理
.PHONY: deps
deps:
	@printf "$(BLUE)>> Installing dependencies...$(RESET)\n"
	$(GO) mod download
	$(GO) mod tidy

# 更新依赖
.PHONY: deps-update
deps-update:
	@printf "$(BLUE)>> Updating dependencies...$(RESET)\n"
	$(GO) get -u ./...
	$(GO) mod tidy

# 生成发布包
.PHONY: dist
dist: build-all
	@printf "$(BLUE)>> Creating distribution packages...$(RESET)\n"
	@mkdir -p $(DIST_DIR)
	@for bin in $(BUILD_DIR)/*; do \
		if [ -f "$$bin" ]; then \
			tar czf $(DIST_DIR)/$$(basename $$bin).tar.gz -C $(BUILD_DIR) $$(basename $$bin); \
			printf "$(GREEN)Created: $(DIST_DIR)/$$(basename $$bin).tar.gz$(RESET)\n"; \
		fi \
	done

# 清理构建文件
.PHONY: clean
clean:
	@printf "$(YELLOW)>> Cleaning up...$(RESET)\n"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html

# 版本信息
.PHONY: version
version:
	@printf "$(BLUE)Version: $(VERSION)\n"
	@printf "Build Time: $(BUILD_TIME)\n"
	@printf "Commit: $(COMMIT_SHA)$(RESET)\n"

# 检查工具安装
.PHONY: check-tools
check-tools:
	@printf "$(BLUE)>> Checking required tools...$(RESET)\n"
	@which $(GOLINT) > /dev/null || printf "$(YELLOW)golangci-lint is not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)\n"
	@which air > /dev/null || printf "$(YELLOW)air is not installed. Run: go install github.com/air-verse/air@latest$(RESET)\n"

# 帮助信息
.PHONY: help
help:
	@printf "$(BLUE)可用命令:$(RESET)\n"
	@printf "$(GREEN)基础命令:$(RESET)\n"
	@printf "  make build          - 构建当前平台的可执行文件\n"
	@printf "  make run           - 运行应用程序\n"
	@printf "  make dev           - 开发模式运行（支持热重载）\n"
	@printf "  make clean         - 清理构建文件\n"
	@printf "  make wire          - 生成依赖注入代码\n"
	@printf "\n$(GREEN)构建相关:$(RESET)\n"
	@printf "  make build-all     - 构建所有支持的平台\n"
	@printf "  make dist          - 创建发布包\n"
	@printf "\n$(GREEN)开发工具:$(RESET)\n"
	@printf "  make fmt           - 格式化代码\n"
	@printf "  make lint          - 运行代码检查\n"
	@printf "  make test          - 运行测试\n"
	@printf "  make coverage      - 生成测试覆盖率报告\n"
	@printf "\n$(GREEN)依赖管理:$(RESET)\n"
	@printf "  make deps          - 安装依赖\n"
	@printf "  make deps-update   - 更新依赖\n"
	@printf "\n$(GREEN)其他:$(RESET)\n"
	@printf "  make version       - 显示版本信息\n"
	@printf "  make check-tools   - 检查必要工具是否安装\n"
