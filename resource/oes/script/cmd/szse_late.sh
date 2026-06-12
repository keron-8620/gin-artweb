#!/usr/bin/env sh

# 处理$0在sh中可能的兼容性问题，且给变量加引号避免空格问题
basepath=$(cd "$(dirname "$0")" || exit; pwd)

# 切换目录前检查目录是否存在，避免cd失败导致后续操作异常
cd "$basepath/../python" || {
    echo "错误:目录 $basepath/../python 不存在!"
    exit 1
}

# 检查第一个参数colony_num是否传入，避免空值导致文件路径异常
if [ -z "$1" ]; then
    echo "错误:未传入colony_num参数!"
    exit 1
fi
colony_num=$1

shift 1

# 定义标志目录路径
flag_rel_path="$basepath/../../../../storage/oes/flags/$colony_num"

# 创建目录（如果不存在）
mkdir -p "$flag_rel_path" || {
    echo "错误:创建目录 $flag_rel_path 失败!"
    exit 1
}

# 获取绝对路径
flag_path=$(cd "$flag_rel_path" && pwd)

# 定义要写入的文件路径
OUTPUT_FILE="$flag_path/.szse_late"

# 检查环境变量是否存在
if [ ! -z "$JOB_RECORD_ID" ]; then
    # 将环境变量的值写入文件
    echo "$JOB_RECORD_ID" > "$OUTPUT_FILE"

    # 验证写入是否成功
    if [ $? -eq 0 ]; then
        echo "成功!环境变量 JOB_RECORD_ID 的值已写入文件:$OUTPUT_FILE"
        echo "值为:$JOB_RECORD_ID"
    else
        echo "错误:写入文件 $OUTPUT_FILE 失败!"
        exit 1
    fi
fi

./playbook.py --playbook_path collector/szse_late_gateway/szse_late_main.yaml --colony_num $colony_num "$@"
