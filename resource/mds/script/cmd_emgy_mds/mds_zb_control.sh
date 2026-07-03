#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

colony_num=$1

zb_active=$2

shift 2

./mds_playbook.py --playbook_path emergency/zb_control.yaml --colony_num $colony_num --extravars "zb_active=$zb_active" "$@"
