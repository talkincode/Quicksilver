# Quicksilver - CCXT 兼容的模拟交易所

> 一个轻量级、高性能的模拟交易所系统，兼容 CCXT API

## 项目简介

Quicksilver 是一个专为量化交易策略测试设计的模拟交易所系统。它提供与真实交易所相同的 API 接口，让您可以在无风险的环境中测试和优化交易策略。

### 核心特性

- ✅ **CCXT 兼容** - 支持 CCXT 标准 API，无缝对接现有策略
- ✅ **实时行情** - 从 Binance 同步真实市场数据
- ✅ **完整撮合** - 支持市价单、限价单撮合
- ✅ **账户管理** - 余额管理、资金冻结/解冻
- ✅ **高性能** - Go 语言实现，单机支持 1000+ TPS

### 技术栈

- **语言**: Go 1.21+
- **框架**: Echo (Web), GORM (ORM)
- **数据库**: PostgreSQL 16+
- **缓存**: 内存缓存
- **部署**: Docker + Docker Compose

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (可选)

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/talkincode/quicksilver.git
cd quicksilver

# 2. 安装依赖
go mod download

# 3. 配置环境变量
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 设置数据库连接等信息

# 4. 初始化数据库
make db-migrate

# 5. 启动服务
make run
```

### 使用 Docker

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

## 项目结构

```
quicksilver/
├── cmd/                    # 应用入口
│   └── server/            # 主服务
├── internal/              # 私有代码
│   ├── api/              # API 处理器
│   ├── config/           # 配置管理
│   ├── engine/           # 撮合引擎
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问
│   ├── service/          # 业务逻辑
│   └── middleware/       # 中间件
├── pkg/                   # 公共库
│   └── utils/            # 工具函数
├── db/                    # 数据库相关
│   ├── migrations/       # 迁移脚本
│   └── seeds/            # 测试数据
├── config/                # 配置文件
├── docs/                  # 文档
├── scripts/               # 脚本工具
├── docker-compose.yml     # Docker 编排
├── Dockerfile            # Docker 镜像
├── Makefile              # 构建脚本
└── go.mod                # Go 模块
```

## API 文档

启动服务后，访问 API 文档：

- Swagger UI: `http://localhost:8080/swagger/index.html`
- API 端点: `http://localhost:8080/v1/`

### 基础使用示例

```python
import ccxt

# 初始化交易所
exchange = ccxt.Exchange({
    'apiKey': 'your-api-key',
    'secret': 'your-api-secret',
    'urls': {
        'api': 'http://localhost:8080/v1'
    }
})

# 查询余额
balance = exchange.fetch_balance()

# 下单
order = exchange.create_order(
    symbol='BTC/USDT',
    type='limit',
    side='buy',
    amount=0.01,
    price=50000
)

# 查询订单
order_info = exchange.fetch_order(order['id'])
```

## 开发指南

### 运行测试

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 查看覆盖率
make test-coverage
```

### 代码规范

```bash
# 格式化代码
make fmt

# 代码检查
make lint

# 生成文档
make docs
```

## 配置说明

主要配置项 (`config/config.yaml`):

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: quicksilver
  user: postgres
  password: password

market:
  update_interval: 1s
  data_source: binance
```

## 路线图

- [x] v1.0 - MVP 版本
  - [x] 基础 API 实现
  - [x] 市价/限价单撮合
  - [x] BTC/USDT 交易对
- [ ] v1.5 - 功能增强
  - [ ] 多交易对支持
  - [ ] WebSocket 推送
  - [ ] 订单薄可视化
- [ ] v2.0 - 架构升级
  - [ ] 微服务拆分
  - [ ] 合约交易
  - [ ] 高可用部署

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 联系方式

- 项目主页: https://github.com/talkincode/quicksilver
- Issue 跟踪: https://github.com/talkincode/quicksilver/issues
- 邮箱: dev@quicksilver.local

## 致谢

- [CCXT](https://github.com/ccxt/ccxt) - 统一的加密货币交易所 API
- [Echo](https://echo.labstack.com/) - 高性能 Go Web 框架
- [GORM](https://gorm.io/) - Go ORM 库
