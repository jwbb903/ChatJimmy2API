package transform

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taalas/chatjimmy2api/api/_internal/types"
)

// MakeCompletionID 生成 OpenAI 风格的完成 ID
func MakeCompletionID() string {
	return "chatcmpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// NormalizeFinishReason 将上游完成原因转换为 OpenAI 兼容值
func NormalizeFinishReason(raw string) string {
	switch raw {
	case "tool_calls", "function_call":
		return "tool_calls"
	case "length":
		return "length"
	case "content_filter":
		return "content_filter"
	default:
		return "stop"
	}
}

// ComputeUsage 计算使用量统计
func ComputeUsage(prompt, completion int) types.Usage {
	promptTokens := prompt
	if prompt < 0 {
		promptTokens = 0
	}
	completionTokens := completion
	if completion < 0 {
		completionTokens = 0
	}
	return types.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

// BuildChatCompletionResponse 构建非流式聊天补全响应
func BuildChatCompletionResponse(model string, content string, toolCalls []types.ToolCall, finishReason string, usage types.Usage) types.ChatCompletionResponse {
	hasToolCalls := len(toolCalls) > 0
	normalizedReason := NormalizeFinishReason(finishReason)
	if hasToolCalls {
		normalizedReason = "tool_calls"
	}

	resp := types.ChatCompletionResponse{
		ID:      MakeCompletionID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []struct {
			Index        int           `json:"index"`
			Message      types.Message `json:"message"`
			FinishReason string        `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: types.Message{
					Role:      types.RoleAssistant,
					Content:   getContentValue(hasToolCalls, content),
					ToolCalls: toolCalls,
				},
				FinishReason: normalizedReason,
			},
		},
		Usage: usage,
	}

	return resp
}

// getContentValue 根据是否有工具调用返回 content 值
func getContentValue(hasToolCalls bool, content string) interface{} {
	if hasToolCalls {
		return nil
	}
	return content
}

// BuildChatCompletionChunk 构建流式聊天补全块
func BuildChatCompletionChunk(id string, created int64, model string, delta types.Message, finishReason *string, usage *types.Usage) types.ChatCompletionChunk {
	// 标准化 delta 的 role，确保不会发送空值或无效的 role
	normalizedDelta := delta
	if normalizedDelta.Role == "" {
		normalizedDelta.Role = types.RoleAssistant
	} else {
		normalizedDelta.Role = normalizeRole(normalizedDelta.Role)
	}

	// 如果 content 为空字符串或 nil，设置为 nil 以便 JSON 序列化时省略
	if normalizedDelta.Content == "" {
		normalizedDelta.Content = nil
	}

	chunk := types.ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []struct {
			Index        int           `json:"index"`
			Delta        types.Message `json:"delta"`
			FinishReason *string       `json:"finish_reason,omitempty"`
		}{
			{
				Index:        0,
				Delta:        normalizedDelta,
				FinishReason: finishReason,
			},
		},
	}

	if usage != nil {
		chunk.Usage = usage
	}

	return chunk
}

// BuildErrorResponse 构建 OpenAI 风格错误响应
func BuildErrorResponse(status int, message string) (int, types.ErrorResponse) {
	errType := "invalid_request_error"
	switch status {
	case 401:
		errType = "authentication_error"
	case 403:
		errType = "permission_error"
	case 404:
		errType = "not_found_error"
	case 429:
		errType = "rate_limit_error"
	case 500, 502, 503, 504:
		errType = "server_error"
	}

	resp := types.ErrorResponse{}
	resp.Error.Message = message
	resp.Error.Type = errType
	resp.Error.Code = status

	return status, resp
}

// BuildModelsResponse 构建模型列表响应
func BuildModelsResponse(models []string) types.ModelsResponse {
	modelInfos := make([]types.ModelInfo, 0, len(models))
	for _, model := range models {
		modelInfos = append(modelInfos, types.ModelInfo{
			ID:      model,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "taalas",
		})
	}

	return types.ModelsResponse{
		Object: "list",
		Data:   modelInfos,
	}
}
