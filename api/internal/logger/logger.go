package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Entry 日志条目
type Entry struct {
	Time      time.Time         `json:"time"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Caller    string            `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger 日志记录器
type Logger struct {
	mu         sync.Mutex
	level      Level
	file       *os.File
	filePath   string
	maxSize    int64
	maxBackups int
	writers    []io.Writer
	buffer     []*Entry
	maxBuffer  int
}

// Config 日志配置
type Config struct {
	Level      Level
	FilePath   string
	MaxSize    int64 // 最大文件大小（字节）
	MaxBackups int   // 最大备份数
	MaxBuffer  int   // 内存缓冲条目数
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		FilePath:   "logs/server.log",
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxBackups: 3,
		MaxBuffer:  1000,
	}
}

// New 创建日志记录器
func New(config Config) (*Logger, error) {
	l := &Logger{
		level:      config.Level,
		filePath:   config.FilePath,
		maxSize:    config.MaxSize,
		maxBackups: config.MaxBackups,
		buffer:     make([]*Entry, 0, config.MaxBuffer),
		maxBuffer:  config.MaxBuffer,
		writers:    []io.Writer{os.Stdout},
	}

	// 如果未指定文件路径，只输出到 stdout
	if config.FilePath == "" {
		return l, nil
	}

	// 确保目录存在
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败：%w", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败：%w", err)
	}
	l.file = file
	l.writers = append(l.writers, file)

	return l, nil
}

// rotate 检查并轮换日志文件
func (l *Logger) rotate() error {
	if l.file == nil {
		return nil
	}

	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil
	}

	// 关闭当前文件
	l.file.Close()

	// 轮换备份文件
	for i := l.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.filePath, i)
		newPath := fmt.Sprintf("%s.%d", l.filePath, i+1)
		os.Rename(oldPath, newPath)
	}

	// 移动当前文件到 .1
	newPath := fmt.Sprintf("%s.1", l.filePath)
	os.Rename(l.filePath, newPath)

	// 创建新文件
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.file = file

	return nil
}

// log 内部日志方法
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	entry := &Entry{
		Time:    time.Now(),
		Level:   level.String(),
		Message: msg,
		Caller:  caller,
		Fields:  fields,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 写入缓冲区
	l.buffer = append(l.buffer, entry)
	if len(l.buffer) > l.maxBuffer {
		l.buffer = l.buffer[1:]
	}

	// 写入文件
	data, _ := json.Marshal(entry)
	for _, w := range l.writers {
		w.Write(data)
		w.Write([]byte("\n"))
	}

	// 检查轮换
	l.rotate()
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelDebug, msg, f)
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelInfo, msg, f)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelWarn, msg, f)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelError, msg, f)
}

// GetRecentLogs 获取最近的日志（从缓冲区）
func (l *Logger) GetRecentLogs(limit int) []*Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.buffer) {
		limit = len(l.buffer)
	}

	start := len(l.buffer) - limit
	if start < 0 {
		start = 0
	}

	// 返回深拷贝
	result := make([]*Entry, limit)
	for i := 0; i < limit; i++ {
		data, _ := json.Marshal(l.buffer[start+i])
		var entry Entry
		json.Unmarshal(data, &entry)
		result[i] = &entry
	}

	return result
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// GetStats 获取日志统计
func (l *Logger) GetStats() map[string]interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	info, err := l.file.Stat()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"file_path":    l.filePath,
		"file_size":    info.Size(),
		"buffer_size":  len(l.buffer),
		"max_buffer":   l.maxBuffer,
	}
}
