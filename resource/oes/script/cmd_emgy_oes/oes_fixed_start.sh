#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

colony_num=$1

runner_nodes=$2

shift 2

./oes_playbook.py --playbook_path emergency/fixed_start/fixed_start_main.yaml --colony_num $colony_num --extravars "runner_nodes=$runner_nodes" "$@"
