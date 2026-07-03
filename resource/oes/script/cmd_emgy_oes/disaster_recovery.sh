#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

colony_num=$1

shift 1

./oes_playbook.py --playbook_path emergency/disaster_recovery_main.yaml --colony_num $colony_num "$@"
