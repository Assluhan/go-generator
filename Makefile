.PHONY: help build clean test release install

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build     - 构建项目"
	@echo "  clean     - 清理构建文件"
	@echo "  test      - 运行测试"
	@echo "  release   - 发布新版本 (需要版本号，如: make release VERSION=v1.0.1)"
	@echo "  install   - 安装到本地"
	@echo "  help      - 显示此帮助信息"

# 构建项目
build:
	@echo "构建项目..."
	go build -o generator main.go
	@echo "✅ 构建完成: ./generator"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f generator
	@echo "✅ 清理完成"

# 运行测试
test:
	@echo "运行测试..."
	go test ./...
	@echo "✅ 测试完成"

# 发布新版本
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "错误: 请指定版本号"; \
		echo "使用方法: make release VERSION=v1.0.1"; \
		exit 1; \
	fi
	@echo "发布版本: $(VERSION)"
	@./scripts/release.sh $(VERSION)

# 安装到本地
install:
	@echo "安装到本地..."
	go install .
	@echo "✅ 安装完成"

# 检查代码质量
lint:
	@echo "检查代码质量..."
	golangci-lint run
	@echo "✅ 代码检查完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "✅ 代码格式化完成"

# 更新依赖
deps:
	@echo "更新依赖..."
	go mod tidy
	go mod download
	@echo "✅ 依赖更新完成"
