package transform

import (
	"encoding/json"
	"strings"

	"github.com/taalas/chatjimmy2api/api/_internal/types"
)

// UpstreamChatOptions 上游聊天选项
type UpstreamChatOptions struct {
	SelectedModel string `json:"selectedModel"`
	SystemPrompt  string `json:"systemPrompt"`
	TopK          int    `json:"topK"`
}

// UpstreamChatRequest 上游聊天请求
type UpstreamChatRequest struct {
	Messages  []json.RawMessage   `json:"messages"`
	ChatOptions UpstreamChatOptions `json:"chatOptions"`
	Tools       []types.Tool        `json:"tools,omitempty"`
	ToolChoice  interface{}         `json:"toolChoice,omitempty"`
}

// UpstreamRequestMeta 上游请求元数据
type UpstreamRequestMeta struct {
	OriginalEstimatedTokens int `json:"original_estimated_tokens"`
	FinalEstimatedTokens    int `json:"final_estimated_tokens"`
	DroppedMessageCount     int `json:"dropped_message_count"`
	TruncatedChars          int `json:"truncated_chars"`
}

// BuildUpstreamChatRequest 将 OpenAI 请求转换为上游格式
func BuildUpstreamChatRequest(req types.ChatCompletionRequest, maxInputTokens int) (*UpstreamChatRequest, UpstreamRequestMeta) {
	meta := UpstreamRequestMeta{}

	// 提取系统提示
	var systemPrompt string
	var userMessages []types.Message

	for _, msg := range req.Messages {
		if msg.Role == types.RoleSystem {
			if str, ok := msg.Content.(string); ok {
				systemPrompt = str
			}
		} else {
			userMessages = append(userMessages, msg)
		}
	}

	// 估算 token 数（粗略估算：4 字符≈1 token）
	estimatedTokens := 0
	messagesJSON, _ := json.Marshal(userMessages)
	estimatedTokens = len(messagesJSON) / 4
	meta.OriginalEstimatedTokens = estimatedTokens

	// 如果超过限制，删除最早的消息
	for estimatedTokens > maxInputTokens && len(userMessages) > 1 {
		userMessages = userMessages[1:]
		meta.DroppedMessageCount++
		messagesJSON, _ = json.Marshal(userMessages)
		estimatedTokens = len(messagesJSON) / 4
	}

	// 如果还是超过限制，截断最后一条消息
	if estimatedTokens > maxInputTokens && len(userMessages) > 0 {
		lastMsg := &userMessages[len(userMessages)-1]
		if str, ok := lastMsg.Content.(string); ok {
			maxChars := maxInputTokens * 4
			if len(str) > maxChars {
				lastMsg.Content = str[:maxChars]
				meta.TruncatedChars = len(str) - maxChars
			}
		}
		estimatedTokens = maxInputTokens
	}
	meta.FinalEstimatedTokens = estimatedTokens

	// 转换消息格式
	upstreamMessages := make([]json.RawMessage, 0, len(userMessages))
	for _, msg := range userMessages {
		// 标准化 role 字段
		role := normalizeRole(msg.Role)
		simpleMsg := map[string]interface{}{
			"role":    role,
			"content": msg.Content,
		}
		msgData, _ := json.Marshal(simpleMsg)
		upstreamMessages = append(upstreamMessages, msgData)
	}

	// 构建上游请求
	upstreamReq := &UpstreamChatRequest{
		Messages: upstreamMessages,
		ChatOptions: UpstreamChatOptions{
			SelectedModel: req.Model,
			SystemPrompt:  systemPrompt,
			TopK:          1,
		},
	}

	// 如果启用了工具，添加工具信息
	if req.Tools != nil {
		upstreamReq.Tools = req.Tools
		upstreamReq.ToolChoice = req.ToolChoice
	}

	return upstreamReq, meta
}

// StatsBlock 上游响应中的统计块
type StatsBlock struct {
	PrefillTokens  int `json:"prefill_tokens,omitempty"`
	PromptTokens   int `json:"prompt_tokens,omitempty"`
	DecodeTokens   int `json:"decode_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	DoneReason     string `json:"done_reason,omitempty"`
	FinishReason   string `json:"finish_reason,omitempty"`
}

// ParseStatsFromText 从文本中提取统计信息
func ParseStatsFromText(raw string) (cleanedText string, prompt, completion int, finishReason string, hasStats bool) {
	cleanedText = raw
	prompt = 0
	completion = 0
	finishReason = "stop"
	hasStats = false

	// 查找 <|stats|>...</|stats|> 块
	statsStart := strings.LastIndex(raw, "<|stats|>")
	if statsStart == -1 {
		return
	}

	statsEnd := strings.Index(raw[statsStart:], "<|/stats|>")
	if statsEnd == -1 {
		return
	}
	statsEnd += statsStart

	statsJSON := raw[statsStart+len("<|stats|>") : statsEnd]
	var stats StatsBlock
	if err := json.Unmarshal([]byte(statsJSON), &stats); err != nil {
		return
	}

	hasStats = true

	// 提取 token 数
	if stats.PrefillTokens > 0 {
		prompt = stats.PrefillTokens
	} else if stats.PromptTokens > 0 {
		prompt = stats.PromptTokens
	}

	if stats.DecodeTokens > 0 {
		completion = stats.DecodeTokens
	} else if stats.CompletionTokens > 0 {
		completion = stats.CompletionTokens
	}

	// 提取完成原因
	if stats.DoneReason != "" {
		finishReason = stats.DoneReason
	} else if stats.FinishReason != "" {
		finishReason = stats.FinishReason
	}

	// 移除统计块
	cleanedText = strings.ReplaceAll(raw, raw[statsStart:statsEnd+len("<|/stats|>")], "")

	return
}
