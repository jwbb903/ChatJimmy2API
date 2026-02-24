package transform

import (
	"testing"

	"github.com/jwbb903/ChatJimmy2API/api/_internal/types"
)

func TestMakeCompletionID(t *testing.T) {
	id := MakeCompletionID()

	// 检查前缀
	if len(id) < 10 {
		t.Fatalf("ID 太短：%s", id)
	}
	if id[:8] != "chatcmpl" {
		t.Errorf("ID 前缀错误：%s", id[:8])
	}
}

func TestNormalizeFinishReason(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"stop", "stop"},
		{"tool_calls", "tool_calls"},
		{"function_call", "tool_calls"},
		{"length", "length"},
		{"content_filter", "content_filter"},
		{"unknown", "stop"},
		{"", "stop"},
	}

	for _, tt := range tests {
		result := NormalizeFinishReason(tt.input)
		if result != tt.expected {
			t.Errorf("输入 %q: 期望 %q, 得到 %q", tt.input, tt.expected, result)
		}
	}
}

func TestComputeUsage(t *testing.T) {
	usage := ComputeUsage(100, 50)

	if usage.PromptTokens != 100 {
		t.Errorf("Prompt tokens 错误：%d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("Completion tokens 错误：%d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("Total tokens 错误：%d", usage.TotalTokens)
	}
}

func TestComputeUsageNegative(t *testing.T) {
	usage := ComputeUsage(-10, -5)

	if usage.PromptTokens != 0 {
		t.Errorf("负值应转为 0: %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 0 {
		t.Errorf("负值应转为 0: %d", usage.CompletionTokens)
	}
}

func TestBuildChatCompletionResponse(t *testing.T) {
	usage := types.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	resp := BuildChatCompletionResponse("llama3.1-8B", "Hello!", nil, "stop", usage)

	if resp.Object != "chat.completion" {
		t.Errorf("Object 错误：%s", resp.Object)
	}
	if resp.Model != "llama3.1-8B" {
		t.Errorf("Model 错误：%s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("期望 1 个选择，得到 %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != types.RoleAssistant {
		t.Errorf("角色错误：%s", resp.Choices[0].Message.Role)
	}
	if resp.Choices[0].Message.Content != "Hello!" {
		t.Errorf("内容错误：%s", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("完成原因错误：%s", resp.Choices[0].FinishReason)
	}
	if resp.Usage.PromptTokens != 100 {
		t.Errorf("Usage 错误：%v", resp.Usage)
	}
}

func TestBuildChatCompletionResponseWithTools(t *testing.T) {
	toolCalls := []types.ToolCall{
		{
			ID:   "call_123",
			Type: "function",
		},
	}

	resp := BuildChatCompletionResponse("llama3.1-8B", "", toolCalls, "stop", types.Usage{})

	// 有工具调用时 content 应为 nil
	if resp.Choices[0].Message.Content != nil {
		t.Errorf("有工具调用时 content 应为 nil: %v", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("完成原因应为 tool_calls: %s", resp.Choices[0].FinishReason)
	}
	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Errorf("工具调用数量错误：%d", len(resp.Choices[0].Message.ToolCalls))
	}
}

func TestBuildChatCompletionChunk(t *testing.T) {
	chunk := BuildChatCompletionChunk(
		"chatcmpl_123",
		1234567890,
		"llama3.1-8B",
		types.Message{Content: "Hello"},
		nil,
		nil,
	)

	if chunk.Object != "chat.completion.chunk" {
		t.Errorf("Object 错误：%s", chunk.Object)
	}
	if chunk.ID != "chatcmpl_123" {
		t.Errorf("ID 错误：%s", chunk.ID)
	}
	if chunk.Model != "llama3.1-8B" {
		t.Errorf("Model 错误：%s", chunk.Model)
	}
	if len(chunk.Choices) != 1 {
		t.Fatalf("期望 1 个选择，得到 %d", len(chunk.Choices))
	}
	if chunk.Choices[0].Delta.Content != "Hello" {
		t.Errorf("内容错误：%s", chunk.Choices[0].Delta.Content)
	}
}

func TestBuildChatCompletionChunkWithFinish(t *testing.T) {
	finishReason := "stop"
	chunk := BuildChatCompletionChunk(
		"chatcmpl_123",
		1234567890,
		"llama3.1-8B",
		types.Message{},
		&finishReason,
		nil,
	)

	if chunk.Choices[0].FinishReason == nil {
		t.Fatal("FinishReason 应为非 nil")
	}
	if *chunk.Choices[0].FinishReason != "stop" {
		t.Errorf("FinishReason 错误：%s", *chunk.Choices[0].FinishReason)
	}
}

func TestBuildChatCompletionChunkWithUsage(t *testing.T) {
	usage := &types.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	chunk := BuildChatCompletionChunk(
		"chatcmpl_123",
		1234567890,
		"llama3.1-8B",
		types.Message{},
		nil,
		usage,
	)

	if chunk.Usage == nil {
		t.Fatal("Usage 应为非 nil")
	}
	if chunk.Usage.PromptTokens != 100 {
		t.Errorf("Usage 错误：%v", chunk.Usage)
	}
}

func TestBuildErrorResponse(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{401, "authentication_error"},
		{403, "permission_error"},
		{404, "not_found_error"},
		{429, "rate_limit_error"},
		{500, "server_error"},
		{502, "server_error"},
		{400, "invalid_request_error"},
	}

	for _, tt := range tests {
		status, resp := BuildErrorResponse(tt.status, "test message")
		if status != tt.status {
			t.Errorf("状态码错误：%d", status)
		}
		if resp.Error.Type != tt.expected {
			t.Errorf("类型错误：%s (status=%d)", resp.Error.Type, tt.status)
		}
		if resp.Error.Message != "test message" {
			t.Errorf("消息错误：%s", resp.Error.Message)
		}
	}
}

func TestBuildModelsResponse(t *testing.T) {
	models := []string{"llama3.1-8B", "llama3.1-70B"}
	resp := BuildModelsResponse(models)

	if resp.Object != "list" {
		t.Errorf("Object 错误：%s", resp.Object)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("期望 2 个模型，得到 %d", len(resp.Data))
	}
	if resp.Data[0].ID != "llama3.1-8B" {
		t.Errorf("第一个模型 ID 错误：%s", resp.Data[0].ID)
	}
	if resp.Data[1].ID != "llama3.1-70B" {
		t.Errorf("第二个模型 ID 错误：%s", resp.Data[1].ID)
	}
}

func TestParseStatsFromText(t *testing.T) {
	input := "Hello<|stats|>{\"prefill_tokens\":100,\"decode_tokens\":50,\"done_reason\":\"stop\"}<|/stats|>World"
	cleaned, prompt, completion, finishReason, hasStats := ParseStatsFromText(input)

	if !hasStats {
		t.Fatal("应检测到 stats")
	}
	if cleaned != "HelloWorld" {
		t.Errorf("清理后文本错误：%q", cleaned)
	}
	if prompt != 100 {
		t.Errorf("Prompt tokens 错误：%d", prompt)
	}
	if completion != 50 {
		t.Errorf("Completion tokens 错误：%d", completion)
	}
	if finishReason != "stop" {
		t.Errorf("完成原因错误：%s", finishReason)
	}
}

func TestParseStatsFromTextNoStats(t *testing.T) {
	input := "Hello World"
	cleaned, prompt, completion, finishReason, hasStats := ParseStatsFromText(input)

	if hasStats {
		t.Error("不应检测到 stats")
	}
	if cleaned != "Hello World" {
		t.Errorf("清理后文本错误：%q", cleaned)
	}
	if prompt != 0 {
		t.Errorf("Prompt tokens 应为 0: %d", prompt)
	}
	if completion != 0 {
		t.Errorf("Completion tokens 应为 0: %d", completion)
	}
	if finishReason != "stop" {
		t.Errorf("默认完成原因错误：%s", finishReason)
	}
}

func TestBuildUpstreamChatRequest(t *testing.T) {
	req := types.ChatCompletionRequest{
		Model: "llama3.1-8B",
		Messages: []types.Message{
			{Role: types.RoleSystem, Content: "You are helpful."},
			{Role: types.RoleUser, Content: "Hello!"},
		},
	}

	upstreamReq, meta := BuildUpstreamChatRequest(req, 1000)

	if len(upstreamReq.Messages) != 1 {
		t.Errorf("应只有 1 条用户消息，得到 %d", len(upstreamReq.Messages))
	}
	if upstreamReq.ChatOptions.SystemPrompt != "You are helpful." {
		t.Errorf("系统提示错误：%s", upstreamReq.ChatOptions.SystemPrompt)
	}
	if upstreamReq.ChatOptions.SelectedModel != "llama3.1-8B" {
		t.Errorf("模型错误：%s", upstreamReq.ChatOptions.SelectedModel)
	}
	if meta.OriginalEstimatedTokens <= 0 {
		t.Errorf("原始 token 估算应大于 0: %d", meta.OriginalEstimatedTokens)
	}
}

func TestBuildUpstreamChatRequestTruncate(t *testing.T) {
	// 创建长消息
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "Hello World! "
	}

	req := types.ChatCompletionRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: longContent},
		},
	}

	// 设置很小的 token 限制
	_, meta := BuildUpstreamChatRequest(req, 100)

	if meta.TruncatedChars == 0 && meta.DroppedMessageCount == 0 {
		t.Error("应有限制操作（截断或删除消息）")
	}
}
