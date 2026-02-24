package stream

import (
	"encoding/json"
	"strings"

	"github.com/taalas/chatjimmy2api/internal/types"
)

// ParsedUpstreamEvent 解析后的上游事件
type ParsedUpstreamEvent struct {
	Type       string         `json:"type"` // "text" 或 "tool_call"
	Delta      string         `json:"delta"`
	Call       types.ToolCall `json:"call,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
	Usage      *UsageInfo     `json:"usage,omitempty"`
}

// UsageInfo 使用量信息
type UsageInfo struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
}

// UpstreamChunkParser 上游数据流解析器
type UpstreamChunkParser struct {
	buffer string
}

// NewUpstreamChunkParser 创建新的解析器
func NewUpstreamChunkParser() *UpstreamChunkParser {
	return &UpstreamChunkParser{}
}

// ParseChunk 解析上游流式数据块
// 上游返回的是 SSE 格式：data: {...}\n\n
func (p *UpstreamChunkParser) ParseChunk(chunk string) []ParsedUpstreamEvent {
	p.buffer += chunk
	events := make([]ParsedUpstreamEvent, 0)

	// 分割 SSE 事件
	lines := strings.Split(p.buffer, "\n")
	p.buffer = ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理 SSE data: 行
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				events = append(events, ParsedUpstreamEvent{
					Type:       "done",
					FinishReason: "stop",
				})
				continue
			}

			// 尝试解析 JSON
			var event ParsedUpstreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				// 如果不是 JSON，可能是纯文本
				events = append(events, ParsedUpstreamEvent{
					Type:  "text",
					Delta: data,
				})
				continue
			}

			if event.Type == "" {
				// 如果没有 type 字段，假设是文本
				event.Type = "text"
			}
			events = append(events, event)
		} else if line != "" {
			// 保留未完成的行到缓冲区
			p.buffer = line
		}
	}

	return events
}

// Flush 刷新缓冲区，返回剩余事件
func (p *UpstreamChunkParser) Flush() []ParsedUpstreamEvent {
	if p.buffer == "" {
		return nil
	}
	return p.ParseChunk("\n")
}

// ParseUpstreamChunk 解析上游响应文本（非流式）
// 上游返回的格式可能包含 <|stats|> 块
func ParseUpstreamChunk(raw string) []ParsedUpstreamEvent {
	events := make([]ParsedUpstreamEvent, 0)

	// 查找并解析 stats 块
	statsStart := strings.LastIndex(raw, "<|stats|>")
	if statsStart != -1 {
		statsEnd := strings.Index(raw[statsStart:], "<|/stats|>")
		if statsEnd != -1 {
			statsEnd += statsStart
			statsJSON := raw[statsStart+len("<|stats|>") : statsEnd]

			// 解析 stats
			var stats struct {
				PrefillTokens    int    `json:"prefill_tokens,omitempty"`
				PromptTokens     int    `json:"prompt_tokens,omitempty"`
				DecodeTokens     int    `json:"decode_tokens,omitempty"`
				CompletionTokens int    `json:"completion_tokens,omitempty"`
				DoneReason       string `json:"done_reason,omitempty"`
			}
			if err := json.Unmarshal([]byte(statsJSON), &stats); err == nil {
				prompt := stats.PrefillTokens
				if prompt == 0 {
					prompt = stats.PromptTokens
				}
				completion := stats.DecodeTokens
				if completion == 0 {
					completion = stats.CompletionTokens
				}
				finishReason := stats.DoneReason
				if finishReason == "" {
					finishReason = "stop"
				}

				// 添加文本事件（不含 stats）
				text := raw[:statsStart] + raw[statsEnd+len("<|/stats|>"):]
				if text != "" {
					events = append(events, ParsedUpstreamEvent{
						Type:  "text",
						Delta: text,
					})
				}

				// 添加使用量事件
				events = append(events, ParsedUpstreamEvent{
					Type:       "done",
					FinishReason: finishReason,
					Usage: &UsageInfo{
						Prompt:     prompt,
						Completion: completion,
					},
				})
				return events
			}
		}
	}

	// 没有 stats 块，直接返回文本
	if raw != "" {
		events = append(events, ParsedUpstreamEvent{
			Type:  "text",
			Delta: raw,
		})
	}

	return events
}
