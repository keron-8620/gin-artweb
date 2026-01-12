.PHONY: build test clean docker-build docker-push run help version

# 项目配置
BINARY_NAME=gin-artweb
VERSION?=0.17.7.0.1
COMMIT_ID?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date +"%Y-%m-%d %H:%M:%S")
DOCKER_IMAGE?=swr.cn-north-4.myhuaweicloud.com/danqingzhao/gin-artweb
DOCKER_TAG?=$(VERSION)

# 环境变量
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

help:  ## 显示帮助信息
	@echo "可用的命令:"
	@echo ""
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build:  ## 构建二进制文件
	@echo "构建项目..."
	@go build \
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

test:  ## 运行测试
	@echo "运行测试..."
	@go test -v ./...
	@echo "测试完成"

clean:  ## 清理构建产物
	@echo "清理构建产物..."
	@if [ -f "bin/$(BINARY_NAME)" ]; then \
		rm -f bin/$(BINARY_NAME); \
		echo "已删除 bin/$(BINARY_NAME)"; \
	fi
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

version:  ## 显示版本信息
	@go run main.go -v