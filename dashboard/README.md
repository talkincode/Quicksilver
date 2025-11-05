# Quicksilver 管理后台

基于 Streamlit 的 Quicksilver 模拟交易所管理仪表盘。

## 技术栈

- **Python**: 3.11+
- **Streamlit**: 1.51.0 (最新版)
- **包管理**: uv (现代化 Python 包管理工具)

## 快速开始

### 1. 安装 uv (如果还没有)

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置 API 凭证
```

### 3. 启动服务

```bash
chmod +x start.sh
./start.sh
```

或手动启动：

```bash
# 同步依赖
uv sync

# 运行应用
uv run streamlit run app.py
```

## 开发

### 添加新依赖

```bash
uv add <package-name>
```

### 更新依赖

```bash
uv sync --upgrade
```

### 锁定依赖

```bash
uv lock
```

## 项目结构

```
dashboard/
├── app.py              # 主入口
├── config.py           # 配置管理
├── pyproject.toml      # 项目配置和依赖
├── uv.lock             # 依赖锁定文件
├── api/                # API 客户端
│   ├── __init__.py
│   └── client.py
├── pages/              # 页面模块
│   ├── home.py         # 首页
│   ├── users.py        # 用户管理
│   ├── orders.py       # 订单管理
│   └── trades.py       # 成交记录
└── .venv/              # 虚拟环境 (uv 自动创建)
```

## 功能特性

### 🏠 首页

- 系统状态监控
- 数据概览统计
- 实时行情展示
- 用户权益曲线

### 👥 用户管理

- 创建新用户（弹窗）
- 用户列表查询
- 邮箱搜索

### 📝 订单管理

- 订单列表查询
- 订单详情查看

### 💰 成交记录

- 成交记录查询
- 统计分析

## 环境变量

在 `.env` 文件中配置：

```bash
# Quicksilver API 配置
API_URL=http://localhost:8080
ADMIN_API_KEY=your-api-key
ADMIN_API_SECRET=your-api-secret
```

## 访问地址

默认端口：http://localhost:8501

## 注意事项

- 首次启动会自动创建虚拟环境 `.venv`
- 使用 uv 管理依赖，比传统 pip 更快
- 确保后端 API 服务正在运行
