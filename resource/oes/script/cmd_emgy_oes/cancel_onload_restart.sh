#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

shift 1

./playbook.py --playbook_path emergency/cancel_onload_main.yaml --colony_num $colony_num "$@"
