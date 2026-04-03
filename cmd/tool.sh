#!/usr/bin/env sh

# 设置脚本选项
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时退出

# 获取并切换到项目根目录
PROJECT_ROOT=$(cd `dirname $0`/..; pwd)
cd "$PROJECT_ROOT"

# 提取版本号
extract_version() {
    # 检查artweb可执行文件是否存在
    if [ ! -x "$PROJECT_ROOT/bin/artweb" ]; then
        echo "错误：$PROJECT_ROOT/bin/artweb 不存在或不可执行"
        exit 1
    fi
    
    # 执行命令获取版本信息
    local version_output
    version_output="$($PROJECT_ROOT/bin/artweb -v 2>&1)"
    
    if [ $? -ne 0 ]; then
        echo "错误：执行 ./bin/artweb -v 命令失败"
        echo "命令输出：$version_output"
        exit 1
    fi
    
    # 提取版本号（兼容 数字+下划线后缀）
    local version
    version="$(echo "$version_output" | grep -E "版本号" | awk -F"[:：]" '{print $2}' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    
    if [ -z "$version" ]; then
        echo "错误：无法从命令输出中提取版本号"
        echo "命令输出：$version_output"
        exit 1
    fi
    
    echo "$version"
}

# 验证参数
validate_params() {
    if [ $# -ne 2 ]; then
        echo "错误：需要提供两个参数：版本号和操作类型（update/rollback）"
        echo "用法：$0 <版本号> <操作类型>"
        exit 1
    fi
    
    local version="$1"
    local operation="$2"
    
    if [ "$operation" != "update" ] && [ "$operation" != "rollback" ]; then
        echo "错误：操作类型必须是 'update' 或 'rollback'"
        echo "用法：$0 <版本号> <操作类型>"
        exit 1
    fi
}

# 版本标准化（兼容任意段 + 下划线后缀，用于比较）
normalize_version() {
    echo "$1" | sed -e 's/_/./g' -e 's/[^0-9.]//g' -e 's/\.\././g' | sed -e 's/^\.//' -e 's/\.$//'
}

# 通用版本比较（支持任意段数字 + 下划线后缀）
compare_versions() {
    local v1="$1"
    local v2="$2"
    
    # 标准化：把 _u1 变成 .u1 → 只保留数字和点
    local v1_norm=$(normalize_version "$v1")
    local v2_norm=$(normalize_version "$v2")
    
    # 逐段比较（支持无限段数字）
    awk -v v1="$v1_norm" -v v2="$v2_norm" '
    BEGIN {
        split(v1, a, ".");
        split(v2, b, ".");
        max_len = (length(a) > length(b)) ? length(a) : length(b);
        
        for (i = 1; i <= max_len; i++) {
            n1 = (a[i] == "") ? 0 : a[i] + 0;
            n2 = (b[i] == "") ? 0 : b[i] + 0;
            if (n1 < n2) { print -1; exit; }
            if (n1 > n2) { print 1; exit; }
        }
        print 0;
    }'
}

# 获取版本目录列表（完全兼容你的目录格式）
get_version_dirs() {
    local start_version="$1"
    local end_version="$2"
    local operation="$3"
    
    # 确保sql目录存在
    if [ ! -d "$PROJECT_ROOT/sql" ]; then
        echo "错误：$PROJECT_ROOT/sql 目录不存在"
        exit 1
    fi
    
    # 获取所有 v开头 版本目录
    local all_dirs=$(ls -d "$PROJECT_ROOT/sql"/v* 2>/dev/null | sort -V)
    
    if [ -z "$all_dirs" ]; then
        echo "错误：sql目录下没有版本目录"
        exit 1
    fi
    
    local filtered_dirs=""
    
    for dir in $all_dirs; do
        # 提取目录版本：去掉 v 前缀
        local dir_version=$(basename "$dir" | sed 's/^v//')
        
        # 版本区间判断
        if [ "$operation" = "update" ]; then
            cmp_start=$(compare_versions "$dir_version" "$start_version")
            cmp_end=$(compare_versions "$dir_version" "$end_version")
            if [ "$cmp_start" -gt 0 ] && [ "$cmp_end" -le 0 ]; then
                filtered_dirs="$filtered_dirs $dir"
            fi
        else
            cmp_start=$(compare_versions "$dir_version" "$end_version")
            cmp_end=$(compare_versions "$dir_version" "$start_version")
            if [ "$cmp_start" -gt 0 ] && [ "$cmp_end" -le 0 ]; then
                filtered_dirs="$filtered_dirs $dir"
            fi
        fi
    done
    
    # 回滚需要倒序
    if [ "$operation" = "rollback" ]; then
        filtered_dirs=$(echo "$filtered_dirs" | tr ' ' '\n' | sort -Vr | tr '\n' ' ')
    fi
    
    echo "$filtered_dirs"
}

# 执行SQL文件
execute_sql() {
    local sql_file="$1"
    
    if [ ! -f "$sql_file" ]; then
        echo "警告：$sql_file 不存在，跳过执行"
        return 1
    fi
    
    echo "执行SQL文件：$sql_file"
    
    local result
    result="$($PROJECT_ROOT/bin/artweb -exec-sql "$sql_file" 2>&1)"
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo "成功：SQL文件执行完成"
        return 0
    else
        echo "失败：SQL文件执行出错"
        echo "错误信息：$result"
        return 1
    fi
}

# 主函数
main() {
    validate_params "$@"
    
    local target_version="$1"
    local operation="$2"
    
    local current_version=$(extract_version)
    echo "当前版本：$current_version"
    echo "目标版本：$target_version"
    echo "操作类型：$operation"
    
    local version_diff=$(compare_versions "$current_version" "$target_version")
    
    if [ "$operation" = "update" ]; then
        if [ "$version_diff" -ge 0 ]; then
            echo "错误：目标版本必须大于当前版本"
            exit 1
        fi
    else
        if [ "$version_diff" -le 0 ]; then
            echo "错误：目标版本必须小于当前版本"
            exit 1
        fi
    fi
    
    local version_dirs=$(get_version_dirs "$current_version" "$target_version" "$operation")
    
    if [ -z "$version_dirs" ]; then
        echo "没有需要执行的版本目录"
        exit 0
    fi
    
    echo "找到以下版本目录："
    echo "$version_dirs"
    
    for dir in $version_dirs; do
        local sql_file
        if [ "$operation" = "update" ]; then
            sql_file="$dir/update.sql"
        else
            sql_file="$dir/rollback.sql"
        fi
        execute_sql "$sql_file"
    done
    
    echo "数据库版本管理操作完成"
}

main "$@"
