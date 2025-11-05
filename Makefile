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

test-ccxt: ## 运行 CCXT 兼容性测试（Python）
	@echo "Running CCXT compatibility tests (Python)..."
	@cd scripts && python3 test_ccxt.py

test-ccxt-js: ## 运行 CCXT 兼容性测试（Node.js）
	@echo "Running CCXT compatibility tests (Node.js)..."
	@cd scripts && node test_ccxt.js

test-ccxt-setup-python: ## 安装 CCXT 测试依赖（Python）
	@echo "Installing Python CCXT test dependencies..."
	@cd scripts && pip3 install -r requirements-test.txt

test-ccxt-setup-nodejs: ## 安装 CCXT 测试依赖（Node.js）
	@echo "Installing Node.js CCXT test dependencies..."
	@cd scripts && npm install

test-api: ## 运行 REST API 自动化测试
	@echo "Running REST API tests..."
	@chmod +x scripts/api_test.sh
	@./scripts/api_test.sh

test-perf: ## 运行性能测试（需要 k6）
	@echo "Running performance tests..."
	@which k6 > /dev/null || (echo "Error: k6 not installed. Install: brew install k6" && exit 1)
	k6 run scripts/performance_test.js

test-perf-report: ## 运行性能测试并生成报告
	@echo "Running performance tests with report..."
	@which k6 > /dev/null || (echo "Error: k6 not installed. Install: brew install k6" && exit 1)
	k6 run --out json=perf_results.json scripts/performance_test.js
	@echo "Performance report saved to: perf_results.json"

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

db-seed-test-user: ## 创建测试用户和初始余额
	@echo "Creating test user for API testing..."
	@which psql > /dev/null || (echo "Error: psql not found. Please install PostgreSQL client." && exit 1)
	@psql -h localhost -U postgres -d quicksilver -f db/seed_test_user.sql
	@echo ""
	@echo "✅ Test user created successfully!"
	@echo "Use these credentials in apitest.http:"
	@echo "  API Key: qs-test-api-key-2024"
	@echo "  API Secret: qs-test-api-secret-change-in-production"

db-seed-test-user-docker: ## 在 Docker 容器中创建测试用户
	@echo "Creating test user in Docker container..."
	@docker exec -i quicksilver-postgres psql -U postgres -d quicksilver < db/seed_test_user.sql
	@echo ""
	@echo "✅ Test user created successfully in Docker!"
	@echo "Use these credentials in apitest.http:"
	@echo "  API Key: qs-test-api-key-2024"
	@echo "  API Secret: qs-test-api-secret-change-in-production"

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
