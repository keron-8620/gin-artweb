#!/bin/bash

basepath=$(cd `dirname $0`/..; pwd)

cd $basepath

if [ -d "$basepath/bin" ]; then
  rm -rf "$basepath/bin"
fi

# 创建临时目录
mkdir bin

# 编译到临时目录
go build -o "$basepath/bin/artweb" main.go

# 执行
"$basepath/bin/artweb" --config "$basepath/config/system.yml"
