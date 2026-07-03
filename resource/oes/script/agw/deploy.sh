#!/usr/bin/env sh

dirpath=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)

cd $dirpath/../python

agw_id=$1

shift 1

./agw_playbook.py --playbook_path agw/deploy.yaml --agw_id $agw_id "$@"
