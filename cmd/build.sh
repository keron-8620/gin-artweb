#!/bin/bash

basepath=$(cd `dirname $0`/..; pwd)

cd $basepath

# 编译前清理旧的可执行文件
if [ -f "$basepath/bin/gin-artweb" ]; then
  rm -rf "$basepath/bin/gin-artweb"
fi

VERSION="v0.17.7.0.1"
COMMIT_ID=$(git rev-parse --short HEAD)  # 获取Git短Commit ID
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")

# 增强版（注入版本/构建信息，便于运维）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags "\
    -s -w \
    -X 'main.version=${VERSION}' \
    -X 'main.commitID=${COMMIT_ID}' \
    -X 'main.buildTime=${BUILD_TIME}' \
    -X 'main.goVersion=$(go version)' \
    -X 'main.goOS=$(go env GOOS)' \
    -X 'main.goArch=$(go env GOARCH)'
  " \
  -o bin/gin-artweb main.go