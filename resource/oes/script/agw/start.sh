#!/usr/bin/env sh

dirpath=$(cd `dirname $0`; pwd)

cd $dirpath/../python

agw_id=$1

shift 1

./agw_playbook.py --playbook_path agw/start_main.yaml --agw_id $agw_id "$@"
