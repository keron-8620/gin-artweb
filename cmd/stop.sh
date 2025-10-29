#!/bin/bash

basepath=$(cd `dirname $0`; pwd)

cd $basepath

# 编译到临时目录
go build -o "$basepath/bin/main" .

# 执行
"$basepath/bin/main"
