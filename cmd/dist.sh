#!/usr/bin/env sh

# 设置脚本选项
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时退出

# 显示帮助信息
show_help() {
    echo "用法：$0"
    echo "功能：打包项目为tar.gz格式，文件名格式为artweb-版本号.tar.gz"
    echo "版本号通过执行./bin/artweb -v命令获取"
}

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# 项目根目录
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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
    
    # 提取版本号
    local version
    version="$(echo "$version_output" | grep "版本号" | awk -F"[:：]" '{print $2}' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    
    if [ -z "$version" ]; then
        echo "错误：无法从命令输出中提取版本号"
        echo "命令输出：$version_output"
        exit 1
    fi
    
    echo "$version"
}

# 验证解压结构
verify_extract_structure() {
    local package_name="$1"
    local version="$2"
    local expected_dir="artweb-${version}"
    
    echo "验证解压结构..."
    
    # 创建临时解压目录
    local temp_extract_dir="temp_extract_$(date +%s)"
    mkdir -p "$temp_extract_dir"
    
    # 解压到临时目录
    tar -xzf "$package_name" -C "$temp_extract_dir"
    
    if [ $? -ne 0 ]; then
        echo "错误：解压验证失败"
        rm -rf "$temp_extract_dir"
        exit 1
    fi
    
    # 检查解压后的目录结构
    local item_count
    item_count="$(ls -1 "$temp_extract_dir" | grep -v "^\." | wc -l)"
    
    if [ "$item_count" -ne 1 ]; then
        echo "错误：解压后应只包含一个顶级目录"
        echo "解压内容："
        ls -la "$temp_extract_dir"
        rm -rf "$temp_extract_dir"
        exit 1
    fi
    
    local extracted_dir
    extracted_dir="$(ls -1 "$temp_extract_dir" | grep -v "^\.")"
    
    if [ "$extracted_dir" != "$expected_dir" ]; then
        echo "错误：顶级目录名称不正确"
        echo "期望：$expected_dir"
        echo "实际：$extracted_dir"
        rm -rf "$temp_extract_dir"
        exit 1
    fi
    
    echo "解压结构验证成功！"
    
    # 清理临时目录
    rm -rf "$temp_extract_dir"
}

# 主函数
main() {
    echo "开始打包项目..."
    
    # 切换到项目根目录
    cd "$PROJECT_ROOT"
    echo "当前工作目录：$(pwd)"
    
    # 提取版本号
    echo "正在获取版本号..."
    local version
    version="$(extract_version)"
    echo "提取到版本号：$version"
    
    # 构建包名
    local package_name="artweb-${version}.tar.gz"
    echo "生成包名：$package_name"
    
    # 定义要打包的文件和目录
    dist_items="config bin .env cmd sql resource html README.md"
    
    # 检查要打包的文件和目录是否存在
    for item in $dist_items; do
        if [ ! -e "$item" ]; then
            echo "警告：$item 不存在，将跳过"
        fi
    done
    
    # 创建临时目录用于打包
    local temp_dir="artweb-${version}"
    mkdir -p "$temp_dir"
    
    # 复制文件到临时目录
    for item in $dist_items; do
        if [ -e "$item" ]; then
            cp -r "$item" "$temp_dir/"
        fi
    done
    
    # 执行打包
    echo "开始创建压缩包..."
    tar -czf "$package_name" "$temp_dir"
    
    if [ $? -ne 0 ]; then
        echo "错误：打包失败"
        # 清理临时目录
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # 清理临时目录
    rm -rf "$temp_dir"
    
    echo "打包成功！生成文件：$package_name"
    echo "文件大小：$(du -h "$package_name" | awk '{print $1}')"
    
    # 验证解压结构
    verify_extract_structure "$package_name" "$version"
    
    echo "所有验证通过！"
}

# 执行主函数
main