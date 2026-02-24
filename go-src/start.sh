#!/bin/bash
# ChatJimmy2API - 快速启动脚本

set -e

echo "🚀 ChatJimmy2API - 快速启动"
echo ""

# 检查 Go 安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误：未找到 Go，请先安装 Go 1.22+"
    exit 1
fi

echo "✅ Go 版本：$(go version)"

# 下载依赖
echo ""
echo "📦 下载依赖..."
GOPROXY=${GOPROXY:-https://goproxy.cn,direct} go mod download

# 编译
echo ""
echo "🔨 编译项目..."
GOPROXY=${GOPROXY:-https://goproxy.cn,direct} go build -o chatjimmy2api .

# 创建必要目录
mkdir -p config logs data

# 检查配置文件
if [ ! -f "config/config.json" ]; then
    echo ""
    echo "📝 创建默认配置文件..."
    cp config/config.example.json config/config.json
fi

echo ""
echo "✅ 准备就绪！"
echo ""
echo "📌 服务信息:"
echo "   - API 服务：http://127.0.0.1:8787"
echo "   - 管理界面：http://127.0.0.1:8788"
echo "   - 配置文件：config/config.json"
echo ""
echo "🔧 常用命令:"
echo "   - 启动服务：./chatjimmy2api"
echo "   - 运行测试：go test ./..."
echo "   - 查看配置：cat config/config.json"
echo ""
echo "🌐 快速测试:"
echo "   curl http://127.0.0.1:8787/health"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

# 启动服务
./chatjimmy2api
