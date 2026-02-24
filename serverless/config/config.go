package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Config 应用配置结构
type Config struct {
	mu sync.RWMutex

	// 上游服务配置
	UpstreamBaseURL           string `json:"upstream_base_url"`
	UpstreamAPIKey            string `json:"upstream_api_key"`
	UpstreamTimeoutMs         int    `json:"upstream_timeout_ms"`
	UpstreamMaxRetries        int    `json:"upstream_max_retries"`
	UpstreamPrefillTokenLimit int    `json:"upstream_prefill_token_limit"`
	UpstreamRequestByteLimit  int    `json:"upstream_request_byte_limit"`

	// 实验性功能
	ExperimentalToolUsage bool `json:"experimental_tool_usage"`

	// 本地服务配置
	Host            string `json:"host"`
	Port            int    `json:"port"`
	DefaultStream   bool   `json:"default_stream"`
	WrapperAPIKey   string `json:"wrapper_api_key"`
	BodyLimitMB     int    `json:"body_limit_mb"`

	// 流式输出配置
	StreamMode         string `json:"stream_mode"`          // "fake" 或 "batch"
	FakeStreamDelayMs  int    `json:"fake_stream_delay_ms"` // 伪造流式的延迟（毫秒）
	BatchStreamSize    int    `json:"batch_stream_size"`    // 批量流式的大小（字符数）

	// Web 管理界面配置
	AdminEnabled bool   `json:"admin_enabled"`
	AdminHost    string `json:"admin_host"`
	AdminPort    int    `json:"admin_port"`

	// 日志和统计配置
	LogFile           string `json:"log_file"`
	StatsFile         string `json:"stats_file"`
	StatsFlushIntervalSec int `json:"stats_flush_interval_sec"`
}

// Manager 配置管理器，支持热重载
type Manager struct {
	config  *Config
	path    string
	watcher *fsnotify.Watcher
	callbacks []func(*Config)
	mu      sync.RWMutex
}

// NewDefaultManager 创建默认配置管理器（不加载文件，不监控）
// 用于 Vercel Serverless 环境配置加载失败时的回退
func NewDefaultManager() *Manager {
	return &Manager{
		config:    DefaultConfig(),
		path:      "",
		callbacks: make([]func(*Config), 0),
		watcher:   nil,
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		UpstreamBaseURL:           "https://chatjimmy.ai",
		UpstreamAPIKey:            "",
		UpstreamTimeoutMs:         15000,
		UpstreamMaxRetries:        2,
		UpstreamPrefillTokenLimit: 6064,
		UpstreamRequestByteLimit:  1200000,
		ExperimentalToolUsage:     false,
		Host:                      "127.0.0.1",
		Port:                      8787,
		DefaultStream:             false,
		WrapperAPIKey:             "local-wrapper-key",
		BodyLimitMB:               25,
		StreamMode:                "fake",
		FakeStreamDelayMs:         50,
		BatchStreamSize:           100,
		AdminEnabled:              true,
		AdminHost:                 "127.0.0.1",
		AdminPort:                 8788,
		LogFile:                   "logs/server.log",
		StatsFile:                 "data/stats.json",
		StatsFlushIntervalSec:     30,
	}
}

// NewManager 创建配置管理器
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		config:    DefaultConfig(),
		path:      configPath,
		callbacks: make([]func(*Config), 0),
	}

	// 如果配置文件存在，加载它
	if _, err := os.Stat(configPath); err == nil {
		if err := m.load(); err != nil {
			return nil, fmt.Errorf("加载配置文件失败：%w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("检查配置文件失败：%w", err)
	}

	// 创建文件监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监视器失败：%w", err)
	}
	m.watcher = watcher

	// 监视配置文件
	dir := filepath.Dir(configPath)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("添加目录监视失败：%w", err)
	}

	// 启动热重载协程
	go m.watchLoop()

	return m, nil
}

// load 从文件加载配置
func (m *Manager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	newConfig := DefaultConfig()
	if err := json.Unmarshal(data, newConfig); err != nil {
		return fmt.Errorf("解析 JSON 失败：%w", err)
	}

	// 验证配置
	if err := m.validate(newConfig); err != nil {
		return fmt.Errorf("配置验证失败：%w", err)
	}

	m.mu.Lock()
	m.config = newConfig
	m.mu.Unlock()

	return nil
}

// validate 验证配置值
func (m *Manager) validate(c *Config) error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port 必须在 1-65535 之间")
	}
	if c.UpstreamTimeoutMs < 1000 || c.UpstreamTimeoutMs > 300000 {
		return fmt.Errorf("upstream_timeout_ms 必须在 1000-300000 之间")
	}
	if c.UpstreamMaxRetries < 0 || c.UpstreamMaxRetries > 10 {
		return fmt.Errorf("upstream_max_retries 必须在 0-10 之间")
	}
	if c.UpstreamPrefillTokenLimit < 256 || c.UpstreamPrefillTokenLimit > 131072 {
		return fmt.Errorf("upstream_prefill_token_limit 必须在 256-131072 之间")
	}
	if c.UpstreamRequestByteLimit < 16384 || c.UpstreamRequestByteLimit > 104857600 {
		return fmt.Errorf("upstream_request_byte_limit 必须在 16384-104857600 之间")
	}
	if c.BodyLimitMB < 1 || c.BodyLimitMB > 512 {
		return fmt.Errorf("body_limit_mb 必须在 1-512 之间")
	}
	if c.StreamMode != "fake" && c.StreamMode != "batch" {
		return fmt.Errorf("stream_mode 必须是 fake 或 batch")
	}
	if c.FakeStreamDelayMs < 10 || c.FakeStreamDelayMs > 1000 {
		return fmt.Errorf("fake_stream_delay_ms 必须在 10-1000 之间")
	}
	if c.BatchStreamSize < 10 || c.BatchStreamSize > 10000 {
		return fmt.Errorf("batch_stream_size 必须在 10-10000 之间")
	}
	return nil
}

// watchLoop 监视配置文件变化
func (m *Manager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			// 配置文件被写入或创建时重新加载
			if event.Name == m.path && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				time.Sleep(100 * time.Millisecond) // 等待写入完成
				if err := m.load(); err != nil {
					fmt.Printf("热重载配置失败：%v\n", err)
				} else {
					fmt.Println("配置已热重载")
					m.notifyCallbacks()
				}
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("配置文件监视错误：%v\n", err)
		}
	}
}

// Get 获取当前配置（线程安全）
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// 返回配置的深拷贝
	data, _ := json.Marshal(m.config)
	newConfig := &Config{}
	json.Unmarshal(data, newConfig)
	return newConfig
}

// Update 更新配置并保存到文件
func (m *Manager) Update(updateFn func(*Config)) error {
	m.mu.Lock()
	updateFn(m.config)
	m.mu.Unlock()

	// 验证配置
	if err := m.validate(m.config); err != nil {
		return err
	}

	// 保存到文件
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return err
	}

	m.notifyCallbacks()
	return nil
}

// OnChange 注册配置变更回调
func (m *Manager) OnChange(fn func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, fn)
}

// notifyCallbacks 通知所有回调
func (m *Manager) notifyCallbacks() {
	m.mu.RLock()
	config := m.config
	callbacks := make([]func(*Config), len(m.callbacks))
	copy(callbacks, m.callbacks)
	m.mu.RUnlock()

	for _, fn := range callbacks {
		fn(config)
	}
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}

// Save 保存当前配置到文件
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}
