# ChatJimmy2API

高性能 Go 语言实现的 OpenAI 兼容 API 包装器，用于代理 https://chatjimmy.ai 服务。

## 特性

- 高性能 Go 语言实现，低内存占用，高并发支持
- OpenAI 兼容 API 端点 (/v1/models, /v1/chat/completions)
- 配置热重载，修改配置文件自动生效
- 实时统计（请求次数、Token 消耗、错误统计）
- Web 管理界面（Material Design 3 风格）
- 伪流式输出，兼容所有客户端和 Vercel Serverless
- 结构化 JSON 日志
- 响应式设计，自动适应各种屏幕尺寸

## 快速部署到 Vercel

### 方法一：Fork 本项目（推荐）

1. **Fork 仓库**
   - 访问 https://github.com/jwbb903/ChatJimmy2API
   - 点击右上角的 "Fork" 按钮
   - 等待 Fork 完成

2. **在 Vercel 导入**
   - 访问 https://vercel.com/new
   - 点击 "Import Git Repository"
   - 找到你 Fork 的 `ChatJimmy2API` 仓库
   - 点击 "Import"

3. **配置环境变量**

在 Vercel 项目设置中，添加以下环境变量：

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `API_KEY` | API 认证密钥 | `my-api-key-123` |
| `ADMIN_PASSWORD` | 管理界面登录密码 | `my-admin-password` |
| `DISABLE_ADMIN_API` | 禁用管理界面（可选） | `true` |
| `CONFIG_PATH` | 配置文件路径 | `/var/task/conf/config.json` |

4. **配置构建选项**

在 Vercel 项目设置的 "Build & Development Settings" 中：

- **Framework Preset**: Other
- **Build Command**: `cp api/go.mod ./go.mod && cp api/go.sum ./go.sum`
- **Output Directory**: 留空

5. **部署**

点击 "Deploy" 开始部署。

### 方法二：手动上传代码

1. **创建 GitHub 仓库**
   - 访问 https://github.com/new
   - 仓库名：`chatjimmy2api`
   - 点击 "Create repository"

2. **推送代码**
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/YOUR_USERNAME/chatjimmy2api.git
   git push -u origin main
   ```

3. **在 Vercel 部署**
   - 访问 https://vercel.com/new
   - 导入你的仓库
   - 按照方法一的步骤 3-5 配置

## 本地开发

### 编译

```bash
cd api
go build -o chatjimmy2api .
```

### 运行

```bash
./chatjimmy2api -config config/config.json
```

服务启动后：
- API 服务：http://127.0.0.1:8787
- 管理界面：http://127.0.0.1:8788

### 配置

编辑 `api/config/config.json`：

```json
{
  "upstream_base_url": "https://chatjimmy.ai",
  "upstream_api_key": "",
  "upstream_timeout_ms": 15000,
  "upstream_max_retries": 2,
  "host": "127.0.0.1",
  "port": 8787,
  "wrapper_api_key": "local-wrapper-key",
  "body_limit_mb": 4,
  "stream_mode": "fake",
  "admin_enabled": true,
  "admin_port": 8788,
  "log_file": "logs/server.log",
  "stats_file": "data/stats.json"
}
```

## API 端点

### OpenAI 兼容端点

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/models` | GET | API_KEY | 获取模型列表 |
| `/v1/chat/completions` | POST | API_KEY | 聊天补全 |
| `/health` | GET | 无 | 健康检查 |

### 管理界面端点

| 端点 | 说明 |
|------|------|
| `/` | 登录页面 |
| `/dashboard` | 仪表盘（实时监控） |
| `/config` | 配置管理 |
| `/stats` | 统计分析 |
| `/logs` | 系统日志 |

### 管理 API

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/login` | POST | 无 | 登录 |
| `/api/admin/logout` | POST | ADMIN_PASSWORD | 登出 |
| `/api/config` | GET/POST | ADMIN_PASSWORD | 获取/更新配置 |
| `/api/stats` | GET | ADMIN_PASSWORD | 获取统计 |
| `/api/logs` | GET | ADMIN_PASSWORD | 获取日志 |

## 使用示例

### cURL

获取模型列表：
```bash
curl -sS \
  -H "Authorization: Bearer YOUR_API_KEY" \
  "https://your-project.vercel.app/v1/models"
```

非流式聊天：
```bash
curl -sS \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","messages":[{"role":"user","content":"Hello!"}]}' \
  "https://your-project.vercel.app/v1/chat/completions"
```

伪流式聊天：
```bash
curl -sS -N \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","stream":true,"messages":[{"role":"user","content":"Hello"}]}' \
  "https://your-project.vercel.app/v1/chat/completions"
```

### Node.js

```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'https://your-project.vercel.app/v1',
  apiKey: 'YOUR_API_KEY',
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
    base_url="https://your-project.vercel.app/v1",
    api_key="YOUR_API_KEY"
)

response = client.chat.completions.create(
    model="llama3.1-8B",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 环境变量说明

### API_KEY

用于 OpenAI 兼容 API 端点的认证（`/v1/*`）。

客户端请求时需要携带：
```
Authorization: Bearer YOUR_API_KEY
```

### ADMIN_PASSWORD

用于管理界面的登录认证。

在管理界面登录时使用此密码。

### DISABLE_ADMIN_API

设置为 `true` 或 `1` 时，完全禁用管理界面 API。

适用场景：
- 只使用 OpenAI API 端点
- 减少攻击面
- 节省 Vercel 函数执行时间

### CONFIG_PATH

配置文件路径，Vercel 环境下应设置为：
```
/var/task/conf/config.json
```

### 注意事项

**Vercel 环境下的数据存储：**

在 Vercel Serverless 环境下，统计数据仅存储在内存中，不会持久化到文件。这是因为：
- Vercel 的 `/tmp` 目录在函数重启后会被清除
- 每次函数调用都是独立的执行环境

因此，在 Vercel 上：
- 统计数据在每次函数执行时重置
- 适合查看当前请求的统计
- 如需持久化统计，建议使用外部数据库（如 Redis、Supabase 等）

## 注意事项

1. **安全建议**
   - 生产环境必须修改默认密码
   - 使用强密码作为 API_KEY 和 ADMIN_PASSWORD
   - 考虑设置 DISABLE_ADMIN_API=true 禁用管理界面

2. **Vercel 限制**
   - Serverless 函数最大执行时间 300 秒
   - 请求体大小限制 4.5 MB
   - /tmp 目录数据重启后清除

3. **伪流式说明**
   - 由于 Vercel Serverless 不支持 http.Flusher，使用伪流式模式
   - 一次性生成所有 SSE 格式数据并返回
   - 兼容所有支持 SSE 的客户端

## 目录结构

```
.
├── api/                    # Go 后端代码
│   ├── vercel.go          # Vercel 入口函数
│   ├── _internal/         # 内部包（Vercel 不扫描）
│   │   ├── client/        # 上游 HTTP 客户端
│   │   ├── config/        # 配置管理
│   │   ├── handler/       # API 处理器
│   │   ├── logger/        # 日志系统
│   │   ├── metrics/       # 统计管理
│   │   ├── stream/        # 流式模拟
│   │   ├── transform/     # 格式转换
│   │   └── types/         # 类型定义
│   ├── go.mod             # Go 模块定义
│   └── go.sum             # 依赖锁定
├── public/                # 前端静态文件
│   ├── index.html         # 登录页面
│   ├── dashboard.html     # 仪表盘
│   ├── config.html        # 配置管理
│   ├── stats.html         # 统计分析
│   └── logs.html          # 系统日志
├── conf/                  # 配置文件
│   └── config.json        # 配置示例
├── vercel.json            # Vercel 部署配置
└── README.md              # 项目说明
```

## 许可证

MIT License
