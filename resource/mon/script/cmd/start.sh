#!/bin/bash

cd ../python

mon_id=$1

shift 1

./playbook.py --playbook_path control/start_main.yml --mon_id $mon_id "$@"
