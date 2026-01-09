#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

shift 1

./playbook.py --playbook_path collector/backup_gateway/backup_main.yml --colony_num $colony_num "$@"
