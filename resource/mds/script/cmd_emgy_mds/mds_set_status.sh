#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

colony_num=$1

task_name=$2

task_status=$3

shift 3

./mds_playbook.py --playbook_path emergency/set_status_main.yaml --colony_num $colony_num --extravars "task_name=$task_name,task_status=$task_status" "$@"
