#!/usr/bin/env sh

# 设置脚本选项
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时退出

# 获取并切换到项目根目录
basepath=$(cd "$(dirname "$0")/.." && pwd)
cd "$basepath"

# 1. 代码格式化
echo "===== 1. go fmt ====="
go fmt ./...

# 2. 静态检查
echo "===== 2. go vet ====="
go vet ./...

# 3. 代码规范检查
echo "===== 3. golangci-lint ====="
golangci-lint run

# 4. 安全漏洞检查
# echo "===== 4. gosec ====="
# gosec -quiet ./...

# 5. 依赖漏洞检查
# echo "===== 5. govulncheck ====="
# govulncheck ./...

# 6. 单元测试（带竞争检测 + 输出覆盖率）
echo "===== 6. go test ====="
CGO_ENABLED=1 go test -race -coverprofile=coverage.out ./...

# 7. 生成覆盖率报告
echo "===== 7. 覆盖率报告 ====="
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
