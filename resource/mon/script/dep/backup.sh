#!/bin/bash

cd ../python

mon_id=$1

shift 1

./playbook.py --playbook_path dep/backup.yml --mon_id $mon_id "$@"
