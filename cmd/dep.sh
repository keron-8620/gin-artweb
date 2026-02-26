#!/usr/bin/env sh

# 脚本功能：将本机指定目录/文件同步到远程服务器，支持自动创建目录、删除远程多余文件
set -e  # 遇到错误立即退出，避免执行后续命令

# 获取脚本所在目录的上级目录（项目根目录）
basepath=$(cd "$(dirname "$0")/.."; pwd)
cd "$basepath" || exit  # 切换到项目根目录，失败则退出

# 定义配置变量
ssh_host="192.168.11.189"
ssh_port=22
ssh_user="ansible"
remote_base_path="/home/ansible/gin-artweb"

# 要同步的文件/目录列表（相对项目根目录）
sync_items=(
    "config"
    "bin"
    ".env"
    "cmd"
    "sql"
    "resource"
    "html"
)

# 遍历要同步的项，逐个执行rsync（确保每个项独立同步，目录不存在自动创建）
for item in "${sync_items[@]}"; do
    # 执行rsync同步
    # 参数说明：
    # -avz：归档模式（保留属性）+ 详细输出 + 压缩传输
    # -e：指定ssh端口
    # --delete：删除远程比本地多的文件/目录
    # --mkdir：自动创建远程不存在的目录
    rsync -avz \
        -e "ssh -p $ssh_port" \
        --delete \
        --mkdir \
        "$item" \
        "${ssh_user}@${ssh_host}:${remote_base_path}/"

    echo "✅  成功同步 $item 到 ${ssh_host}:${remote_base_path}/$item"
done

echo "🎉  所有指定项同步完成！"