package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taalas/chatjimmy2api/config"
	"github.com/taalas/chatjimmy2api/pkg/logger"
	"github.com/taalas/chatjimmy2api/pkg/metrics"
)

// AdminHandler 管理界面处理器
type AdminHandler struct {
	configManager *config.Manager
	metrics       *metrics.Manager
	logger        *logger.Logger
	adminPassword string
	activeTokens  map[string]bool
}

// NewAdminHandler 创建管理界面处理器
func NewAdminHandler(cfgMgr *config.Manager, metricsMgr *metrics.Manager, log *logger.Logger) *AdminHandler {
	cfg := cfgMgr.Get()
	return &AdminHandler{
		configManager: cfgMgr,
		metrics:       metricsMgr,
		logger:        log,
		adminPassword: cfg.WrapperAPIKey,
		activeTokens:  make(map[string]bool),
	}
}

// RegisterRoutes 注册管理界面路由（独立服务器模式）
func (h *AdminHandler) RegisterRoutes(router *gin.Engine) {
	// 登录页面和根路径（无需认证）
	router.GET("/login", h.handleLoginPage)
	router.GET("/", h.handleIndex)

	// 登录 API
	router.POST("/api/admin/login", h.handleLogin)
	router.POST("/api/admin/logout", h.handleLogout)

	// 需要认证的路由
	admin := router.Group("/")
	admin.Use(h.authMiddleware())
	{
		// 静态文件
		admin.StaticFS("/static", http.Dir("static"))

		// HTML 页面（带.html 后缀）
		admin.GET("/dashboard.html", h.handleDashboardPage)
		admin.GET("/config.html", h.handleConfigPage)
		admin.GET("/stats.html", h.handleStatsPage)
		admin.GET("/logs.html", h.handleLogsPage)

		// HTML 页面（不带.html 后缀）
		admin.GET("/dashboard", h.handleDashboardPage)
		admin.GET("/config", h.handleConfigPage)
		admin.GET("/stats", h.handleStatsPage)
		admin.GET("/logs", h.handleLogsPage)

		// API 端点
		api := admin.Group("/api")
		{
			// 配置 API
			api.GET("/config", h.handleGetConfig)
			api.POST("/config", h.handleUpdateConfig)

			// 统计 API
			api.GET("/stats", h.handleGetStats)
			api.POST("/stats/reset", h.handleResetStats)

			// 日志 API
			api.GET("/logs", h.handleGetLogs)
			api.GET("/logs/stats", h.handleGetLogStats)

			// 健康检查
			api.GET("/health", h.handleHealth)
		}

		// WebSocket 实时更新
		admin.GET("/ws/stats", h.handleWebSocket)
	}
}

// RegisterWebRoutes 注册 Web 路由（Vercel Serverless 模式）
func (h *AdminHandler) RegisterWebRoutes(router *gin.Engine) {
	// 登录页面
	router.GET("/login", h.handleLoginPage)

	// 登录/登出 API
	router.POST("/api/admin/login", h.handleLogin)
	router.POST("/api/admin/logout", h.handleLogout)

	// 需要认证的路由
	admin := router.Group("/")
	admin.Use(h.authMiddleware())
	{
		// HTML 页面
		admin.GET("/", h.handleIndex)
		admin.GET("/dashboard", h.handleDashboardPage)
		admin.GET("/config", h.handleConfigPage)
		admin.GET("/stats", h.handleStatsPage)
		admin.GET("/logs", h.handleLogsPage)

		// API 端点
		api := admin.Group("/api")
		{
			// 配置 API
			api.GET("/config", h.handleGetConfig)
			api.POST("/config", h.handleUpdateConfig)

			// 统计 API
			api.GET("/stats", h.handleGetStats)
			api.POST("/stats/reset", h.handleResetStats)

			// 日志 API
			api.GET("/logs", h.handleGetLogs)
			api.GET("/logs/stats", h.handleGetLogStats)

			// 健康检查
			api.GET("/health", h.handleHealth)
		}
	}
}

// authMiddleware 认证中间件
func (h *AdminHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过登录 API、登录页面
		if c.Request.URL.Path == "/login" ||
		   c.Request.URL.Path == "/api/admin/login" ||
		   c.Request.URL.Path == "/api/admin/logout" {
			c.Next()
			return
		}

		// 从 Header 获取 Authorization
		authHeader := c.GetHeader("Authorization")
		var token string
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		if token == "" {
			// 尝试从 cookie 获取
			cookie, err := c.Cookie("admin_token")
			if err != nil {
				// API 请求返回 401，页面请求重定向到登录页
				if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
					c.JSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"error":   "未授权访问",
					})
				} else {
					c.Redirect(http.StatusTemporaryRedirect, "/login")
				}
				c.Abort()
				return
			}
			token = cookie
		}

		// 验证 token
		if !h.activeTokens[token] {
			if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Token 无效",
				})
			} else {
				c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateToken 生成随机 token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// handleLoginPage 处理登录页面
func (h *AdminHandler) handleLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "登录 - ChatJimmy2API",
	})
}

// handleLogin 处理登录请求
func (h *AdminHandler) handleLogin(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求",
		})
		return
	}

	// 验证密码
	if req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "密码错误",
		})
		return
	}

	// 生成 token
	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成 token 失败",
		})
		return
	}

	// 保存 token
	h.activeTokens[token] = true

	// 设置 cookie（30 天过期）
	c.SetCookie("admin_token", token, 3600*24*30, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"token":   token,
	})
}

// handleLogout 处理登出请求
func (h *AdminHandler) handleLogout(c *gin.Context) {
	token := c.GetHeader("X-Admin-Token")
	if token != "" {
		delete(h.activeTokens, token)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// handleIndex 处理首页请求
func (h *AdminHandler) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "ChatJimmy2API - 管理界面",
	})
}

// handleDashboardPage 处理仪表盘页面
func (h *AdminHandler) handleDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "仪表盘 - ChatJimmy2API",
	})
}

// handleConfigPage 处理配置页面
func (h *AdminHandler) handleConfigPage(c *gin.Context) {
	c.HTML(http.StatusOK, "config.html", gin.H{
		"title": "配置管理",
	})
}

// handleStatsPage 处理统计页面
func (h *AdminHandler) handleStatsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "stats.html", gin.H{
		"title": "统计信息",
	})
}

// handleLogsPage 处理日志页面
func (h *AdminHandler) handleLogsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "logs.html", gin.H{
		"title": "日志查看",
	})
}

// handleGetConfig 获取当前配置（直接返回配置对象）
func (h *AdminHandler) handleGetConfig(c *gin.Context) {
	cfg := h.configManager.Get()
	c.JSON(http.StatusOK, cfg)
}

// handleUpdateConfig 更新配置
func (h *AdminHandler) handleUpdateConfig(c *gin.Context) {
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求体：" + err.Error(),
		})
		return
	}

	err := h.configManager.Update(func(cfg *config.Config) {
		// 更新配置字段
		if v, ok := input["upstream_base_url"].(string); ok {
			cfg.UpstreamBaseURL = v
		}
		if v, ok := input["upstream_api_key"].(string); ok {
			cfg.UpstreamAPIKey = v
		}
		if v, ok := input["upstream_timeout_ms"].(float64); ok {
			cfg.UpstreamTimeoutMs = int(v)
		}
		if v, ok := input["upstream_max_retries"].(float64); ok {
			cfg.UpstreamMaxRetries = int(v)
		}
		if v, ok := input["upstream_prefill_token_limit"].(float64); ok {
			cfg.UpstreamPrefillTokenLimit = int(v)
		}
		if v, ok := input["upstream_request_byte_limit"].(float64); ok {
			cfg.UpstreamRequestByteLimit = int(v)
		}
		if v, ok := input["experimental_tool_usage"].(bool); ok {
			cfg.ExperimentalToolUsage = v
		}
		if v, ok := input["host"].(string); ok {
			cfg.Host = v
		}
		if v, ok := input["port"].(float64); ok {
			cfg.Port = int(v)
		}
		if v, ok := input["default_stream"].(bool); ok {
			cfg.DefaultStream = v
		}
		if v, ok := input["wrapper_api_key"].(string); ok {
			cfg.WrapperAPIKey = v
		}
		if v, ok := input["body_limit_mb"].(float64); ok {
			cfg.BodyLimitMB = int(v)
		}
		if v, ok := input["stream_mode"].(string); ok {
			cfg.StreamMode = v
		}
		if v, ok := input["fake_stream_delay_ms"].(float64); ok {
			cfg.FakeStreamDelayMs = int(v)
		}
		if v, ok := input["batch_stream_size"].(float64); ok {
			cfg.BatchStreamSize = int(v)
		}
		if v, ok := input["admin_enabled"].(bool); ok {
			cfg.AdminEnabled = v
		}
		if v, ok := input["admin_port"].(float64); ok {
			cfg.AdminPort = int(v)
		}
		if v, ok := input["stats_flush_interval_sec"].(float64); ok {
			cfg.StatsFlushIntervalSec = int(v)
		}
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置验证失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置已更新并保存",
	})
}

// handleGetStats 获取统计信息（直接返回统计对象）
func (h *AdminHandler) handleGetStats(c *gin.Context) {
	stats := h.metrics.GetStats()
	// 转换为 JSON 字节以避免复制锁
	data, _ := json.Marshal(stats)
	c.Data(http.StatusOK, "application/json", data)
}

// handleResetStats 重置统计
func (h *AdminHandler) handleResetStats(c *gin.Context) {
	h.metrics.Reset()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "统计已重置",
	})
}

// handleGetLogs 获取日志（直接返回日志数组）
func (h *AdminHandler) handleGetLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	logs := h.logger.GetRecentLogs(limit)
	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}

// handleGetLogStats 获取日志统计
func (h *AdminHandler) handleGetLogStats(c *gin.Context) {
	stats := h.logger.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// handleHealth 健康检查
func (h *AdminHandler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

// handleWebSocket WebSocket 实时更新
func (h *AdminHandler) handleWebSocket(c *gin.Context) {
	// WebSocket 处理将在单独的文件中实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "WebSocket 暂未实现",
	})
}
