#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

mon_id=$1
operate=$2

shift 2

./mon_playbook.py --playbook_path dep/deploy.yaml --mon_id $mon_id --extravars "operate=$operate" "$@"
