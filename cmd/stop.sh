#!/usr/bin/env sh

basepath=$(cd "$(dirname "$0")/.."; pwd)
project_path=$basepath
process_name="artweb"
pid_file="$basepath/${process_name}.pid"

# 先尝试从PID文件读取PID
if [ -f "$pid_file" ]; then
    pid_from_file=$(cat "$pid_file")
fi

# 查找正在运行的进程
pid=$(ps aux | grep "$project_path/bin/$process_name" | grep -v grep | awk '{print $2}')

if [ -z "$pid" ]; then
    # 如果通过ps没找到进程，但PID文件存在，尝试使用文件中的PID
    if [ ! -z "$pid_from_file" ]; then
        if kill -0 "$pid_from_file" 2>/dev/null; then
            echo "Found $process_name process with PID from file: $pid_from_file"
            pid=$pid_from_file
        fi
    fi
fi

if [ -z "$pid" ]; then
    echo "No running $process_name process found"
    # 即使没找到运行进程，也要清理可能存在的PID文件
    if [ -f "$pid_file" ]; then
        rm -f "$pid_file"
        echo "Removed stale PID file: $pid_file"
    fi
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
            # 删除PID文件
            if [ -f "$pid_file" ]; then
                rm -f "$pid_file"
                echo "Removed PID file: $pid_file"
            fi
            exit 0
        fi
        sleep 1
        count=$((count + 1))
    done
    
    # 如果进程仍未停止，强制终止
    echo "Process did not stop gracefully, forcing termination..."
    kill -KILL "$pid" 2>/dev/null
    
    if kill -0 "$pid" 2>/dev/null; then
        echo "Failed to stop $process_name process (PID: $pid)"
        exit 1
    else
        echo "$process_name process (PID: $pid) forcefully terminated"
        # 删除PID文件
        if [ -f "$pid_file" ]; then
            rm -f "$pid_file"
            echo "Removed PID file: $pid_file"
        fi
    fi
fi