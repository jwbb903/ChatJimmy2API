#!/bin/bash
# ChatJimmy Wrapper Go - curl 示例

BASE_URL="http://127.0.0.1:8787"
API_KEY="local-wrapper-key"

echo "=== 获取模型列表 ==="
curl -sS \
  -H "Authorization: Bearer $API_KEY" \
  "$BASE_URL/v1/models" | jq

echo ""
echo "=== 非流式聊天 ==="
curl -sS \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1-8B",
    "messages": [{"role": "user", "content": "用一句话打招呼"}]
  }' \
  "$BASE_URL/v1/chat/completions" | jq

echo ""
echo "=== 流式聊天 ==="
curl -sS -N \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1-8B",
    "stream": true,
    "messages": [{"role": "user", "content": "从 1 数到 3"}]
  }' \
  "$BASE_URL/v1/chat/completions"

echo ""
echo "=== 健康检查 ==="
curl -sS "$BASE_URL/health" | jq
