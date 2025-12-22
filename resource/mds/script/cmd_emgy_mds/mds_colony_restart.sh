#!/bin/bash

#获取当前时间-YYYYMMDD
nowdate=$(date +%Y%m%d)
# 获取脚本当前路径
basepath=$(cd `dirname $0`/..; pwd)

colony_num=$1
shift 1

flag_path=$(cd $basepath/../../../storage/mds/flags/$colony_num; pwd)

################### success标志文件列表 ##################
mon_success_flag=$flag_path/mon_collector_$nowdate.success
sse_success_flag=$flag_path/sse_collector_$nowdate.success
szse_success_flag=$flag_path/szse_collector_$nowdate.success

################## 执行任务 ##################
# mon的success标志文件未生成
if [ ! -f $mon_success_flag ];then
	echo "mon成功标志文件未生成, 重新拉取mon上场数据"
	sleep 5
	echo "sh mon.sh $colony_num $@"
	cd $basepath/cmd/;sh mon.sh $colony_num "$@"
fi
# sse的success标志文件未生成
if [ ! -f $sse_success_flag ];then
	echo "拉取上海产品文件成功标志文件未生成, 重新拉取上海产品文件"
	sleep 5
	echo "sh sse.sh $colony_num $@"
	cd $basepath/cmd/;sh sse.sh $colony_num "$@"
fi
# szse的success标志文件未生成
if [ ! -f $szse_success_flag ];then
	echo "拉取深圳产品文件成功标志文件未生成, 重新拉取深圳产品文件"
	sleep 5
	echo "sh szse.sh $colony_num $@"
	cd $basepath/cmd/;sh szse.sh $colony_num "$@"
fi
# 检测所有success标志文件都生成
if [ -f $mon_success_flag ]&&[ -f $sse_success_flag ]&&[ -f $szse_success_flag ];then
	echo "所有标志文件都已经生成, 开始重启mds集群服务器"
	sleep 5
	echo "sh mds_cleartx_start.sh $colony_num;sh mds_set_status.sh $colony_num --extravars 'mds_runner_nodes=master,follow,arbiter;task_name=RESET;task_status=5'"
	cd $basepath/cmd_emgy_mds;sh mds_cleartx_start.sh $colony_num "$@";sh mds_set_status.sh $colony_num --extravars 'mds_runner_nodes=master,follow,arbiter;task_name=RESET;task_status=5' "$@"
else
	echo "标志文件未完全生成, 请检查"
	exit 1
fi
