# 定义变量
APP_NAME := my-app
BUILD_DIR := build
SRC_DIR := ./cmd/app
MAIN_FILE := $(SRC_DIR)/main.go
LINUX_X86_BIN := $(BUILD_DIR)/$(APP_NAME)-linux-x86
MAC_ARM_BIN := $(BUILD_DIR)/$(APP_NAME)-mac-arm

# Go 工具链
GO := go
GOFMT := gofmt
GOLINT := golangci-lint

# 默认目标
.PHONY: all
all: build

# 构建适配当前平台的可执行文件
.PHONY: build
build:
	@echo ">> Building for local platform (macOS ARM)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) $(GO) build -o $(MAC_ARM_BIN) $(MAIN_FILE)

# 交叉编译：Linux x86
.PHONY: build-linux-x86
build-linux-x86:
	@echo ">> Building for Linux x86..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(LINUX_X86_BIN) $(MAIN_FILE)

# 运行应用程序 (macOS 本地)
.PHONY: run
run:
	@echo ">> Running the application (local)..."
	$(GO) run $(MAIN_FILE)

# 格式化代码
.PHONY: fmt
fmt:
	@echo ">> Formatting code..."
	$(GOFMT) -w .

# 静态分析与代码检查
.PHONY: lint
lint:
	@echo ">> Linting code..."
	$(GOLINT) run ./...

# 运行单元测试
.PHONY: test
test:
	@echo ">> Running tests..."
	$(GO) test -v -cover ./...

# 生成代码覆盖率报告
.PHONY: cover
cover:
	@echo ">> Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 清理构建文件
.PHONY: clean
clean:
	@echo ">> Cleaning up..."
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

# 安装依赖
.PHONY: deps
deps:
	@echo ">> Installing dependencies..."
	$(GO) mod tidy

# 更新依赖
.PHONY: update-deps
update-deps:
	@echo ">> Updating dependencies..."
	$(GO) get -u ./...

# 帮助信息
.PHONY: help
help:
	@echo "Makefile 使用说明"
	@echo
	@echo "可用命令:"
	@echo "  make build           构建本地 (macOS ARM) 平台的应用程序"
	@echo "  make build-linux-x86 交叉编译适配 Linux x86 的应用程序"
	@echo "  make run             运行本地应用程序"
	@echo "  make fmt             格式化代码"
	@echo "  make lint            执行代码静态检查"
	@echo "  make test            运行单元测试"
	@echo "  make cover           生成覆盖率报告"
	@echo "  make clean           清理构建文件"
	@echo "  make deps            安装依赖"
	@echo "  make update-deps     更新依赖到最新版本"
	@echo "  make help            查看帮助信息"
