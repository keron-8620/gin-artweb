#!/usr/bin/env sh

# 设置脚本选项
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时退出

# 获取并切换到项目根目录
basepath=$(cd `dirname $0`/..; pwd)
cd $basepath

# 更新swag文档
swag init

# 自动化测试并检查结果
go test -v ./...

# 编译前清理旧的可执行文件
if [ -f "$basepath/bin/artweb" ]; then
  rm -rf "$basepath/bin/artweb"
fi

# 注入版本、Commit ID、构建时间等
VERSION="0.17.7.0.2" # 项目版本号
COMMIT_ID=$(git rev-parse --short HEAD) # 获取Git短Commit ID
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S") # 获取当前时间

# 编译程序
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
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
  -o bin/artweb main.go

echo "Build success!"
