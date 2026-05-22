#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

process_name=$2

shift 2

./playbook.py --playbook_path emergency/cancel_onload/cancel_onload_main.yaml --colony_num $colony_num "$@"
