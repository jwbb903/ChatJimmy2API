# ChatJimmy2API - 项目上下文

## 项目概述

**ChatJimmy2API** 是一个高性能的 Go 语言实现的 OpenAI 兼容 API 包装器，用于代理 `https://chatjimmy.ai` 服务。项目支持两种部署模式：
1. **独立服务器模式** - 传统的 Go HTTP 服务
2. **Vercel Serverless 模式** - 无服务器函数部署

### 核心技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **配置管理**: 支持热重载（基于 fsnotify）
- **部署平台**: Vercel Serverless / 独立服务器
- **前端设计**: Material Design 3 (MD3)

### 主要功能

- ✅ OpenAI 兼容 API 端点 (`/v1/models`, `/v1/chat/completions`)
- ✅ 配置热重载（修改配置文件自动生效）
- ✅ 实时统计（请求次数、Token 消耗、错误统计）
- ✅ Web 管理界面（Material Design 3 风格）
- ✅ 流式输出模拟（Fake 模式和 Batch 模式）
- ✅ 结构化 JSON 日志
- ✅ 响应式设计（自动适应各种屏幕尺寸）
- ✅ 透明卡片 + 毛玻璃效果
- ✅ 动态动漫背景

## 目录结构

```
chatjimmy2api/
├── api/                          # Go 后端代码
│   ├── main.go                   # 独立模式入口
│   ├── vercel.go                 # Vercel Serverless 入口
│   ├── go.mod                    # Go 模块定义
│   ├── go.sum                    # Go 依赖锁定
│   ├── Makefile                  # 构建脚本
│   ├── start.sh                  # 快速启动脚本
│   ├── config/
│   │   ├── config.go             # 配置管理器（支持热重载）
│   │   └── config_test.go        # 配置测试
│   ├── internal/
│   │   ├── client/               # 上游 HTTP 客户端
│   │   ├── handler/              # API 和管理界面处理器
│   │   │   ├── api_handler.go    # OpenAI API 处理器
│   │   │   ├── admin_handler.go  # 管理界面处理器
│   │   │   └── websocket.go      # WebSocket 支持
│   │   ├── logger/               # 日志系统
│   │   ├── metrics/              # 统计指标管理
│   │   ├── stream/               # 流式模拟器
│   │   ├── transform/            # 请求/响应格式转换
│   │   └── types/                # 类型定义
│   ├── data/                     # 统计数据持久化目录
│   ├── logs/                     # 日志文件目录
│   └── examples/                 # 使用示例
├── public/                       # 前端静态文件（MD3 管理界面）
│   ├── index.html                # 登录页面
│   ├── dashboard.html            # 仪表盘（实时监控）
│   ├── config.html               # 配置管理
│   ├── stats.html                # 统计分析
│   └── logs.html                 # 系统日志
├── vercel.json                   # Vercel 部署配置
└── README.md                     # 项目说明文档
```

## 构建与运行

### 本地开发（独立模式）

```bash
# 进入 api 目录
cd api

# 下载依赖
go mod download

# 运行（使用默认配置）
go run main.go

# 或指定配置文件
go run main.go -config config/config.json
```

### 编译

```bash
cd api

# 当前平台编译
make build

# 交叉编译
make build-linux    # Linux amd64
make build-macos    # macOS arm64
make build-windows  # Windows amd64

# 清理
make clean
```

### 运行测试

```bash
cd api

# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
make test-coverage
```

### Vercel 部署

```bash
# 安装 Vercel CLI
npm i -g vercel

# 登录
vercel login

# 部署
vercel --prod
```

## 配置说明

### 配置文件位置

- 默认：`api/config/config.json`
- 可通过 `-config` 参数或 `CONFIG_PATH` 环境变量指定

### 完整配置示例

```json
{
  "upstream_base_url": "https://chatjimmy.ai",
  "upstream_api_key": "",
  "upstream_timeout_ms": 15000,
  "upstream_max_retries": 2,
  "upstream_prefill_token_limit": 6064,
  "upstream_request_byte_limit": 1200000,
  "experimental_tool_usage": false,
  "host": "127.0.0.1",
  "port": 8787,
  "default_stream": false,
  "wrapper_api_key": "local-wrapper-key",
  "body_limit_mb": 25,
  "stream_mode": "fake",
  "fake_stream_delay_ms": 50,
  "batch_stream_size": 100,
  "admin_enabled": true,
  "admin_host": "127.0.0.1",
  "admin_port": 8788,
  "log_file": "logs/server.log",
  "stats_file": "data/stats.json",
  "stats_flush_interval_sec": 30
}
```

### 关键配置项

| 配置项 | 说明 | 默认值 | 有效范围 |
|--------|------|--------|----------|
| `upstream_base_url` | 上游服务地址 | `https://chatjimmy.ai` | - |
| `wrapper_api_key` | 本地 API 认证密钥 | `local-wrapper-key` | - |
| `port` | API 服务端口 | `8787` | 1-65535 |
| `admin_port` | 管理界面端口 | `8788` | 1-65535 |
| `stream_mode` | 流式模式 | `fake` | `fake` / `batch` |
| `upstream_timeout_ms` | 上游超时时间 | `15000` | 1000-300000 |
| `body_limit_mb` | 请求体大小限制 | `25` | 1-512 |

### 配置热重载

修改 `api/config/config.json` 后自动生效，无需重启服务。配置管理器使用 `fsnotify` 监控文件变化。

## API 端点

### OpenAI 兼容端点（端口 8787）

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/models` | GET | ✅ | 获取模型列表 |
| `/v1/chat/completions` | POST | ✅ | 聊天补全 |
| `/health` | GET | ❌ | 健康检查 |

### Web 管理界面（端口 8788 / Vercel 同端口）

| 页面 | 路由 | 说明 |
|------|------|------|
| 登录页 | `/` 或 `/login` | 密码认证页面 |
| 仪表盘 | `/dashboard` | 实时监控、统计卡片 |
| 配置管理 | `/config` | 可视化配置编辑 |
| 统计分析 | `/stats` | 模型分布、错误统计 |
| 系统日志 | `/logs` | 日志查看、搜索、下载 |

### 管理 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/admin/login` | POST | 登录认证 |
| `/api/admin/logout` | POST | 登出 |
| `/api/config` | GET/POST | 获取/更新配置 |
| `/api/stats` | GET | 获取统计数据 |
| `/api/logs` | GET | 获取日志列表 |
| `/health` | GET | 健康检查 |

### 认证方式

- **API 请求**: `Authorization: Bearer <wrapper_api_key>`
- **管理界面**: `Authorization: Bearer <admin_token>` 或 Cookie

## 前端界面特性

### 设计规范
- **Material Design 3**: 遵循 Google 最新设计语言
- **透明卡片**: 毛玻璃效果（backdrop-filter: blur）
- **动态背景**: 来自 `https://www.loliapi.com/acg/` 的动漫图片
- **响应式布局**: 自动适应桌面、平板、手机

### 颜色主题
- 主色：`#6750A4` (紫色)
- 背景：半透明深色渐变
- 卡片：半透明表面色 + 模糊效果

### 交互特性
- 平滑动画过渡
- 侧边导航抽屉（移动端）
- 实时数据刷新（30 秒间隔）
- Token 持久化（LocalStorage）

## 使用示例

### cURL

```bash
# 获取模型列表
curl -sS \
  -H "Authorization: Bearer local-wrapper-key" \
  "http://127.0.0.1:8787/v1/models"

# 非流式聊天
curl -sS \
  -H "Authorization: Bearer local-wrapper-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","messages":[{"role":"user","content":"Hello!"}]}' \
  "http://127.0.0.1:8787/v1/chat/completions"

# 流式聊天
curl -sS -N \
  -H "Authorization: Bearer local-wrapper-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","stream":true,"messages":[{"role":"user","content":"Hello"}]}' \
  "http://127.0.0.1:8787/v1/chat/completions"
```

### Node.js

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

### Python

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8787/v1",
    api_key="local-wrapper-key"
)

response = client.chat.completions.create(
    model="llama3.1-8B",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 开发约定

### 代码风格

- 遵循 Go 官方代码规范
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 进行代码检查

### 日志格式

采用 JSON 格式的结构化日志：
```json
{"time":"2024-01-01T12:00:00Z","level":"INFO","message":"服务器启动","fields":{"version":"1.0.0"}}
```

### 错误处理

- 所有错误应记录到日志系统
- API 错误应返回标准格式响应
- 配置验证失败应阻止服务启动

### 测试实践

- 配置包有单元测试 (`config_test.go`)
- 运行测试使用 `go test ./...`
- 覆盖率报告使用 `make test-coverage`

## 注意事项

1. **安全**：
   - 生产环境必须修改 `wrapper_api_key` 为强密码
   - 管理界面 Token 存储在 LocalStorage，注意 XSS 防护

2. **Vercel 限制**：
   - Serverless 函数最大执行时间 60 秒
   - `/tmp` 目录数据重启后清除

3. **文件权限**：
   - 确保 `logs/` 和 `data/` 目录有写入权限

4. **流式模式**：
   - **Fake 模式**：按词分割，模拟打字机效果（默认）
   - **Batch 模式**：按字符数分割，适合大段文本

5. **背景图片**：
   - 使用 `https://www.loliapi.com/acg/` 提供的随机动漫图片
   - 每次页面加载可能显示不同图片

## 相关文档

- 主 README: `README.md`
- Vercel 配置：`vercel.json`
- 配置管理器：`api/config/config.go`
- API 处理器：`api/internal/handler/api_handler.go`
- 统计管理：`api/internal/metrics/metrics.go`
