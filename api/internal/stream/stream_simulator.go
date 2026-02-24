package stream

import (
	"strings"
	"time"
)

// StreamMode 流式模式
type StreamMode string

const (
	// StreamModeFake 伪造流式：按时间延迟分块发送
	StreamModeFake StreamMode = "fake"
	// StreamModeBatch 批量流式：按固定大小分块发送
	StreamModeBatch StreamMode = "batch"
)

// FakeStreamConfig 伪造流式配置
type FakeStreamConfig struct {
	DelayMs int // 每块之间的延迟（毫秒）
	MinChunkSize int // 最小块大小（字符数）
	MaxChunkSize int // 最大块大小（字符数）
}

// BatchStreamConfig 批量流式配置
type BatchStreamConfig struct {
	ChunkSize int // 每块的大小（字符数）
}

// StreamSimulator 流式模拟器
type StreamSimulator struct {
	mode StreamMode
	fakeConfig FakeStreamConfig
	batchConfig BatchStreamConfig
}

// NewStreamSimulator 创建新的流式模拟器
func NewStreamSimulator(mode StreamMode, fakeDelayMs, batchSize int) *StreamSimulator {
	return &StreamSimulator{
		mode: mode,
		fakeConfig: FakeStreamConfig{
			DelayMs: fakeDelayMs,
			MinChunkSize: 1,
			MaxChunkSize: 10,
		},
		batchConfig: BatchStreamConfig{
			ChunkSize: batchSize,
		},
	}
}

// UpdateConfig 更新配置
func (s *StreamSimulator) UpdateConfig(mode StreamMode, fakeDelayMs, batchSize int) {
	s.mode = mode
	s.fakeConfig.DelayMs = fakeDelayMs
	s.batchConfig.ChunkSize = batchSize
}

// StreamResult 流式输出结果
type StreamResult struct {
	Chunk string
	Done  bool
}

// StreamContent 将完整内容转换为流式输出通道
func (s *StreamSimulator) StreamContent(content string, doneChan chan<- StreamResult) {
	defer close(doneChan)

	switch s.mode {
	case StreamModeFake:
		s.streamFake(content, doneChan)
	case StreamModeBatch:
		s.streamBatch(content, doneChan)
	default:
		// 默认使用批量模式
		s.streamBatch(content, doneChan)
	}
}

// streamFake 伪造流式：按时间延迟逐字符/词发送
func (s *StreamSimulator) streamFake(content string, doneChan chan<- StreamResult) {
	// 按空格分割成词
	words := strings.Fields(content)
	
	for i, word := range words {
		// 添加空格（如果不是第一个词）
		chunk := word
		if i > 0 {
			chunk = " " + chunk
		}
		
		doneChan <- StreamResult{
			Chunk: chunk,
			Done:  false,
		}
		
		// 延迟
		time.Sleep(time.Duration(s.fakeConfig.DelayMs) * time.Millisecond)
	}
	
	doneChan <- StreamResult{
		Chunk: "",
		Done:  true,
	}
}

// streamBatch 批量流式：按固定大小分块发送
func (s *StreamSimulator) streamBatch(content string, doneChan chan<- StreamResult) {
	chunkSize := s.batchConfig.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 100
	}

	contentRunes := []rune(content)
	for i := 0; i < len(contentRunes); i += chunkSize {
		end := i + chunkSize
		if end > len(contentRunes) {
			end = len(contentRunes)
		}
		
		chunk := string(contentRunes[i:end])
		doneChan <- StreamResult{
			Chunk: chunk,
			Done:  false,
		}
		
		// 小延迟避免过快发送
		time.Sleep(10 * time.Millisecond)
	}
	
	doneChan <- StreamResult{
		Chunk: "",
		Done:  true,
	}
}

// SplitContentByWords 按词分割内容（用于流式输出）
func SplitContentByWords(content string) []string {
	return strings.Fields(content)
}

// SplitContentBySize 按大小分割内容（用于流式输出）
func SplitContentBySize(content string, size int) []string {
	if size <= 0 {
		size = 100
	}
	
	runes := []rune(content)
	chunks := make([]string, 0, len(runes)/size+1)
	
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	
	return chunks
}
