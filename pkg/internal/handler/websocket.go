package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/taalas/chatjimmy2api/pkg/metrics"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	conn *websocket.Conn
	send chan []byte
}

// WebSocketManager WebSocket 管理器
type WebSocketManager struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan []byte
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	metrics    *metrics.Manager
	mu         sync.RWMutex
}

// NewWebSocketManager 创建 WebSocket 管理器
func NewWebSocketManager(metricsMgr *metrics.Manager) *WebSocketManager {
	m := &WebSocketManager{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		metrics:    metricsMgr,
	}

	go m.run()
	return m
}

// run 运行 WebSocket 管理器
func (m *WebSocketManager) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
			m.mu.Unlock()

		case message := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.mu.RUnlock()

		case <-ticker.C:
			m.broadcastStats()
		}
	}
}

// broadcastStats 广播统计信息
func (m *WebSocketManager) broadcastStats() {
	stats := m.metrics.GetStats()
	uptime := m.metrics.GetUptime()

	data := map[string]interface{}{
		"type": "stats",
		"data": map[string]interface{}{
			"total_requests":       stats.TotalRequests,
			"success_requests":     stats.SuccessRequests,
			"failed_requests":      stats.FailedRequests,
			"total_tokens":         stats.TotalTokens,
			"prompt_tokens":        stats.TotalPromptTokens,
			"completion_tokens":    stats.TotalCompletionTokens,
			"stream_requests":      stats.StreamRequests,
			"non_stream_requests":  stats.NonStreamRequests,
			"model_requests":       stats.ModelRequests,
			"uptime_seconds":       uptime.Seconds(),
			"requests_per_minute":  m.metrics.GetRequestsPerMinute(),
			"avg_tokens_per_request": m.metrics.GetAvgTokensPerRequest(),
			"timestamp":            time.Now().Unix(),
		},
	}

	// 简化处理，实际应该使用 json.Marshal 并广播
	_ = data
}

// Register 注册客户端
func (m *WebSocketManager) Register(client *WebSocketClient) {
	m.register <- client
}

// Unregister 注销客户端
func (m *WebSocketManager) Unregister(client *WebSocketClient) {
	m.unregister <- client
}

// Broadcast 广播消息
func (m *WebSocketManager) Broadcast(message []byte) {
	m.broadcast <- message
}
