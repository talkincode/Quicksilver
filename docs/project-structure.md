# Quicksilver 项目结构说明

## 📁 目录结构

```
quicksilver/
├── cmd/                          # 应用程序入口点
│   └── server/
│       └── main.go              # 主服务入口
│
├── internal/                     # 私有应用代码
│   ├── api/                     # HTTP API 处理器
│   │   └── handlers.go          # API 路由处理函数
│   ├── config/                  # 配置管理
│   │   └── config.go            # 配置结构和加载逻辑
│   ├── database/                # 数据库连接
│   │   └── database.go          # 数据库初始化和迁移
│   ├── model/                   # 数据模型
│   │   └── models.go            # GORM 模型定义
│   ├── router/                  # 路由配置
│   │   └── router.go            # 路由注册
│   ├── service/                 # 业务逻辑层 (待实现)
│   ├── repository/              # 数据访问层 (待实现)
│   ├── engine/                  # 撮合引擎 (待实现)
│   └── middleware/              # 中间件 (待实现)
│
├── pkg/                         # 可复用的公共库
│   └── utils/                   # 工具函数 (待实现)
│
├── config/                      # 配置文件
│   └── config.example.yaml      # 配置示例
│
├── db/                          # 数据库相关
│   ├── init.sql                 # 初始化脚本
│   ├── migrations/              # 迁移脚本 (待实现)
│   └── seeds/                   # 测试数据 (待实现)
│
├── docs/                        # 文档
│   ├── system-design-mvp.md     # MVP 系统设计
│   └── database.md              # 数据库设计文档
│
├── scripts/                     # 脚本工具 (待创建)
│   ├── migrate.go               # 数据库迁移
│   └── seed.go                  # 数据填充
│
├── .air.toml                    # Air 热重载配置
├── .gitignore                   # Git 忽略文件
├── docker-compose.yml           # Docker Compose 配置
├── Dockerfile                   # Docker 镜像配置
├── go.mod                       # Go 模块定义
├── Makefile                     # 构建脚本
└── README.md                    # 项目说明
```

## 🏗️ 架构分层

```
┌─────────────────────────────────────────┐
│         cmd/server (入口层)             │
│         • 应用启动                      │
│         • 依赖注入                      │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         internal/router (路由层)        │
│         • 路由注册                      │
│         • 中间件配置                    │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         internal/api (控制器层)         │
│         • 请求处理                      │
│         • 参数验证                      │
│         • 响应封装                      │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         internal/service (业务层)       │
│         • 业务逻辑                      │
│         • 事务管理                      │
│         • 撮合引擎                      │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│      internal/repository (数据层)       │
│         • 数据访问                      │
│         • CRUD 操作                     │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         internal/model (模型层)         │
│         • 数据结构                      │
│         • ORM 映射                      │
└─────────────────────────────────────────┘
```

## 📦 核心模块说明

### cmd/server

- **用途**: 应用程序入口
- **职责**:
  - 初始化配置
  - 初始化数据库连接
  - 启动 HTTP 服务器
  - 优雅关闭处理

### internal/config

- **用途**: 配置管理
- **职责**:
  - 读取 YAML 配置文件
  - 支持环境变量覆盖
  - 提供配置结构体

### internal/database

- **用途**: 数据库连接管理
- **职责**:
  - 建立数据库连接
  - 配置连接池
  - 自动迁移表结构

### internal/model

- **用途**: 数据模型定义
- **职责**:
  - 定义 GORM 模型
  - 定义表关系
  - 定义 JSON 序列化

### internal/router

- **用途**: 路由配置
- **职责**:
  - 注册 API 路由
  - 配置中间件
  - 路由分组管理

### internal/api

- **用途**: API 处理器
- **职责**:
  - 处理 HTTP 请求
  - 调用业务逻辑
  - 返回 JSON 响应

### internal/service (待实现)

- **用途**: 业务逻辑层
- **计划功能**:
  - 订单处理服务
  - 撮合引擎服务
  - 余额管理服务
  - 行情同步服务

### internal/repository (待实现)

- **用途**: 数据访问层
- **计划功能**:
  - 用户仓储
  - 订单仓储
  - 交易仓储
  - 行情仓储

### internal/engine (待实现)

- **用途**: 撮合引擎
- **计划功能**:
  - 市价单撮合
  - 限价单撮合
  - 订单簿管理

## 🚀 快速开始

### 1. 初始化项目

```bash
# 下载依赖
go mod download

# 复制配置文件
cp config/config.example.yaml config/config.yaml
```

### 2. 启动数据库

```bash
# 使用 Docker Compose
docker-compose up -d db

# 或手动启动 PostgreSQL
```

### 3. 运行应用

```bash
# 开发模式 (热重载)
make dev

# 或直接运行
make run

# 或使用 go run
go run cmd/server/main.go
```

### 4. 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 获取服务器时间
curl http://localhost:8080/v1/time

# 获取交易对
curl http://localhost:8080/v1/markets
```

## 📝 开发规范

### 代码组织

- 每个包有明确的职责
- 避免循环依赖
- 使用依赖注入

### 命名规范

- 文件名: 小写 + 下划线 (`user_service.go`)
- 包名: 小写单数 (`package user`)
- 接口: 以 `er` 结尾 (`Handler`, `Repository`)
- 结构体: 驼峰命名 (`UserService`)

### 错误处理

- 使用 `fmt.Errorf` 包装错误
- 在合适的层级处理错误
- 返回有意义的错误信息

### 日志

- 使用 zap 结构化日志
- 记录关键操作
- 不记录敏感信息

## 🔄 下一步开发

### Phase 1: 基础设施 ✅

- [x] 项目结构
- [x] 配置管理
- [x] 数据库连接
- [x] 基础路由

### Phase 2: 核心功能 (进行中)

- [ ] Service 层实现
- [ ] Repository 层实现
- [ ] 用户认证中间件
- [ ] API Key 管理

### Phase 3: 业务逻辑

- [ ] 订单创建
- [ ] 撮合引擎
- [ ] 余额管理
- [ ] 行情同步

### Phase 4: 测试 & 优化

- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档完善

## 📚 参考资料

- [Echo 文档](https://echo.labstack.com/)
- [GORM 文档](https://gorm.io/)
- [CCXT 文档](https://docs.ccxt.com/)
- [Go 项目布局](https://github.com/golang-standards/project-layout)
