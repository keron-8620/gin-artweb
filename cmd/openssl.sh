#!/bin/sh

# 设置脚本选项
set -e  # 遇到错误时退出
set -u  # 使用未定义变量时报错

# 配置变量
OPENSSL_CONFIG_FILE="${OPENSSL_CONFIG_FILE:-config/openssl.cnf}"

# 如果是绝对路径，则直接使用，否则按照相对路径处理
if [ "${OPENSSL_CONFIG_FILE#/}" != "$OPENSSL_CONFIG_FILE" ]; then
    # 这是一个绝对路径 (/ 开头)
    OUTPUT_DIR=$(dirname "$OPENSSL_CONFIG_FILE")
    CONFIG_FILE_NAME=$(basename "$OPENSSL_CONFIG_FILE")
else
    # 这是一个相对路径
    OUTPUT_DIR="$(cd "$(dirname "$0")/.." && pwd)/$(dirname "$OPENSSL_CONFIG_FILE")"
    CONFIG_FILE_NAME="$(basename "$OPENSSL_CONFIG_FILE")"
fi

CERT_KEY_NAME="${CERT_KEY_NAME:-server.key}"
CERT_CSR_NAME="${CERT_CSR_NAME:-server.csr}"
CERT_CRT_NAME="${CERT_CRT_NAME:-server.crt}"
KEY_SIZE="${KEY_SIZE:-2048}"
VALID_DAYS="${VALID_DAYS:-1095}"  # 默认有效期3年

# 检查配置文件是否存在
if [ ! -f "$OUTPUT_DIR/$CONFIG_FILE_NAME" ]; then
    echo "错误: 找不到配置文件 $OUTPUT_DIR/$CONFIG_FILE_NAME"
    exit 1
fi

# 进入输出目录
cd "$OUTPUT_DIR" || {
    echo "错误: 无法进入目录 $OUTPUT_DIR"
    exit 1
}

echo "使用配置:"
echo "  配置文件: $CONFIG_FILE_NAME"
echo "  输出目录: $OUTPUT_DIR"
echo "  私钥文件: $CERT_KEY_NAME"
echo "  CSR文件:  $CERT_CSR_NAME"
echo "  证书文件: $CERT_CRT_NAME"
echo "  密钥长度: $KEY_SIZE"
echo "  有效天数: $VALID_DAYS"
echo ""

echo "正在生成SSL证书..."

# 生成私钥
echo "1. 生成私钥 ($CERT_KEY_NAME)..."
if openssl genrsa -out "$CERT_KEY_NAME" "$KEY_SIZE"; then
    chmod 600 "$CERT_KEY_NAME"  # 设置适当的权限
    echo "   私钥生成成功"
else
    echo "错误: 生成私钥失败"
    exit 1
fi

# 生成证书签名请求
echo "2. 生成证书签名请求 ($CERT_CSR_NAME)..."
if openssl req -new -key "$CERT_KEY_NAME" -out "$CERT_CSR_NAME" -config "$CONFIG_FILE_NAME"; then
    echo "   证书签名请求生成成功"
else
    echo "错误: 生成证书签名请求失败"
    exit 1
fi

# 生成自签名证书
echo "3. 生成自签名证书 ($CERT_CRT_NAME)..."
if openssl x509 -req -days "$VALID_DAYS" -in "$CERT_CSR_NAME" -signkey "$CERT_KEY_NAME" -out "$CERT_CRT_NAME" -extfile "$CONFIG_FILE_NAME" -extensions v3_req; then
    echo "   自签名证书生成成功"
else
    echo "错误: 生成自签名证书失败"
    exit 1
fi

echo ""
echo "SSL证书生成完成!"
echo "生成的文件位置:"
echo "  私钥: $OUTPUT_DIR/$CERT_KEY_NAME"
echo "  CSR:  $OUTPUT_DIR/$CERT_CSR_NAME"
echo "  CRT:  $OUTPUT_DIR/$CERT_CRT_NAME"
echo ""
echo "注意: $CERT_KEY_NAME 是私钥，请妥善保管，不要泄露"