#!/usr/bin/env sh

basepath=$(cd "$(dirname "$0")/.."; pwd)
cd "$basepath"

# 定义变量
process_name="gin-artweb"
pid_file="$basepath/${process_name}.pid"

# 检查程序是否存在
if [ ! -f "$basepath/$process_name" ]; then
    echo "Error: $process_name executable not found at $basepath/$process_name"
    exit 1
fi

# 检查是否已经在运行
if [ -f "$pid_file" ]; then
    pid=$(cat "$pid_file")
    if kill -0 "$pid" 2>/dev/null; then
        echo "$process_name is already running with PID: $pid"
        exit 1
    else
        # PID文件存在但进程已退出，移除旧PID文件
        rm -f "$pid_file"
    fi
fi

echo "Starting $process_name..."

# 后台运行程序，丢弃输出或重定向到/dev/null
nohup "$basepath/$process_name" >/dev/null 2>&1 &

# 保存进程ID
echo $! > "$pid_file"

# 验证启动是否成功
sleep 1
if kill -0 $(cat "$pid_file") 2>/dev/null; then
    echo "$process_name started successfully with PID: $(cat "$pid_file")"
else
    echo "Failed to start $process_name"
    rm -f "$pid_file"
    exit 1
fi