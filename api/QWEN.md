# ChatJimmy2API - Go Server

## 项目概述

这是一个用 **Go 1.22+** 实现的高性能 API 包装器，用于代理 `https://chatjimmy.ai` 服务，提供 **OpenAI 兼容** 的 API 接口。

### 核心技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.22+ |
| Web 框架 | Gin |
| 配置热重载 | fsnotify |
| WebSocket | gorilla/websocket |
| UUID | google/uuid |

### 主要功能

- **OpenAI 兼容端点** - `/v1/models`, `/v1/chat/completions`
- **配置热重载** - 修改 `config/config.json` 后自动生效，无需重启
- **实时统计** - 请求次数、Token 消耗、错误统计，持久化存储
- **Web 管理界面** - 可视化配置、状态监控、日志查看
- **流式输出模拟** - 支持 `fake`（按词延迟）和 `batch`（按块大小）两种模式
- **API 认证** - Bearer Token 认证
- **结构化日志** - JSON 格式日志，支持文件轮换

## 项目结构

```
go-server/
├── main.go                     # 应用入口
├── go.mod / go.sum             # Go 模块依赖
├── Makefile                    # 构建脚本
├── start.sh                    # 快速启动脚本
├── config/
│   ├── config.go               # 配置管理器（热重载）
│   ├── config.json             # 当前配置
│   └── config_test.go          # 配置测试
├── internal/
│   ├── client/                 # 上游 HTTP 客户端
│   ├── handler/                # API 和管理界面处理器
│   ├── logger/                 # 日志系统
│   ├── metrics/                # 统计指标管理
│   ├── stream/                 # 流式模拟器
│   ├── transform/              # 请求/响应格式转换
│   └── types/                  # 类型定义
├── static/                     # Web 管理界面 HTML
├── logs/                       # 日志目录
└── data/                       # 统计数据持久化
```

## 构建与运行

### 环境要求

- Go 1.22+

### 快速开始

```bash
# 1. 下载依赖
go mod download

# 2. 运行（开发模式）
go run main.go

# 或使用启动脚本
./start.sh
```

### 编译

```bash
# 当前平台
make build

# 交叉编译
make build-linux    # Linux amd64
make build-macos    # macOS arm64
make build-windows  # Windows amd64
make build-all      # 所有平台
```

### 运行测试

```bash
# 运行所有测试
make test
# 或
go test ./...

# 带覆盖率
make test-coverage
```

### 清理

```bash
make clean
```

## 配置说明

配置文件：`config/config.json`

### 关键配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `upstream_base_url` | 上游服务地址 | `https://chatjimmy.ai` |
| `upstream_api_key` | 上游 API 密钥 | `""` |
| `upstream_timeout_ms` | 超时时间 (ms) | `15000` |
| `wrapper_api_key` | 本地 API 认证密钥 | `local-wrapper-key` |
| `port` | API 服务端口 | `8787` |
| `admin_port` | 管理界面端口 | `8788` |
| `stream_mode` | 流式模式：`fake` / `batch` | `fake` |
| `fake_stream_delay_ms` | 伪造流式每词延迟 | `50` |
| `batch_stream_size` | 批量流式每块字符数 | `100` |
| `log_file` | 日志文件路径 | `logs/server.log` |
| `stats_file` | 统计数据文件 | `data/stats.json` |

### 配置热重载

修改 `config/config.json` 后自动生效，无需重启服务。

## API 端点

### OpenAI 兼容端点 (端口 8787)

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/v1/models` | GET | 获取模型列表 |
| `/v1/chat/completions` | POST | 聊天补全 |

### 管理界面端点 (端口 8788)

| 端点 | 方法 | 说明 |
|------|------|------|
| `/` | GET | 管理首页 |
| `/config` | GET | 配置页面 |
| `/stats` | GET | 统计页面 |
| `/logs` | GET | 日志页面 |
| `/api/config` | GET/POST | 配置 API |
| `/api/stats` | GET | 统计 API |
| `/api/logs` | GET | 日志 API |

## 使用示例

### curl 请求

```bash
# 健康检查
curl http://127.0.0.1:8787/health

# 获取模型列表
curl -sS -H "Authorization: Bearer local-wrapper-key" \
  "http://127.0.0.1:8787/v1/models" | jq

# 非流式聊天
curl -sS -H "Authorization: Bearer local-wrapper-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","messages":[{"role":"user","content":"Hello!"}]}' \
  "http://127.0.0.1:8787/v1/chat/completions" | jq

# 流式聊天
curl -sS -N -H "Authorization: Bearer local-wrapper-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","stream":true,"messages":[{"role":"user","content":"Hello!"}]}' \
  "http://127.0.0.1:8787/v1/chat/completions"
```

### Node.js 示例

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://127.0.0.1:8787/v1',
  apiKey: 'local-wrapper-key',
});

const completion = await openai.chat.completions.create({
  model: 'llama3.1-8B',
  messages: [{ role: 'user', content: 'Hello!' }],
});
```

## 开发约定

### 代码风格

- 遵循 Go 官方 `gofmt` 格式化
- 使用 `go vet` 进行静态检查
- 包命名采用小写，避免下划线

### 测试实践

- 每个核心模块都有对应的 `_test.go` 文件
- 测试覆盖配置、流式模拟、格式转换、统计指标

### 日志规范

JSON 格式结构化日志：

```json
{"time":"2024-01-01T12:00:00Z","level":"INFO","message":"服务器启动","fields":{"version":"1.0.0"}}
```

### 错误处理

- 所有错误记录到日志并返回适当的 HTTP 状态码
- 使用 `map[string]interface{}` 传递日志上下文

## 常见问题

### 修改配置后不生效

确保修改的是 `config/config.json`，文件保存后会自动热重载。

### 流式输出过快/过慢

调整配置：
- `fake_stream_delay_ms` - 控制每词之间的延迟
- `batch_stream_size` - 控制每块的大小

### 查看日志

```bash
tail -f logs/server.log
# 或访问管理界面 http://127.0.0.1:8788/logs
```
