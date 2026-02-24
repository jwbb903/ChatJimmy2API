package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/config"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/client"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/logger"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/metrics"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/stream"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/transform"
	"github.com/jwbb903/ChatJimmy2API/api/_internal/types"
)

// APIHandler API 处理器
type APIHandler struct {
	client      *client.ChatJimmyClient
	configMgr   *config.Manager
	metrics     *metrics.Manager
	logger      *logger.Logger
	streamSim   *stream.StreamSimulator
}

// NewAPIHandler 创建 API 处理器
func NewAPIHandler(
	cfgMgr *config.Manager,
	metricsMgr *metrics.Manager,
	log *logger.Logger,
	upstreamClient *client.ChatJimmyClient,
) *APIHandler {
	cfg := cfgMgr.Get()
	return &APIHandler{
		client:    upstreamClient,
		configMgr: cfgMgr,
		metrics:   metricsMgr,
		logger:    log,
		streamSim: stream.NewStreamSimulator(
			stream.StreamMode(cfg.StreamMode),
			cfg.FakeStreamDelayMs,
			cfg.BatchStreamSize,
		),
	}
}

// RegisterRoutes 注册 API 路由
func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	// 健康检查（不需要认证）
	router.GET("/health", h.handleHealth)

	// 需要认证的 API
	auth := router.Group("/")
	auth.Use(h.authMiddleware())
	{
		// OpenAI 兼容端点
		v1 := auth.Group("/v1")
		{
			v1.GET("/models", h.handleModels)
			v1.POST("/chat/completions", h.handleChatCompletions)
		}
	}
}

// authMiddleware API 认证中间件
func (h *APIHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := h.configMgr.Get()

		// 优先使用环境变量 ADMIN_PASSWORD（用于 Vercel）
		apiKey := os.Getenv("ADMIN_PASSWORD")
		if apiKey == "" {
			apiKey = cfg.WrapperAPIKey
		}

		// 如果未设置密钥，跳过认证
		if apiKey == "" {
			c.Next()
			return
		}

		// 检查 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing or invalid authorization header.",
					"type":    "authentication_error",
					"code":    401,
				},
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimPrefix(token, "bearer ")

		if token != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid API key.",
					"type":    "authentication_error",
					"code":    401,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleHealth 健康检查
func (h *APIHandler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// handleModels 获取模型列表
func (h *APIHandler) handleModels(c *gin.Context) {
	// 获取上游模型
	headers := h.buildForwardHeaders(c)
	resp, err := h.client.GetModels(headers)
	if err != nil {
		h.logger.Error("获取模型列表失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to fetch models from upstream.",
				"type":    "server_error",
				"code":    502,
			},
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取模型响应失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to read upstream response.",
				"type":    "server_error",
				"code":    502,
			},
		})
		return
	}

	if !respOK(resp) {
		h.logger.Warn("上游返回非 OK 状态", map[string]interface{}{
			"status": resp.Status,
			"body":   string(body),
		})
	}

	// 解析上游响应并转换为 OpenAI 格式
	var upstreamModels struct {
		Data []struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &upstreamModels); err != nil {
		// 如果解析失败，返回默认模型
		modelsResp := transform.BuildModelsResponse([]string{"llama3.1-8B"})
		c.JSON(http.StatusOK, modelsResp)
		return
	}

	// 转换模型列表
	models := make([]string, 0)
	for _, m := range upstreamModels.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		} else if m.Name != "" {
			models = append(models, m.Name)
		}
	}

	if len(models) == 0 {
		models = []string{"llama3.1-8B"}
	}

	modelsResp := transform.BuildModelsResponse(models)
	c.JSON(http.StatusOK, modelsResp)
}

// handleChatCompletions 处理聊天补全请求
func (h *APIHandler) handleChatCompletions(c *gin.Context) {
	cfg := h.configMgr.Get()

	// 解析请求体
	var req types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.metrics.RecordRequest("", false, 0, 0, false, "400")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request body: " + err.Error(),
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
		return
	}

	// 验证消息
	if len(req.Messages) == 0 {
		h.metrics.RecordRequest("", false, 0, 0, false, "400")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Messages array is required.",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
		return
	}

	// 检查工具使用
	requestedToolChoice := req.ToolChoice
	requestedToolCount := len(req.Tools)
	toolUsageRequested := requestedToolCount > 0 || requestedToolChoice != nil

	if !cfg.ExperimentalToolUsage && requestedToolChoice == "required" {
		h.metrics.RecordRequest(req.Model, false, 0, 0, false, "400")
		h.logger.Warn("工具使用被拒绝", map[string]interface{}{
			"reason": "required_requested_while_disabled",
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Tool usage is disabled (EXPERIMENTAL_TOOL_USAGE=false).",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
		return
	}

	// 确定是否使用流式
	isStream := cfg.DefaultStream
	if req.Stream != nil {
		isStream = *req.Stream
	}

	// 如果未启用工具模式，移除工具相关字段
	requestBody := req
	if !cfg.ExperimentalToolUsage && toolUsageRequested {
		requestBody.Tools = nil
		requestBody.ToolChoice = nil
		h.logger.Info("工具字段被移除", map[string]interface{}{
			"experimentalToolUsage": cfg.ExperimentalToolUsage,
		})
	}

	// 转换为上游请求
	model := req.Model
	if model == "" {
		model = "llama3.1-8B"
	}

	upstreamReq, meta := transform.BuildUpstreamChatRequest(requestBody, cfg.UpstreamPrefillTokenLimit)
	if meta.DroppedMessageCount > 0 || meta.TruncatedChars > 0 {
		h.logger.Warn("请求被裁剪以适应上游限制", map[string]interface{}{
			"droppedMessageCount": meta.DroppedMessageCount,
			"truncatedChars":      meta.TruncatedChars,
		})
	}

	// 序列化上游请求
	upstreamBody, err := json.Marshal(upstreamReq)
	if err != nil {
		h.metrics.RecordRequest(model, isStream, 0, 0, false, "500")
		h.logger.Error("序列化上游请求失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error.",
				"type":    "server_error",
				"code":    500,
			},
		})
		return
	}

	// 检查请求大小
	if len(upstreamBody) > cfg.UpstreamRequestByteLimit {
		h.metrics.RecordRequest(model, isStream, 0, 0, false, "413")
		h.logger.Warn("上游请求过大", map[string]interface{}{
			"bytes": len(upstreamBody),
			"limit": cfg.UpstreamRequestByteLimit,
		})
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": gin.H{
				"message": "Request is too large for upstream.",
				"type":    "invalid_request_error",
				"code":    413,
			},
		})
		return
	}

	// 发送上游请求
	headers := h.buildForwardHeaders(c)
	resp, err := h.client.PostChat(upstreamBody, headers)
	if err != nil {
		h.metrics.RecordRequest(model, isStream, 0, 0, false, "502")
		h.logger.Error("上游请求失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Upstream request failed: " + err.Error(),
				"type":    "server_error",
				"code":    502,
			},
		})
		return
	}
	defer resp.Body.Close()

	if !respOK(resp) {
		body, _ := io.ReadAll(resp.Body)
		h.logger.Error("上游返回错误", map[string]interface{}{
			"status": resp.Status,
			"body":   string(body),
		})
		c.JSON(resp.StatusCode, gin.H{
			"error": gin.H{
				"message": string(body),
				"type":    "server_error",
				"code":    resp.StatusCode,
			},
		})
		return
	}

	// 处理响应
	if !isStream {
		h.handleNonStream(c, resp, model, meta)
	} else {
		h.handleStream(c, resp, model, req, meta)
	}
}

// handleNonStream 处理非流式响应
func (h *APIHandler) handleNonStream(c *gin.Context, resp *http.Response, model string, meta transform.UpstreamRequestMeta) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordRequest(model, false, 0, 0, false, "502")
		h.logger.Error("读取上游响应失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to read upstream response.",
				"type":    "server_error",
				"code":    502,
			},
		})
		return
	}

	// 解析上游响应
	rawText := string(body)
	cleanedText, prompt, completion, finishReason, hasStats := transform.ParseStatsFromText(rawText)

	if cleanedText == "" && !hasStats {
		h.metrics.RecordRequest(model, false, 0, 0, false, "422")
		h.logger.Warn("上游返回空响应", map[string]interface{}{})
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"message": "Upstream returned an empty response.",
				"type":    "invalid_request_error",
				"code":    422,
			},
		})
		return
	}

	// 构建 OpenAI 响应
	usage := transform.ComputeUsage(prompt, completion)
	response := transform.BuildChatCompletionResponse(
		model,
		cleanedText,
		nil, // 工具调用
		finishReason,
		usage,
	)

	h.metrics.RecordRequest(model, false, prompt, completion, true, "")
	h.logger.Info("聊天补全完成", map[string]interface{}{
		"model":         model,
		"prompt_tokens": prompt,
		"completion_tokens": completion,
	})

	c.JSON(http.StatusOK, response)
}

// handleStream 处理流式响应（伪造）
func (h *APIHandler) handleStream(c *gin.Context, resp *http.Response, model string, req types.ChatCompletionRequest, meta transform.UpstreamRequestMeta) {
	cfg := h.configMgr.Get()

	// 读取完整响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordRequest(model, true, 0, 0, false, "502")
		h.logger.Error("读取上游流式响应失败", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to read upstream stream response.",
				"type":    "server_error",
				"code":    502,
			},
		})
		return
	}

	// 解析上游响应
	rawText := string(body)
	cleanedText, prompt, completion, finishReason, hasStats := transform.ParseStatsFromText(rawText)

	if cleanedText == "" && !hasStats {
		h.metrics.RecordRequest(model, true, 0, 0, false, "422")
		h.logger.Warn("上游返回空流式响应", map[string]interface{}{})
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"message": "Upstream returned an empty stream response.",
				"type":    "invalid_request_error",
				"code":    422,
			},
		})
		return
	}

	// 设置 SSE 头
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 生成响应 ID 和时间
	completionID := transform.MakeCompletionID()
	created := time.Now().Unix()

	// 更新流模拟器配置
	h.streamSim.UpdateConfig(
		stream.StreamMode(cfg.StreamMode),
		cfg.FakeStreamDelayMs,
		cfg.BatchStreamSize,
	)

	// 流式发送内容
	doneChan := make(chan stream.StreamResult, 64)
	go h.streamSim.StreamContent(cleanedText, doneChan)

	for result := range doneChan {
		if result.Done {
			// 发送结束块
			finishReasonOpenAI := transform.NormalizeFinishReason(finishReason)
			finalChunk := transform.BuildChatCompletionChunk(
				completionID,
				created,
				model,
				types.Message{Role: types.RoleAssistant},
				&finishReasonOpenAI,
				nil,
			)
			data, _ := json.Marshal(finalChunk)
			c.Writer.WriteString("data: " + string(data) + "\n\n")
			c.Writer.Flush()

			// 如果请求包含 usage，发送使用量
			if req.StreamOptions != nil && req.StreamOptions.IncludeUsage != nil && *req.StreamOptions.IncludeUsage {
				usage := transform.ComputeUsage(prompt, completion)
				usageChunk := transform.BuildChatCompletionChunk(
					completionID,
					created,
					model,
					types.Message{Role: types.RoleAssistant},
					nil,
					&usage,
				)
				data, _ = json.Marshal(usageChunk)
				c.Writer.WriteString("data: " + string(data) + "\n\n")
				c.Writer.Flush()
			}

			// 发送 [DONE]
			c.Writer.WriteString("data: [DONE]\n\n")
			c.Writer.Flush()
			break
		}

		// 发送内容块
		chunk := transform.BuildChatCompletionChunk(
			completionID,
			created,
			model,
			types.Message{Role: types.RoleAssistant, Content: result.Chunk},
			nil,
			nil,
		)
		data, _ := json.Marshal(chunk)
		c.Writer.WriteString("data: " + string(data) + "\n\n")
		c.Writer.Flush()
	}

	h.metrics.RecordRequest(model, true, prompt, completion, true, "")
	h.logger.Info("流式聊天补全完成", map[string]interface{}{
		"model":         model,
		"stream_mode":   cfg.StreamMode,
		"prompt_tokens": prompt,
		"completion_tokens": completion,
	})
}

// buildForwardHeaders 构建转发请求头
func (h *APIHandler) buildForwardHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)

	// 转发 Authorization 和 Cookie（如果需要）
	if auth := c.GetHeader("Authorization"); auth != "" {
		headers["Authorization"] = auth
	}
	if cookie := c.GetHeader("Cookie"); cookie != "" {
		headers["Cookie"] = cookie
	}

	return headers
}

// respOK 检查 HTTP 响应是否成功
func respOK(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
