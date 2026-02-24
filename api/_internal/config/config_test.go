package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.UpstreamBaseURL != "https://chatjimmy.ai" {
		t.Errorf("默认上游地址错误：%s", cfg.UpstreamBaseURL)
	}
	if cfg.Port != 8787 {
		t.Errorf("默认端口错误：%d", cfg.Port)
	}
	if cfg.StreamMode != "fake" {
		t.Errorf("默认流式模式错误：%s", cfg.StreamMode)
	}
	if cfg.FakeStreamDelayMs != 50 {
		t.Errorf("默认伪造流式延迟错误：%d", cfg.FakeStreamDelayMs)
	}
	if cfg.BatchStreamSize != 100 {
		t.Errorf("默认批量流式大小错误：%d", cfg.BatchStreamSize)
	}
}

func TestValidate(t *testing.T) {
	m := &Manager{
		config: DefaultConfig(),
	}

	// 测试有效配置
	cfg := DefaultConfig()
	if err := m.validate(cfg); err != nil {
		t.Errorf("有效配置验证失败：%v", err)
	}

	// 测试无效端口
	cfg.Port = 0
	if err := m.validate(cfg); err == nil {
		t.Error("端口 0 应该验证失败")
	}
	cfg.Port = 70000
	if err := m.validate(cfg); err == nil {
		t.Error("端口 70000 应该验证失败")
	}

	// 测试无效超时
	cfg.Port = 8787
	cfg.UpstreamTimeoutMs = 500
	if err := m.validate(cfg); err == nil {
		t.Error("超时 500ms 应该验证失败")
	}

	// 测试无效流式模式
	cfg.UpstreamTimeoutMs = 15000
	cfg.StreamMode = "invalid"
	if err := m.validate(cfg); err == nil {
		t.Error("无效流式模式应该验证失败")
	}
}

func TestNewManager(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 创建新管理器（配置文件不存在）
	m, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("创建管理器失败：%v", err)
	}
	defer m.Close()

	// 验证默认配置
	cfg := m.Get()
	if cfg.Port != 8787 {
		t.Errorf("默认端口错误：%d", cfg.Port)
	}

	// 等待文件创建
	time.Sleep(100 * time.Millisecond)

	// 验证配置文件已创建
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("配置文件未创建")
	}
}

func TestUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	m, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("创建管理器失败：%v", err)
	}
	defer m.Close()

	// 等待初始保存
	time.Sleep(100 * time.Millisecond)

	// 更新配置
	err = m.Update(func(cfg *Config) {
		cfg.Port = 9000
		cfg.FakeStreamDelayMs = 100
	})
	if err != nil {
		t.Errorf("更新配置失败：%v", err)
	}

	// 验证更新
	cfg := m.Get()
	if cfg.Port != 9000 {
		t.Errorf("端口更新失败：%d", cfg.Port)
	}
	if cfg.FakeStreamDelayMs != 100 {
		t.Errorf("流式延迟更新失败：%d", cfg.FakeStreamDelayMs)
	}

	// 验证文件已保存
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("读取配置文件失败：%v", err)
	}
	if len(data) == 0 {
		t.Error("配置文件为空")
	}
}

func TestUpdateInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	m, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("创建管理器失败：%v", err)
	}
	defer m.Close()

	time.Sleep(100 * time.Millisecond)

	// 尝试更新为无效配置
	err = m.Update(func(cfg *Config) {
		cfg.Port = 99999 // 无效端口
	})
	if err == nil {
		t.Error("更新无效配置应该失败")
	}
}

func TestOnChange(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	m, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("创建管理器失败：%v", err)
	}
	defer m.Close()

	time.Sleep(100 * time.Millisecond)

	// 注册回调
	called := false
	m.OnChange(func(cfg *Config) {
		called = true
	})

	// 更新配置
	err = m.Update(func(cfg *Config) {
		cfg.Port = 9001
	})
	if err != nil {
		t.Fatalf("更新配置失败：%v", err)
	}

	// 等待回调执行
	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("配置变更回调未被调用")
	}
}

func TestLoadExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 创建预定义配置
	cfg := DefaultConfig()
	cfg.Port = 9999
	cfg.WrapperAPIKey = "test-key"
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("创建配置文件失败：%v", err)
	}

	// 创建管理器（应加载已有配置）
	m, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("创建管理器失败：%v", err)
	}
	defer m.Close()

	// 验证配置已加载
	loadedCfg := m.Get()
	if loadedCfg.Port != 9999 {
		t.Errorf("端口未从文件加载：%d", loadedCfg.Port)
	}
	if loadedCfg.WrapperAPIKey != "test-key" {
		t.Errorf("API Key 未从文件加载：%s", loadedCfg.WrapperAPIKey)
	}
}
