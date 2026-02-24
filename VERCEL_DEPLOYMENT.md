# Vercel 部署指南

## ✅ Vercel 兼容性检查

### 1. 目录结构 ✅

```
chatjimmy2api/
├── api/                  # ✅ Go 函数目录（Vercel 要求）
│   ├── vercel.go         # ✅ Vercel 入口函数
│   ├── main.go           # ✅ 独立模式入口
│   ├── internal/         # ✅ 内部包
│   ├── config/           # ✅ 配置文件
│   └── go.mod            # ✅ Go 模块定义
├── public/               # ✅ 静态文件目录
│   ├── index.html        # 登录页面
│   ├── dashboard.html    # 仪表盘
│   ├── config.html       # 配置管理
│   ├── stats.html        # 统计分析
│   └── logs.html         # 系统日志
└── vercel.json           # ✅ Vercel 部署配置
```

### 2. vercel.json 配置 ✅

已修复的配置：

```json
{
  "version": 2,
  "buildCommand": "cp api/go.mod ./go.mod && cp api/go.sum ./go.sum",
  "outputDirectory": "public",
  "functions": {
    "api/*.go": {
      "maxDuration": 300
    }
  },
  "routes": [
    {
      "src": "/v1/(.*)",
      "dest": "/api/vercel.go"
    },
    {
      "src": "/health",
      "dest": "/api/vercel.go"
    },
    {
      "src": "/api/(.*)",
      "dest": "/api/vercel.go"
    },
    {
      "src": "/login",
      "dest": "/public/index.html"
    },
    {
      "src": "/dashboard",
      "dest": "/public/dashboard.html"
    },
    {
      "src": "/config",
      "dest": "/public/config.html"
    },
    {
      "src": "/stats",
      "dest": "/public/stats.html"
    },
    {
      "src": "/logs",
      "dest": "/public/logs.html"
    },
    {
      "src": "/",
      "dest": "/public/index.html"
    },
    {
      "src": "/(.*)",
      "dest": "/public/$1"
    }
  ]
}
```

**关键修复：**
- ✅ `buildCommand`: 复制 `go.mod` 和 `go.sum` 到根目录（Vercel 需要）
- ✅ `maxDuration`: 从 60 秒增加到 300 秒（5 分钟），适配 AI 响应时间
- ✅ 路由规则：正确映射 API 和静态文件

### 3. Vercel 限制合规性 ✅

| 限制项 | Vercel 限制 | 项目配置 | 状态 |
|--------|------------|---------|------|
| **执行时长** | 300 秒（Hobby） | 300 秒 | ✅ 符合 |
| **请求体大小** | 4.5 MB | 4 MB | ✅ 符合 |
| **上游请求体** | 4.5 MB | 4 MB (4000000 字节) | ✅ 符合 |
| **内存** | 2 GB | 默认 | ✅ 符合 |
| **文件描述符** | 1024 | 使用连接池 | ✅ 符合 |

### 4. 配置文件修复 ✅

`api/config/config.json` 已更新：

```json
{
  "body_limit_mb": 4,                    // ✅ 从 25 改为 4，符合 Vercel 限制
  "upstream_request_byte_limit": 4000000, // ✅ 从 1200000 改为 4000000
  "log_file": "/tmp/logs/server.log",    // ✅ Vercel 环境使用 /tmp
  "stats_file": "/tmp/data/stats.json"   // ✅ Vercel 环境使用 /tmp
}
```

### 5. Vercel 环境适配 ✅

`api/vercel.go` 已更新：

```go
// 自动检测 Vercel 环境
if os.Getenv("VERCEL") == "1" {
    configPath = "/var/task/config/config.json"
} else {
    configPath = "config/config.json"
}

// 日志和统计使用 /tmp 目录
logCfg.FilePath = "/tmp/logs/server.log"
metricsMgr := metrics.NewManager("/tmp/data/stats.json", ...)
```

### 6. 代码修复 ✅

| 文件 | 修复内容 |
|------|---------|
| `config/config.go` | 添加 `NewDefaultManager()` 函数，支持无配置文件模式 |
| `config/config.go` | `Close()` 方法支持 `watcher` 为 `nil` |
| `logger/logger.go` | `New()` 函数支持空文件路径（只输出到 stdout） |
| `metrics/metrics.go` | `GetStats()` 返回指针避免锁复制 |
| `handler/admin_handler.go` | 使用 `c.Data()` 避免锁复制警告 |
| `public/dashboard.html` | 修复 API 路径 `/health` → `/api/health` |
| `public/dashboard.html` | 添加运行时间计算和显示 |

---

## 🚀 部署步骤

### 1. 准备工作

```bash
# 安装 Vercel CLI
npm i -g vercel

# 登录 Vercel
vercel login
```

### 2. 配置项目

编辑 `api/config/config.json`：

```json
{
  "upstream_base_url": "https://chatjimmy.ai",
  "wrapper_api_key": "your-secure-password",  // ⚠️ 修改为强密码
  "port": 8787,
  "admin_port": 8788
}
```

### 3. 部署到 Vercel

```bash
# 进入项目根目录
cd /root/nixian/chatjimmy2api

# 部署（预览）
vercel

# 部署到生产环境
vercel --prod
```

### 4. 环境变量（可选）

在 Vercel 项目设置中添加环境变量：

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `CONFIG_PATH` | 配置文件路径 | `/var/task/config/config.json` |
| `UPSTREAM_API_KEY` | 上游 API 密钥 | `your-upstream-key` |

---

## ⚠️ 注意事项

### Vercel 限制

1. **执行时长**：最大 300 秒（5 分钟），超时返回 `504 FUNCTION_INVOCATION_TIMEOUT`
2. **请求体大小**：最大 4.5 MB，超过返回 `413 FUNCTION_PAYLOAD_TOO_LARGE`
3. **冷启动**：首次请求可能有 1-3 秒延迟
4. **数据持久化**：`/tmp` 目录数据在函数重启后清除

### 配置建议

1. **修改默认密码**：生产环境必须修改 `wrapper_api_key`
2. **监控用量**：Vercel 按 CPU 时间和内存计费
3. **连接池**：上游 HTTP 客户端应使用连接池，避免文件描述符泄漏

### 前端修复

`dashboard.html` 已修复：
- ✅ API 路径从 `/health` 改为 `/api/health`
- ✅ 添加认证头 `Authorization: Bearer <token>`
- ✅ 运行时间从统计 API 计算显示

---

## 🧪 验证部署

部署后访问：

```
# 管理界面
https://your-project.vercel.app

# API 端点
https://your-project.vercel.app/v1/models
https://your-project.vercel.app/v1/chat/completions

# 健康检查
https://your-project.vercel.app/health
```

### 测试命令

```bash
# 获取模型列表
curl -sS \
  -H "Authorization: Bearer your-wrapper-api-key" \
  "https://your-project.vercel.app/v1/models"

# 聊天测试
curl -sS \
  -H "Authorization: Bearer your-wrapper-api-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.1-8B","messages":[{"role":"user","content":"Hello!"}]}' \
  "https://your-project.vercel.app/v1/chat/completions"
```

---

## 📊 性能优化建议

1. **启用流式模式**：减少单次请求时长
   ```json
   {
     "stream_mode": "batch",
     "batch_stream_size": 100
   }
   ```

2. **调整超时**：根据上游服务响应时间调整
   ```json
   {
     "upstream_timeout_ms": 15000
   }
   ```

3. **监控指标**：定期检查统计和日志
   - `/api/stats` - 统计信息
   - `/api/logs` - 系统日志

---

## 🆘 故障排查

### 常见问题

**1. 504 Gateway Timeout**
- 原因：上游服务响应超时
- 解决：增加 `upstream_timeout_ms` 或优化上游服务

**2. 413 Payload Too Large**
- 原因：请求体超过 4.5 MB
- 解决：减小 `body_limit_mb` 配置

**3. 配置加载失败**
- 原因：配置文件路径错误
- 解决：检查 `CONFIG_PATH` 环境变量或配置文件位置

**4. 日志/统计丢失**
- 原因：Vercel 重启后 `/tmp` 数据清除
- 解决：这是正常行为，数据会在下次请求时重新创建

---

## ✅ 检查清单

部署前确认：

- [ ] `vercel.json` 配置正确
- [ ] `api/config/config.json` 中的 `body_limit_mb` ≤ 4
- [ ] `wrapper_api_key` 已修改为强密码
- [ ] 日志和统计路径使用 `/tmp`
- [ ] 代码编译通过（`go build`）
- [ ] 本地测试通过

部署后验证：

- [ ] 管理界面可访问
- [ ] API 端点可访问
- [ ] 登录功能正常
- [ ] 统计数据显示正常
- [ ] 聊天功能正常
