package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Stats 统计信息
type Stats struct {
	mu sync.RWMutex

	// 启动时间
	StartTime time.Time `json:"start_time"`

	// 请求统计
	TotalRequests   int64 `json:"total_requests"`
	SuccessRequests int64 `json:"success_requests"`
	FailedRequests  int64 `json:"failed_requests"`

	// Token 统计
	TotalPromptTokens     int64 `json:"total_prompt_tokens"`
	TotalCompletionTokens int64 `json:"total_completion_tokens"`
	TotalTokens           int64 `json:"total_tokens"`

	// 流式统计
	StreamRequests   int64 `json:"stream_requests"`
	NonStreamRequests int64 `json:"non_stream_requests"`

	// 模型统计（按模型名）
	ModelRequests map[string]int64 `json:"model_requests"`

	// 错误统计
	ErrorCounts map[string]int64 `json:"error_counts"`

	// 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
}

// Manager 统计管理器
type Manager struct {
	stats      *Stats
	statsFile  string
	flushInterval time.Duration
	stopChan   chan struct{}
	mu         sync.RWMutex
}

// NewManager 创建统计管理器
func NewManager(statsFile string, flushIntervalSec int) *Manager {
	if flushIntervalSec <= 0 {
		flushIntervalSec = 30
	}

	m := &Manager{
		stats: &Stats{
			StartTime:     time.Now(),
			ModelRequests: make(map[string]int64),
			ErrorCounts:   make(map[string]int64),
			LastUpdated:   time.Now(),
		},
		statsFile:     statsFile,
		flushInterval: time.Duration(flushIntervalSec) * time.Second,
		stopChan:      make(chan struct{}),
	}

	// 加载已有统计
	m.load()

	// 启动定时刷新协程
	go m.flushLoop()

	return m
}

// load 从文件加载统计
func (m *Manager) load() {
	data, err := os.ReadFile(m.statsFile)
	if err != nil {
		return // 文件不存在或读取失败，使用默认值
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		return // 解析失败，使用默认值
	}

	m.mu.Lock()
	m.stats = &stats
	m.stats.LastUpdated = time.Now()
	m.mu.Unlock()
}

// flushLoop 定时刷新统计到文件
func (m *Manager) flushLoop() {
	ticker := time.NewTicker(m.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.flush()
		case <-m.stopChan:
			m.flush() // 停止前最后刷新一次
			return
		}
	}
}

// flush 将统计刷新到文件
func (m *Manager) flush() {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.stats, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return
	}

	// 确保目录存在
	dir := filepath.Dir(m.statsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	os.WriteFile(m.statsFile, data, 0644)
}

// RecordRequest 记录请求
func (m *Manager) RecordRequest(model string, isStream bool, promptTokens, completionTokens int, success bool, errorCode string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalRequests++
	m.stats.LastUpdated = time.Now()

	if success {
		m.stats.SuccessRequests++
	} else {
		m.stats.FailedRequests++
	}

	if isStream {
		m.stats.StreamRequests++
	} else {
		m.stats.NonStreamRequests++
	}

	if model != "" {
		m.stats.ModelRequests[model]++
	}

	if promptTokens > 0 {
		m.stats.TotalPromptTokens += int64(promptTokens)
	}
	if completionTokens > 0 {
		m.stats.TotalCompletionTokens += int64(completionTokens)
	}
	m.stats.TotalTokens += int64(promptTokens + completionTokens)

	if errorCode != "" {
		m.stats.ErrorCounts[errorCode]++
	}
}

// GetStats 获取统计信息
func (m *Manager) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回深拷贝
	data, _ := json.Marshal(m.stats)
	var stats Stats
	json.Unmarshal(data, &stats)
	return &stats
}

// GetUptime 获取运行时长
func (m *Manager) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.stats.StartTime)
}

// GetRequestsPerMinute 获取每分钟请求数（自启动以来平均）
func (m *Manager) GetRequestsPerMinute() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.stats.StartTime).Minutes()
	if uptime <= 0 {
		return 0
	}
	return float64(m.stats.TotalRequests) / uptime
}

// GetAvgTokensPerRequest 获取平均每请求 token 数
func (m *Manager) GetAvgTokensPerRequest() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.stats.TotalRequests <= 0 {
		return 0
	}
	return float64(m.stats.TotalTokens) / float64(m.stats.TotalRequests)
}

// Close 关闭管理器
func (m *Manager) Close() {
	close(m.stopChan)
}

// Reset 重置统计（保留启动时间）
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	startTime := m.stats.StartTime
	m.stats = &Stats{
		StartTime:     startTime,
		ModelRequests: make(map[string]int64),
		ErrorCounts:   make(map[string]int64),
		LastUpdated:   time.Now(),
	}

	// 立即刷新
	m.flush()
}

// ExportJSON 导出为 JSON
func (m *Manager) ExportJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return json.MarshalIndent(m.stats, "", "  ")
}
