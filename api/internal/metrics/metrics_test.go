package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 1)
	defer mgr.Close()

	stats := mgr.GetStats()
	if stats.TotalRequests != 0 {
		t.Errorf("初始请求数应为 0: %d", stats.TotalRequests)
	}
	if stats.StartTime.IsZero() {
		t.Error("启动时间不应为零")
	}
}

func TestRecordRequest(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	// 记录成功请求
	mgr.RecordRequest("llama3.1-8B", false, 100, 50, true, "")

	stats := mgr.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("总请求数应为 1: %d", stats.TotalRequests)
	}
	if stats.SuccessRequests != 1 {
		t.Errorf("成功请求数应为 1: %d", stats.SuccessRequests)
	}
	if stats.TotalTokens != 150 {
		t.Errorf("总 token 数应为 150: %d", stats.TotalTokens)
	}
	if stats.ModelRequests["llama3.1-8B"] != 1 {
		t.Errorf("模型请求数应为 1: %d", stats.ModelRequests["llama3.1-8B"])
	}
}

func TestRecordRequestFailed(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	// 记录失败请求
	mgr.RecordRequest("llama3.1-8B", true, 0, 0, false, "502")

	stats := mgr.GetStats()
	if stats.FailedRequests != 1 {
		t.Errorf("失败请求数应为 1: %d", stats.FailedRequests)
	}
	if stats.ErrorCounts["502"] != 1 {
		t.Errorf("502 错误计数应为 1: %d", stats.ErrorCounts["502"])
	}
	if stats.StreamRequests != 1 {
		t.Errorf("流式请求数应为 1: %d", stats.StreamRequests)
	}
}

func TestGetUptime(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	time.Sleep(100 * time.Millisecond)

	uptime := mgr.GetUptime()
	if uptime < 100*time.Millisecond {
		t.Errorf("运行时间应至少 100ms: %v", uptime)
	}
}

func TestGetRequestsPerMinute(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	// 初始应为 0
	rpm := mgr.GetRequestsPerMinute()
	if rpm != 0 {
		t.Errorf("初始 RPM 应为 0: %f", rpm)
	}

	// 记录一些请求
	for i := 0; i < 10; i++ {
		mgr.RecordRequest("test", false, 10, 5, true, "")
	}

	// 由于时间很短，RPM 应该很大
	rpm = mgr.GetRequestsPerMinute()
	if rpm <= 0 {
		t.Errorf("RPM 应大于 0: %f", rpm)
	}
}

func TestGetAvgTokensPerRequest(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	// 初始应为 0
	avg := mgr.GetAvgTokensPerRequest()
	if avg != 0 {
		t.Errorf("初始平均 token 应为 0: %f", avg)
	}

	// 记录请求
	mgr.RecordRequest("test", false, 100, 50, true, "")
	mgr.RecordRequest("test", false, 200, 100, true, "")

	avg = mgr.GetAvgTokensPerRequest()
	expected := 450.0 / 2.0 // (150 + 300) / 2
	if avg != expected {
		t.Errorf("平均 token 应为 %f, 得到 %f", expected, avg)
	}
}

func TestReset(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	// 记录一些数据
	mgr.RecordRequest("test", false, 100, 50, true, "")
	mgr.RecordRequest("test", false, 0, 0, false, "500")

	// 重置
	mgr.Reset()

	stats := mgr.GetStats()
	if stats.TotalRequests != 0 {
		t.Errorf("重置后总请求应为 0: %d", stats.TotalRequests)
	}
	if stats.TotalTokens != 0 {
		t.Errorf("重置后总 token 应为 0: %d", stats.TotalTokens)
	}
	if stats.StartTime.IsZero() {
		t.Error("启动时间应保留")
	}
}

func TestExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	mgr.RecordRequest("test", false, 100, 50, true, "")

	data, err := mgr.ExportJSON()
	if err != nil {
		t.Fatalf("导出 JSON 失败：%v", err)
	}

	if len(data) == 0 {
		t.Error("导出的 JSON 不应为空")
	}

	// 验证包含关键字段
	jsonStr := string(data)
	if !contains(jsonStr, "total_requests") {
		t.Error("JSON 应包含 total_requests")
	}
	if !contains(jsonStr, "total_tokens") {
		t.Error("JSON 应包含 total_tokens")
	}
}

func TestFlush(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "stats.json")

	mgr := NewManager(statsFile, 1) // 1 秒刷新间隔
	defer mgr.Close()

	mgr.RecordRequest("test", false, 100, 50, true, "")

	// 等待刷新
	time.Sleep(1500 * time.Millisecond)

	// 验证文件已创建
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Error("统计文件应已创建")
	}

	// 创建新管理器并验证数据已加载
	mgr2 := NewManager(statsFile, 60)
	defer mgr2.Close()

	stats := mgr2.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("应从文件加载总请求数： %d", stats.TotalRequests)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	statsFile := filepath.Join(tmpDir, "nonexistent.json")

	mgr := NewManager(statsFile, 60)
	defer mgr.Close()

	stats := mgr.GetStats()
	if stats.TotalRequests != 0 {
		t.Errorf("不存在的文件应使用默认值： %d", stats.TotalRequests)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
