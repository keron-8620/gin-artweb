#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)
project_path=$(cd "$basepath/.."; pwd)
process_name="artweb"

# 查找正在运行的进程
pid=$(ps aux | grep "$project_path/$process_name" | grep -v grep | awk '{print $2}')

if [ -z "$pid" ]; then
    echo "No running $process_name process found"
    exit 1
else
    echo "Found running $process_name process with PID: $pid"
    # 发送 SIGTERM 信号优雅关闭
    kill -TERM "$pid"
    
    # 等待进程结束
    timeout=30
    count=0
    
    while [ $count -lt $timeout ]; do
        if ! kill -0 "$pid" 2>/dev/null; then
            echo "$process_name process (PID: $pid) stopped successfully"
            exit 0
        fi
        sleep 1
        ((count++))
    done
    
    # 如果进程仍未停止，强制终止
    echo "Process did not stop gracefully, forcing termination..."
    kill -KILL "$pid" 2>/dev/null
    
    if kill -0 "$pid" 2>/dev/null; then
        echo "Failed to stop $process_name process (PID: $pid)"
        exit 1
    else
        echo "$process_name process (PID: $pid) forcefully terminated"
    fi
fi