# ChatJimmy2API - Vercel Serverless 版本

高性能 Go 语言实现的 OpenAI 兼容 API 包装器，适配 Vercel Serverless 部署。

## ✨ 特性

- 🚀 **高性能** - Go 语言实现，低内存占用，高并发支持
- 🎨 **MD3 设计** - Material Design 3 风格，现代化管理界面
- 📱 **响应式布局** - 自动适应桌面、平板、手机等各种屏幕尺寸
- 🌈 **透明卡片** - 毛玻璃效果，动态动漫背景
- ⚙️ **配置热重载** - 修改配置文件自动生效，无需重启服务
- 📊 **实时统计** - 请求次数、Token 消耗、错误统计，持久化存储
- 🖥️ **Web 管理界面** - 可视化配置管理、状态监控、日志查看
- 🔐 **安全认证** - 管理界面访问密码保护，Token 认证
- 📡 **流式输出** - 支持 Fake 和 Batch 两种流式模拟模式
- 📝 **结构化日志** - JSON 格式日志，便于日志收集和分析

## 📁 目录结构

```
.
├── api/                    # Go 后端代码
│   ├── vercel.go          # Vercel 入口函数
│   ├── main.go            # standalone 模式入口
│   ├── internal/          # 内部包
│   │   ├── client/        # 上游 HTTP 客户端
│   │   ├── handler/       # API 和管理界面处理器
│   │   ├── logger/        # 日志系统
│   │   ├── metrics/       # 统计指标管理
│   │   ├── stream/        # 流式模拟器
│   │   ├── transform/     # 请求/响应格式转换
│   │   └── types/         # 类型定义
│   ├── config/            # 配置管理
│   └── config/config.json # 配置文件
├── public/                # 前端静态文件（MD3 管理界面）
│   ├── index.html         # 登录页面
│   ├── dashboard.html     # 仪表盘（实时监控）
│   ├── config.html        # 配置管理
│   ├── stats.html         # 统计分析
│   └── logs.html          # 系统日志
└── vercel.json            # Vercel 部署配置
```

## 🚀 部署到 Vercel

### 1. 配置项目

编辑 `api/config/config.json`：

```json
{
  "upstream_base_url": "https://chatjimmy.ai",
  "wrapper_api_key": "your-secure-password",
  "port": 8787,
  "admin_port": 8788
}
```

### 2. 部署

```bash
# 安装 Vercel CLI
npm i -g vercel

# 登录 Vercel
vercel login

# 部署
vercel --prod
```

部署后访问 `https://your-project.vercel.app` 即可看到管理界面。

## 💻 本地开发

```bash
# 进入 api 目录
cd api

# 安装依赖
go mod download

# 运行
go run main.go
```

服务启动后：
- **API 服务**: http://127.0.0.1:8787
- **管理界面**: http://127.0.0.1:8788

默认密码：`local-wrapper-key`（请在生产环境中修改）

## 🎨 管理界面功能

### 登录页面 (`/`)
- 密码认证
- Token 持久化（关闭浏览器前无需重复登录）
- 动态动漫背景

### 仪表盘 (`/dashboard`)
- 实时统计卡片（总请求数、成功请求、失败请求、Token 消耗）
- 服务状态监控
- 最近活动列表
- 快速刷新

### 配置管理 (`/config`)
- 上游服务配置（地址、API 密钥、超时、重试）
- 本地服务配置（端口、密钥、请求限制）
- 流式输出配置（模式选择、延迟、块大小）
- 管理界面配置
- 日志与存储配置
- 配置验证与热重载

### 统计分析 (`/stats`)
- 概览统计（成功率进度条）
- 模型使用分布（百分比展示）
- 错误统计分类
- 运行时间统计
- 时间范围筛选（全部/今天/本周）

### 系统日志 (`/logs`)
- 日志级别筛选（全部/DEBUG/INFO/SUCCESS/WARNING/ERROR）
- 实时日志搜索
- 自动滚动开关
- 分页浏览
- 日志下载
- 清空日志

## 🔌 API 端点

### OpenAI 兼容端点（端口 8787）

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/models` | GET | ✅ | 获取模型列表 |
| `/v1/chat/completions` | POST | ✅ | 聊天补全 |
| `/health` | GET | ❌ | 健康检查 |

### 管理界面端点（端口 8788）

| 端点 | 说明 |
|------|------|
| `/` | 登录页面 |
| `/dashboard` | 仪表盘 |
| `/config` | 配置管理 |
| `/stats` | 统计分析 |
| `/logs` | 系统日志 |

### 管理 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/admin/login` | POST | 登录 |
| `/api/admin/logout` | POST | 登出 |
| `/api/config` | GET/POST | 获取/更新配置 |
| `/api/stats` | GET | 获取统计 |
| `/api/logs` | GET | 获取日志 |

## ⚙️ 配置说明

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
  "stream_mode": "fake",
  "fake_stream_delay_ms": 50,
  "admin_enabled": true,
  "admin_port": 8788,
  "log_file": "logs/server.log",
  "stats_file": "data/stats.json"
}
```

### 关键配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `upstream_base_url` | 上游服务地址 | `https://chatjimmy.ai` |
| `wrapper_api_key` | 本地 API 认证密钥 | `local-wrapper-key` |
| `port` | API 服务端口 | `8787` |
| `admin_port` | 管理界面端口 | `8788` |
| `stream_mode` | 流式模式 | `fake` / `batch` |

## 🔐 安全建议

1. **修改默认密码**：生产环境必须修改 `wrapper_api_key`
2. **使用 HTTPS**：建议通过反向代理提供 HTTPS
3. **限制访问**：配置防火墙规则限制访问 IP
4. **定期更新 Token**：管理界面 Token 应定期更新

## ⚠️ 注意事项

1. **Vercel 限制**：Serverless 函数最大执行时间 60 秒
2. **数据持久化**：Vercel 环境下 `/tmp` 目录数据重启后清除
3. **日志缓冲**：日志缓冲默认 1000 条，统计刷新间隔默认 30 秒
4. **背景图片**：使用 https://www.loliapi.com/acg/ 提供的动态动漫背景

## 📸 界面预览

管理界面采用 Material Design 3 设计规范：
- 毛玻璃透明卡片效果
- 平滑动画过渡
- 响应式侧边导航
- 移动端适配
- 动态动漫背景

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
