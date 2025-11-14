#!/bin/bash

basepath=$(cd `dirname $0`/..; pwd)

cd $basepath

# 编译前清理旧的可执行文件
if [ -f "$basepath/gin-artweb" ]; then
  rm -rf "$basepath/gin-artweb"
fi

# 编译到临时目录
CGO_ENABLED=0 GOOS=linux go build -o gin-artweb .

# 执行
"$basepath/gin-artweb" --config "$basepath/config/system.yml"
