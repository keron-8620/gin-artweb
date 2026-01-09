#!/usr/bin/env sh

cd ../python

mon_id=$1

shift 1

./playbook.py --playbook_path control/stop_main.yml --mon_id $mon_id "$@"
