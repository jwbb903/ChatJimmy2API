package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatJimmyClient 上游 ChatJimmy 客户端
type ChatJimmyClient struct {
	baseURL    string
	apiKey     string
	timeoutMs  int
	retries    int
	httpClient *http.Client
}

// NewChatJimmyClient 创建新的上游客户端
func NewChatJimmyClient(baseURL, apiKey string, timeoutMs, retries int) *ChatJimmyClient {
	return &ChatJimmyClient{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		apiKey:    apiKey,
		timeoutMs: timeoutMs,
		retries:   retries,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

// UpdateConfig 更新客户端配置
func (c *ChatJimmyClient) UpdateConfig(baseURL, apiKey string, timeoutMs, retries int) {
	c.baseURL = strings.TrimSuffix(baseURL, "/")
	c.apiKey = apiKey
	c.timeoutMs = timeoutMs
	c.retries = retries
	c.httpClient.Timeout = time.Duration(timeoutMs) * time.Millisecond
}

// GetModels 获取模型列表
func (c *ChatJimmyClient) GetModels(headers map[string]string) (*http.Response, error) {
	return c.request("/api/models", http.MethodGet, nil, headers)
}

// PostChat 发送聊天请求
func (c *ChatJimmyClient) PostChat(body []byte, headers map[string]string) (*http.Response, error) {
	return c.request("/api/chat", http.MethodPost, body, headers)
}

// request 执行上游请求，带重试逻辑
func (c *ChatJimmyClient) request(path string, method string, body []byte, headers map[string]string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutMs)*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败：%w", err)
		}

		// 添加请求头
		for k, v := range headers {
			if v != "" {
				req.Header.Set(k, v)
			}
		}

		// 添加上游 API Key
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		if method == http.MethodPost && body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 重试前等待
		if attempt < c.retries {
			time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
		}
	}

	return nil, lastErr
}

// ReadBody 读取响应体
func ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ReadBodyAndClone 读取响应体并返回可再次读取的 ReadCloser
func ReadBodyAndClone(resp *http.Response) ([]byte, io.ReadCloser, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	return body, io.NopCloser(bytes.NewReader(body)), nil
}
