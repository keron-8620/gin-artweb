#!/usr/bin/env sh
set -euo pipefail
IFS=$'\n\t'

# 切换到项目根目录
basepath=$(cd "$(dirname "$0")/.." && pwd)
cd "$basepath"

# ====================== 关键修改：创建 html 目录 ======================
HTML_DIR="html"
mkdir -p "$HTML_DIR" # 自动创建目录，不存在则创建，存在不报错
# =====================================================================

# 清理旧文件
rm -f coverage.out
rm -f "$HTML_DIR/coverage.html"

echo "===== 1. go fmt 代码格式化 ====="
go fmt ./...

echo "===== 2. go vet 静态检查 ====="
go vet ./...

echo "===== 3. golangci-lint 代码规范 ====="
golangci-lint run --timeout 5m

echo "===== 4. gosec 安全漏洞扫描 ====="
gosec -quiet --exclude-dir=.venv ./...

# echo "===== 5. govulncheck 依赖漏洞 ====="
# govulncheck ./...

echo "===== 6. go test 单元测试（含竞争检测） ====="
CGO_ENABLED=1 go test \
  -race \
  -timeout 120s \
  -parallel $(nproc) \
  -coverprofile=coverage.out \
  -failfast \
  ./...

echo "===== 7. 生成覆盖率报告 → html/coverage.html ====="
go tool cover -func=coverage.out
# ====================== 关键修改：输出到 html 文件夹 ======================
go tool cover -html=coverage.out -o "$HTML_DIR/coverage.html"
# ========================================================================

echo "=================================================="
echo "✅ 所有检查与测试通过！"
echo "📄 覆盖率报告：$HTML_DIR/coverage.html"
echo "=================================================="
