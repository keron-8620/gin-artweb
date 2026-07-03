#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

colony_num=$1

process_name=$2

shift 2

./oes_playbook.py --playbook_path emergency/cancel_onload/cancel_onload_main.yaml --colony_num $colony_num "$@"
