.PHONY: build run test clean download fmt vet

# 变量
BINARY_NAME=chatjimmy2api
GO=go
GOFLAGS=-ldflags "-s -w"

# 默认目标
all: build

# 下载依赖
download:
	$(GO) mod download

# 格式化代码
fmt:
	$(GO) fmt ./...

# 代码检查
vet:
	$(GO) vet ./...

# 运行测试
test:
	$(GO) test -v ./...

# 运行测试（带覆盖率）
test-coverage:
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 编译
build: fmt vet
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# 交叉编译
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux .

build-macos:
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-macos .

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME).exe .

build-all: build-linux build-macos build-windows

# 运行
run:
	$(GO) run main.go

# 清理
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-* coverage.out coverage.html

# 安装
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# 帮助
help:
	@echo "可用命令:"
	@echo "  make build          - 编译当前平台"
	@echo "  make build-all      - 交叉编译所有平台"
	@echo "  make run            - 直接运行"
	@echo "  make test           - 运行测试"
	@echo "  make test-coverage  - 运行测试并生成覆盖率报告"
	@echo "  make fmt            - 格式化代码"
	@echo "  make vet            - 代码检查"
	@echo "  make clean          - 清理编译产物"
	@echo "  make install        - 安装到系统"
