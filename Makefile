.PHONY: build test verify coverage clean docker-build docker-push run dist help version

# 项目配置：版本默认来自 Git，发布时可通过 VERSION=x.y.z 覆盖。
BINARY_NAME=artweb
# VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
VERSION=0.17.7.1.0
COMMIT_ID?=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME?=$(shell date +"%Y-%m-%d %H:%M:%S")
DOCKER_IMAGE?=swr.cn-north-4.myhuaweicloud.com/danqingzhao/gin-artweb
DOCKER_TAG?=$(VERSION)

# 环境变量
export GOOS=linux
export GOARCH=amd64

help:  ## 显示帮助信息
	@echo "可用的命令:"
	@echo ""
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build:  ## 构建二进制文件
	@echo "构建项目..."
	@CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags "-s -w \
			-X 'main.version=$(VERSION)' \
			-X 'main.commitID=$(COMMIT_ID)' \
			-X 'main.buildTime=$(BUILD_TIME)' \
			-X 'main.goVersion=$(shell go version)' \
			-X 'main.goOS=$(shell go env GOOS)' \
			-X 'main.goArch=$(shell go env GOARCH)'" \
		-o bin/$(BINARY_NAME) main.go
	@echo "构建完成: bin/$(BINARY_NAME)"

swag:  ## 生成swagger文档
	@echo "开始生成生成swagger文档..."
	@swag init
	@echo "生成swagger文档完成"

verify:  ## 执行格式、静态检查、竞争检测和安全扫描

	@echo "===== go vet 静态检查 ====="
	@go vet ./...

	@echo "===== golangci-lint 代码规范 ====="
	@golangci-lint run --timeout 5m

	@echo "===== gosec 安全漏洞扫描 ====="
	@gosec -quiet --exclude-dir=.venv ./...

	@echo "===== go test 单元测试（含竞争检测） ====="
	@CGO_ENABLED=1 go test -race -timeout 300s -parallel $(shell nproc) -coverprofile=coverage.out -failfast ./...


coverage:  ## 生成覆盖率报告
	@mkdir -p html
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o html/coverage.html
	@go tool cover -func=coverage.out | tail -1

clean:  ## 清理构建产物
	@echo "清理构建产物..."
	@if [ -f "bin/$(BINARY_NAME)" ]; then \
		rm -f bin/$(BINARY_NAME); \
		echo "已删除 bin/$(BINARY_NAME)"; \
	fi
	@for f in artweb-*.tar.gz; do \
		if [ -f "$$f" ]; then \
			rm -f "$$f"; \
			echo "已删除 $$f"; \
		fi \
	done
	@echo "清理完成"

docker-build:  ## 构建 Docker 镜像
	@echo "构建 Docker 镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@podman build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker 镜像构建完成"

docker-push:  ## 推送 Docker 镜像
	@echo "推送 Docker 镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@podman push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker 镜像推送完成"

run:  ## 运行应用
	@echo "运行应用..."
	@go run main.go

dist: build  ## 打包发布包 (tar.gz)
	@echo "为 resource/*/script/ 添加可执行权限..."
	@find resource -path '*/script/*' -type f -exec chmod +x {} + 2>/dev/null || true
	@echo "打包为 artweb-$(VERSION).tar.gz ..."
	@mkdir -p /tmp/artweb-dist/artweb-$(VERSION)
	@for item in config bin .env cmd sql resource html README.md; do \
		if [ -e "$$item" ]; then \
			cp -r "$$item" /tmp/artweb-dist/artweb-$(VERSION)/; \
		fi \
	done
	@tar -czf artweb-$(VERSION).tar.gz -C /tmp/artweb-dist artweb-$(VERSION)
	@rm -rf /tmp/artweb-dist
	@echo "打包成功: artweb-$(VERSION).tar.gz ($$(du -h artweb-$(VERSION).tar.gz | awk '{print $$1}'))"

version:  ## 显示版本信息
	@go run main.go -v