# 🚀 Quicksilver 快速启动指南

## 项目已初始化完成！

恭喜！您的 Quicksilver 项目结构已经创建完成。以下是下一步操作指南。

## 📋 已创建的文件

```
✅ README.md                    # 项目说明文档
✅ go.mod                       # Go 模块定义
✅ Makefile                     # 构建脚本
✅ Dockerfile                   # Docker 镜像配置
✅ docker-compose.yml           # Docker 编排
✅ .gitignore                   # Git 忽略文件
✅ .air.toml                    # 热重载配置

配置文件:
✅ config/config.example.yaml   # 配置示例

代码文件:
✅ cmd/server/main.go          # 主程序入口
✅ internal/config/config.go   # 配置管理
✅ internal/database/database.go # 数据库连接
✅ internal/model/models.go    # 数据模型
✅ internal/router/router.go   # 路由配置
✅ internal/api/handlers.go    # API 处理器

数据库:
✅ db/init.sql                 # 数据库初始化脚本

文档:
✅ docs/system-design-mvp.md   # MVP 系统设计
✅ docs/database.md            # 数据库设计文档
✅ docs/project-structure.md   # 项目结构说明
```

## 🎯 下一步操作

### Step 1: 初始化 Go 模块

```bash
cd /Volumes/ExtDISK/github/Quicksilver

# 初始化 go.mod (如果需要修改模块路径)
# go mod init github.com/your-username/quicksilver

# 下载依赖
go mod download
go mod tidy
```

### Step 2: 创建配置文件

```bash
# 复制配置示例
cp config/config.example.yaml config/config.yaml

# 编辑配置文件 (可选)
# vim config/config.yaml
```

### Step 3: 启动数据库

**方式 1: 使用 Docker Compose (推荐)**

```bash
# 启动 PostgreSQL
docker-compose up -d db

# 查看日志
docker-compose logs -f db

# 等待数据库就绪 (约 10 秒)
```

**方式 2: 本地 PostgreSQL**

```bash
# 确保本地已安装 PostgreSQL 16+
# 创建数据库
createdb quicksilver

# 或使用 psql
psql -U postgres -c "CREATE DATABASE quicksilver;"
```

### Step 4: 运行应用

**方式 1: 使用 Make (推荐)**

```bash
# 开发模式 (需要先安装 air)
make dev

# 或直接运行
make run
```

**方式 2: 直接使用 Go**

```bash
go run cmd/server/main.go
```

**方式 3: 完整 Docker 部署**

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app
```

### Step 5: 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 服务器时间
curl http://localhost:8080/v1/time

# 获取交易对
curl http://localhost:8080/v1/markets

# 获取行情 (需要先同步数据)
curl http://localhost:8080/v1/ticker/BTC/USDT
```

## 🛠️ 开发工具安装

### 安装 Air (热重载工具)

```bash
go install github.com/cosmtrek/air@latest
```

### 安装 golangci-lint (代码检查)

```bash
# macOS
brew install golangci-lint

# 或使用 go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 安装 migrate (数据库迁移工具 - 可选)

```bash
# macOS
brew install golang-migrate

# 其他系统
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## 📝 常用命令

### 开发相关

```bash
make help          # 查看所有命令
make run           # 运行应用
make dev           # 开发模式 (热重载)
make build         # 编译应用
make test          # 运行测试
make fmt           # 格式化代码
make lint          # 代码检查
make clean         # 清理构建产物
```

### Docker 相关

```bash
make docker-build  # 构建镜像
make docker-up     # 启动服务
make docker-down   # 停止服务
make docker-logs   # 查看日志
```

### 数据库相关

```bash
make db-migrate    # 运行迁移 (待实现)
make db-seed       # 填充测试数据 (待实现)
make db-reset      # 重置数据库 (待实现)
```

## 🔧 配置说明

主配置文件: `config/config.yaml`

```yaml
server:
  port: 8080 # 服务端口
  mode: debug # 运行模式: debug/release

database:
  host: localhost # 数据库主机
  port: 5432 # 数据库端口
  name: quicksilver # 数据库名
  user: postgres # 用户名
  password: password # 密码

market:
  update_interval: 1s # 行情更新间隔
  data_source: binance # 数据源
  symbols:
    - BTC/USDT # 支持的交易对

trading:
  default_fee_rate: 0.001 # 默认手续费率 0.1%
  maker_fee_rate: 0.0005 # Maker 费率 0.05%
  taker_fee_rate: 0.001 # Taker 费率 0.1%
```

## ⚠️ 注意事项

### 当前状态

这是一个 **精简的初始化版本**，包含：

✅ 完整的项目结构
✅ 基础的 HTTP 服务器
✅ 数据库连接和模型
✅ 基础 API 路由
✅ Docker 部署配置

⏳ 待实现的功能：

- [ ] Service 业务逻辑层
- [ ] Repository 数据访问层
- [ ] 用户认证中间件
- [ ] 订单创建和撮合逻辑
- [ ] 行情数据同步
- [ ] 余额管理逻辑
- [ ] 完整的单元测试

### 编译错误说明

当前代码可能存在编译错误，这是正常的，因为：

1. `go.mod` 需要先运行 `go mod download`
2. 某些导入的包还未安装
3. 部分功能标记为 `TODO` 待实现

**解决方法**:

```bash
# 1. 下载依赖
go mod download
go mod tidy

# 2. 编译检查
go build ./...

# 3. 如果有错误，安装缺失的包
go get -u github.com/labstack/echo/v4
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

## 📚 下一步开发建议

### Phase 1: 完善基础设施 (1-2 天)

1. 实现 Service 层
2. 实现 Repository 层
3. 添加认证中间件
4. 完善错误处理

### Phase 2: 实现核心功能 (3-5 天)

1. 用户注册和 API Key 生成
2. 余额管理 (冻结/解冻)
3. 行情数据同步服务
4. 订单创建流程

### Phase 3: 实现撮合引擎 (5-7 天)

1. 市价单撮合
2. 限价单撮合
3. 订单状态管理
4. 成交记录生成

### Phase 4: 测试和优化 (3-5 天)

1. 单元测试
2. 集成测试
3. 性能测试
4. 文档完善

## 🆘 常见问题

### Q: 无法连接数据库？

```bash
# 检查数据库是否运行
docker-compose ps

# 查看数据库日志
docker-compose logs db

# 重启数据库
docker-compose restart db
```

### Q: 端口被占用？

```bash
# 检查端口占用
lsof -i :8080

# 修改配置文件中的端口
vim config/config.yaml
```

### Q: 依赖下载失败？

```bash
# 设置 Go 代理 (中国用户)
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod download
```

## 📖 推荐阅读

- [项目结构说明](docs/project-structure.md)
- [MVP 系统设计](docs/system-design-mvp.md)
- [数据库设计文档](docs/database.md)
- [Echo 框架文档](https://echo.labstack.com/)
- [GORM 文档](https://gorm.io/)

---

**祝开发顺利！** 🎉

如有问题，请查看文档或提交 Issue。
