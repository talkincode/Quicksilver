.PHONY: help run build test clean docker-up docker-down db-migrate db-seed

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := quicksilver
VERSION := 1.0.0
BUILD_DIR := bin
MAIN_PATH := cmd/server/main.go

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

help: ## 显示帮助信息
	@echo "Quicksilver - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## 运行应用
	@echo "Starting $(APP_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

build: ## 编译应用
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

test: ## 运行所有测试
	@echo "Running all tests..."
	CGO_ENABLED=1 $(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage summary:"
	@$(GOCMD) tool cover -func=coverage.out | grep total

test-unit: ## 运行单元测试
	@echo "Running unit tests..."
	CGO_ENABLED=1 $(GOTEST) -v -short -coverprofile=coverage.out ./...

test-integration: ## 运行集成测试
	@echo "Running integration tests..."
	CGO_ENABLED=1 $(GOTEST) -v -run Integration ./...

test-watch: ## 监听文件变化自动运行测试
	@echo "Watching for changes..."
	@which gotestsum > /dev/null || (echo "Installing gotestsum..." && go install gotest.tools/gotestsum@latest)
	gotestsum --watch -- -short ./...

test-coverage: test ## 生成测试覆盖率报告
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt: ## 格式化代码
	@echo "Formatting code..."
	$(GOFMT) ./...

lint: ## 代码检查
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

tidy: ## 整理依赖
	@echo "Tidying modules..."
	$(GOMOD) tidy

clean: ## 清理构建产物
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## 构建 Docker 镜像
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

docker-up: ## 启动 Docker 服务
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down: ## 停止 Docker 服务
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f app

db-migrate: ## 运行数据库迁移
	@echo "Running database migrations..."
	$(GOCMD) run scripts/migrate.go up

db-rollback: ## 回滚数据库迁移
	@echo "Rolling back database migrations..."
	$(GOCMD) run scripts/migrate.go down

db-seed: ## 填充测试数据
	@echo "Seeding database..."
	$(GOCMD) run scripts/seed.go

db-reset: ## 重置数据库（危险操作！）
	@echo "WARNING: This will drop all tables!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(GOCMD) run scripts/migrate.go reset; \
		$(GOCMD) run scripts/migrate.go up; \
		$(GOCMD) run scripts/seed.go; \
	fi

install: ## 安装依赖
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

dev: ## 开发模式（热重载）
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

.PHONY: all
all: clean fmt lint test build ## 执行完整构建流程
