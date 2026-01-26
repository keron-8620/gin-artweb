#!/usr/bin/env sh

# 检查环境变量是否存在
if [ -z "$JOBS_BASE_DIR" ]; then
    echo "错误:环境变量 JOBS_BASE_DIR 未设置或值为空!"
    exit 1  # 退出脚本并返回错误码
fi

# 检查环境变量是否存在
if [ -z "$JOBS_RECORD_ID" ]; then
    echo "错误:环境变量 JOBS_RECORD_ID 未设置或值为空!"
    exit 1  # 退出脚本并返回错误码
fi

# 修复1:处理$0在sh中可能的兼容性问题，且给变量加引号避免空格问题
basepath=$(cd "$(dirname "$0")" || exit; pwd)

# 修复2:切换目录前检查目录是否存在，避免cd失败导致后续操作异常
cd "$basepath/../python" || {
    echo "错误:目录 $basepath/../python 不存在!"
    exit 1
}

# 修复3:检查第一个参数colony_num是否传入，避免空值导致文件路径异常
if [ -z "$1" ]; then
    echo "错误:未传入colony_num参数!"
    exit 1
fi
colony_num=$1

shift 1

# 定义要写入的文件路径（加引号避免路径含空格）
OUTPUT_FILE="$JOBS_BASE_DIR/storage/oes/flags/$colony_num/.counter_fetch"

# 修复4:先创建文件所在的目录（如果不存在），否则写入会失败
mkdir -p "$(dirname "$OUTPUT_FILE")" || {
    echo "错误:创建目录 $(dirname "$OUTPUT_FILE") 失败!"
    exit 1
}

# 将环境变量的值写入文件
echo "$JOBS_RECORD_ID" > "$OUTPUT_FILE"

# 验证写入是否成功
if [ $? -eq 0 ]; then
    echo "成功!环境变量 JOBS_RECORD_ID 的值已写入文件:$OUTPUT_FILE"
    echo "值为:$JOBS_RECORD_ID"
else
    echo "错误:写入文件 $OUTPUT_FILE 失败!"
    exit 1
fi

./playbook.py --playbook_path collector/counter_fetch/fetch_main.yaml --colony_num $colony_num "$@"
