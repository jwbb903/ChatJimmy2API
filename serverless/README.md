# ChatJimmy2API

[![Go Version](https://img.shields.io/github/go-mod/go-version/jwbb903/chatjimmy2api)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

高性能 Go 语言实现的 OpenAI 兼容 API 包装器，代理 `https://chatjimmy.ai` 服务。

## ✨ 特性

- 🚀 **高性能** - Go 语言实现，低内存占用，高并发支持
- ⚙️ **配置热重载** - 修改配置文件自动生效，无需重启服务
- 📊 **实时统计** - 请求次数、Token 消耗、错误统计，持久化存储
- 🖥️ **Web 管理界面** - Material Design 3 风格，可视化配置管理、状态监控、日志查看
- 🔐 **安全认证** - 管理界面访问密码保护，Token 认证
- 📡 **流式输出** - 支持 Fake 和 Batch 两种流式模拟模式
- 📝 **结构化日志** - JSON 格式日志，便于日志收集和分析

## 🚀 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/jwbb903/chatjimmy2api.git
cd chatjimmy2api/go-server

# 下载依赖
go mod download

# 编译
make build

# 运行
./chatjimmy2api
```

### 默认服务地址

| 服务 | 地址 |
|------|------|
| API 服务 | http://127.0.0.1:8787 |
| Web 管理界面 | http://127.0.0.1:8788 |

### 管理界面认证

管理界面默认使用 `wrapper_api_key` 作为访问密码（默认值：`local-wrapper-key`）。

首次访问会跳转到登录页面，输入密码后即可访问。Token 会保存在本地，关闭浏览器前无需重复登录。

**修改管理密码：**

编辑 `config/config.json`，修改 `wrapper_api_key` 字段：

```json
{
  "wrapper_api_key": "your-secure-password"
}
```

保存后自动生效，无需重启服务。

### 健康检查

```bash
curl http://127.0.0.1:8787/health
# 返回：{"ok":true}
```

## 📖 使用示例

### cURL

```bash
# 获取模型列表
curl -sS \
  -H "Authorization: Bearer local-wrapper-key" \
  "http://127.0.0.1:8787/v1/models" | jq

# 非流式聊天
curl -sS \
  -H "Authorization: Bearer local-wrapper-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","messages":[{"role":"user","content":"Hello!"}]}' \
  "http://127.0.0.1:8787/v1/chat/completions" | jq

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

console.log(completion.choices[0].message.content);
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

print(response.choices[0].message.content)
```

## ⚙️ 配置说明

配置文件位于 `config/config.json`，支持热重载（修改后自动生效）。

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

| 配置项 | 说明 | 默认值 | 范围 |
|--------|------|--------|------|
| `upstream_base_url` | 上游服务地址 | `https://chatjimmy.ai` | - |
| `upstream_api_key` | 上游 API 密钥 | `""` | - |
| `wrapper_api_key` | 本地 API 认证密钥 | `local-wrapper-key` | - |
| `port` | API 服务端口 | `8787` | 1-65535 |
| `admin_port` | 管理界面端口 | `8788` | 1-65535 |
| `stream_mode` | 流式模式 | `fake` | `fake` / `batch` |
| `fake_stream_delay_ms` | 伪造流式延迟 (ms) | `50` | 10-1000 |
| `batch_stream_size` | 批量流式块大小 | `100` | 10-10000 |
| `upstream_timeout_ms` | 上游超时时间 (ms) | `15000` | 1000-300000 |

## 🔌 API 端点

### OpenAI 兼容端点 (端口 8787)

| 端点 | 方法 | 说明 |
|------|------|------|
| `/v1/models` | GET | 获取模型列表 |
| `/v1/chat/completions` | POST | 聊天补全 |
| `/health` | GET | 健康检查 |

### Web 管理界面 (端口 8788)

| 端点 | 方法 | 说明 |
|------|------|------|
| `/login` | GET | 登录页面 |
| `/` | GET | 管理首页（需认证） |
| `/config` | GET | 配置页面 |
| `/stats` | GET | 统计页面 |
| `/logs` | GET | 日志页面 |
| `/api/admin/login` | POST | 登录 API |
| `/api/admin/logout` | POST | 登出 API |
| `/api/config` | GET/POST | 配置 API |
| `/api/stats` | GET | 统计 API |
| `/api/logs` | GET | 日志 API |

**管理界面功能：**
- 📊 实时统计仪表盘
- ⚙️ 可视化配置编辑
- 📝 日志查看与过滤
- 📈 请求统计图表
- 🔐 安全认证保护

## 📊 流式模式说明

由于上游服务不支持真正的流式输出，本项目提供两种伪造流式的方式：

### Fake 模式（默认）

按词分割响应内容，每个词之间添加可配置的延迟，模拟真实的打字机效果。

```json
{
  "stream_mode": "fake",
  "fake_stream_delay_ms": 50
}
```

### Batch 模式

按固定字符数分割响应，每块快速发送，适合大段文本。

```json
{
  "stream_mode": "batch",
  "batch_stream_size": 100
}
```

## 🏗️ 项目结构

```
go-server/
├── main.go                     # 应用入口
├── go.mod                      # Go 模块定义
├── Makefile                    # 构建脚本
├── start.sh                    # 快速启动脚本
├── config/
│   ├── config.go               # 配置管理（热重载）
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

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./config/...
go test ./internal/stream/...
go test ./internal/transform/...
go test ./internal/metrics/...

# 带覆盖率
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 🔧 编译

```bash
# 当前平台
make build

# 交叉编译
make build-linux    # Linux amd64
make build-macos    # macOS arm64
make build-windows  # Windows amd64
make build-all      # 所有平台

# 手动编译
go build -o chatjimmy2api
GOOS=linux GOARCH=amd64 go build -o chatjimmy2api-linux
GOOS=darwin GOARCH=arm64 go build -o chatjimmy2api-macos
GOOS=windows GOARCH=amd64 go build -o chatjimmy2api.exe
```

## 📝 日志

日志文件位于 `logs/server.log`，采用 JSON 格式：

```json
{"time":"2024-01-01T12:00:00Z","level":"INFO","message":"服务器启动","fields":{"version":"1.0.0"}}
{"time":"2024-01-01T12:00:01Z","level":"INFO","message":"请求","fields":{"method":"POST","path":"/v1/chat/completions"}}
```

查看实时日志：

```bash
tail -f logs/server.log
# 或访问 Web 管理界面 http://127.0.0.1:8788/logs
```

## 📈 统计

统计数据持久化在 `data/stats.json`，包含：

- 总请求数、成功/失败请求数
- Token 使用统计
- 模型使用分布
- 错误类型统计
- 运行时长

## ⚠️ 注意事项

1. **配置验证**：所有配置项都有范围验证，修改时请确保值在有效范围内
2. **文件权限**：确保 `logs/` 和 `data/` 目录有写入权限
3. **生产部署**：
   - **必须修改** `wrapper_api_key` 为强密码
   - 建议使用反向代理（如 Nginx）提供 HTTPS
   - 考虑配置防火墙规则限制访问
4. **内存缓冲**：日志缓冲默认 1000 条，统计刷新间隔默认 30 秒
5. **管理界面安全**：
   - 默认密码为 `local-wrapper-key`，生产环境必须修改
   - Token 存储在浏览器本地，注意清除缓存

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
