#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

task_name=$2

task_status=$3

shift 3

./playbook.py --playbook_path emergency/set_status_main.yaml --colony_num $colony_num --extravars "task_name=$task_name;task_status=$task_status" "$@"
