# 星云盘 V2 Makefile

.PHONY: help dev build test clean docker-up docker-down

# 默认目标
help:
	@echo "星云盘 V2 - 可用命令:"
	@echo "  make dev         - 启动开发服务器"
	@echo "  make build       - 编译项目"
	@echo "  make test        - 运行测试"
	@echo "  make clean       - 清理构建文件"
	@echo "  make docker-up   - 启动 Docker 依赖服务"
	@echo "  make docker-down - 停止 Docker 依赖服务"
	@echo "  make deps        - 下载依赖"
	@echo "  make fmt         - 格式化代码"
	@echo "  make lint        - 代码检查"

# 启动开发服务器
dev:
	go run cmd/server/main.go

# 编译项目
build:
	go build -o bin/xingyunpan-server cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/
	rm -rf tmp/
	go clean

# 启动 Docker 依赖服务
docker-up:
	docker-compose up -d

# 停止 Docker 依赖服务
docker-down:
	docker-compose down

# 下载依赖
deps:
	go mod download
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 查看 Docker 服务日志
logs:
	docker-compose logs -f

# 重启 Docker 服务
restart:
	docker-compose restart
