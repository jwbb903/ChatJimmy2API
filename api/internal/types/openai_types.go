package types

// OpenAI 消息角色
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// OpenAI 工具函数定义
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// OpenAI 工具
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// OpenAI 工具调用
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// 消息内容部分（支持多模态）
type MessageContentPart struct {
	Type     string `json:"type,omitempty"`
	Text     string `json:"text,omitempty"`
	InputText string `json:"input_text,omitempty"`
	Content  string `json:"content,omitempty"`
}

// OpenAI 消息
type Message struct {
	Role       MessageRole         `json:"role,omitempty"`
	Content    interface{}         `json:"content,omitempty"` // string 或 []MessageContentPart
	Name       string              `json:"name,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall          `json:"tool_calls,omitempty"`
}

// OpenAI 聊天补全请求
type ChatCompletionRequest struct {
	Model       string     `json:"model,omitempty"`
	Messages    []Message  `json:"messages"`
	Tools       []Tool     `json:"tools,omitempty"`
	ToolChoice  interface{} `json:"tool_choice,omitempty"` // "none", "auto", "required" 或对象
	Stream      *bool      `json:"stream,omitempty"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
	Temperature *float64   `json:"temperature,omitempty"`
	TopP        *float64   `json:"top_p,omitempty"`
	MaxTokens   *int       `json:"max_tokens,omitempty"`
}

// 流式选项
type StreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

// OpenAI 使用量统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAI 聊天补全响应（非流式）
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

// OpenAI 流式块
type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Delta        Message `json:"delta"`
		FinishReason *string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

// OpenAI 错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// 模型信息
type ModelInfo struct {
	ID      string `json:"id"`
	Created int64  `json:"created,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
	Object  string `json:"object,omitempty"`
}

// 模型列表响应
type ModelsResponse struct {
	Object string     `json:"object"`
	Data   []ModelInfo `json:"data"`
}
