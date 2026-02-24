package stream

import (
	"testing"
	"time"
)

func TestParseUpstreamChunk(t *testing.T) {
	// 测试不含 stats 的纯文本
	events := ParseUpstreamChunk("Hello, world!")
	if len(events) != 1 {
		t.Fatalf("期望 1 个事件，得到 %d", len(events))
	}
	if events[0].Type != "text" {
		t.Errorf("期望类型为 text，得到 %s", events[0].Type)
	}
	if events[0].Delta != "Hello, world!" {
		t.Errorf("期望文本为 Hello, world!，得到 %s", events[0].Delta)
	}
}

func TestParseUpstreamChunkWithStats(t *testing.T) {
	// 测试含 stats 的文本
	input := "Hello, world!<|stats|>{\"prefill_tokens\":100,\"decode_tokens\":50,\"done_reason\":\"stop\"}<|/stats|>"
	events := ParseUpstreamChunk(input)

	if len(events) != 2 {
		t.Fatalf("期望 2 个事件，得到 %d", len(events))
	}

	// 检查文本事件
	if events[0].Type != "text" {
		t.Errorf("第一个事件类型应为 text，得到 %s", events[0].Type)
	}
	if events[0].Delta != "Hello, world!" {
		t.Errorf("第一个事件文本错误：%s", events[0].Delta)
	}

	// 检查 done 事件
	if events[1].Type != "done" {
		t.Errorf("第二个事件类型应为 done，得到 %s", events[1].Type)
	}
	if events[1].FinishReason != "stop" {
		t.Errorf("完成原因应为 stop，得到 %s", events[1].FinishReason)
	}
	if events[1].Usage == nil {
		t.Fatal("Usage 应为非 nil")
	}
	if events[1].Usage.Prompt != 100 {
		t.Errorf("Prompt tokens 应为 100，得到 %d", events[1].Usage.Prompt)
	}
	if events[1].Usage.Completion != 50 {
		t.Errorf("Completion tokens 应为 50，得到 %d", events[1].Usage.Completion)
	}
}

func TestParseUpstreamChunkWithAltStats(t *testing.T) {
	// 测试使用替代字段的 stats
	input := "Test<|stats|>{\"prompt_tokens\":200,\"completion_tokens\":80,\"done_reason\":\"length\"}<|/stats|>"
	events := ParseUpstreamChunk(input)

	if len(events) != 2 {
		t.Fatalf("期望 2 个事件，得到 %d", len(events))
	}

	if events[1].Usage.Prompt != 200 {
		t.Errorf("Prompt tokens 应为 200，得到 %d", events[1].Usage.Prompt)
	}
	if events[1].Usage.Completion != 80 {
		t.Errorf("Completion tokens 应为 80，得到 %d", events[1].Usage.Completion)
	}
	if events[1].FinishReason != "length" {
		t.Errorf("完成原因应为 length，得到 %s", events[1].FinishReason)
	}
}

func TestUpstreamChunkParser(t *testing.T) {
	parser := NewUpstreamChunkParser()

	// 测试 SSE 格式解析
	chunk := "data: {\"type\":\"text\",\"delta\":\"Hello\"}\n\n"
	events := parser.ParseChunk(chunk)

	if len(events) != 1 {
		t.Fatalf("期望 1 个事件，得到 %d", len(events))
	}
	if events[0].Type != "text" {
		t.Errorf("期望类型为 text，得到 %s", events[0].Type)
	}
	if events[0].Delta != "Hello" {
		t.Errorf("期望文本为 Hello，得到 %s", events[0].Delta)
	}
}

func TestUpstreamChunkParserDone(t *testing.T) {
	parser := NewUpstreamChunkParser()

	chunk := "data: [DONE]\n\n"
	events := parser.ParseChunk(chunk)

	if len(events) != 1 {
		t.Fatalf("期望 1 个事件，得到 %d", len(events))
	}
	if events[0].Type != "done" {
		t.Errorf("期望类型为 done，得到 %s", events[0].Type)
	}
}

func TestUpstreamChunkParserBuffering(t *testing.T) {
	parser := NewUpstreamChunkParser()

	// 分割的块
	chunk1 := "data: {\"type\":\"text\",\"delta\":\"Hel"
	chunk2 := "lo\"}\n\n"

	events1 := parser.ParseChunk(chunk1)
	if len(events1) != 0 {
		t.Errorf("第一个块不应产生事件，得到 %d", len(events1))
	}

	events2 := parser.ParseChunk(chunk2)
	if len(events2) != 1 {
		t.Fatalf("第二个块应产生 1 个事件，得到 %d", len(events2))
	}
	if events2[0].Delta != "Hello" {
		t.Errorf("期望文本为 Hello，得到 %s", events2[0].Delta)
	}
}

func TestStreamSimulatorFake(t *testing.T) {
	sim := NewStreamSimulator(StreamModeFake, 10, 100)

	doneChan := make(chan StreamResult, 64)
	content := "Hello World Test"
	go sim.StreamContent(content, doneChan)

	var results []string
	for result := range doneChan {
		if result.Done {
			break
		}
		results = append(results, result.Chunk)
	}

	// 验证内容完整
	combined := ""
	for _, r := range results {
		combined += r
	}
	if combined != content {
		t.Errorf("期望内容 %q，得到 %q", content, combined)
	}
}

func TestStreamSimulatorBatch(t *testing.T) {
	sim := NewStreamSimulator(StreamModeBatch, 50, 10)

	doneChan := make(chan StreamResult, 64)
	content := "Hello World! This is a test message for batch streaming."
	go sim.StreamContent(content, doneChan)

	var results []string
	for result := range doneChan {
		if result.Done {
			break
		}
		results = append(results, result.Chunk)
	}

	// 验证内容完整
	combined := ""
	for _, r := range results {
		combined += r
	}
	if combined != content {
		t.Errorf("期望内容 %q，得到 %q", content, combined)
	}
}

func TestStreamSimulatorUpdateConfig(t *testing.T) {
	sim := NewStreamSimulator(StreamModeFake, 50, 100)

	sim.UpdateConfig(StreamModeBatch, 200, 50)

	if sim.mode != StreamModeBatch {
		t.Errorf("模式更新失败：%s", sim.mode)
	}
	if sim.batchConfig.ChunkSize != 50 {
		t.Errorf("批量大小更新失败：%d", sim.batchConfig.ChunkSize)
	}
}

func TestSplitContentByWords(t *testing.T) {
	content := "Hello World Test"
	words := SplitContentByWords(content)

	if len(words) != 3 {
		t.Fatalf("期望 3 个词，得到 %d", len(words))
	}
	if words[0] != "Hello" {
		t.Errorf("第一个词错误：%s", words[0])
	}
	if words[1] != "World" {
		t.Errorf("第二个词错误：%s", words[1])
	}
	if words[2] != "Test" {
		t.Errorf("第三个词错误：%s", words[2])
	}
}

func TestSplitContentBySize(t *testing.T) {
	content := "Hello World Test"
	chunks := SplitContentBySize(content, 5)

	if len(chunks) != 4 {
		t.Fatalf("期望 4 块，得到 %d", len(chunks))
	}
	if chunks[0] != "Hello" {
		t.Errorf("第一块错误：%s", chunks[0])
	}
	if chunks[1] != " Worl" {
		t.Errorf("第二块错误：%s", chunks[1])
	}
}

func TestStreamSimulatorSpeed(t *testing.T) {
	sim := NewStreamSimulator(StreamModeFake, 5, 100)

	doneChan := make(chan StreamResult, 64)
	content := "A B C D E"
	start := time.Now()
	go sim.StreamContent(content, doneChan)

	count := 0
	for result := range doneChan {
		if result.Done {
			break
		}
		count++
	}
	elapsed := time.Since(start)

	// 5 个词，每个延迟 5ms，应该至少 20ms（4 次延迟）
	if elapsed < 20*time.Millisecond {
		t.Errorf("流式速度过快：%v", elapsed)
	}

	if count != 5 {
		t.Errorf("期望 5 块，得到 %d", count)
	}
}
