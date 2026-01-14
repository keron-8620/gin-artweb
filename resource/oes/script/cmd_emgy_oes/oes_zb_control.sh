#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

zb_active=$2

shift 2

./playbook.py --playbook_path emergency/disaster_recovery_main.yaml --colony_num $colony_num --extravars "zb_active=$zb_active" "$@"
