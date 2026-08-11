#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

mon_id=$1

shift 1

./mon_playbook.py --playbook_path dep/deploy.yaml --mon_id $mon_id "$@"
