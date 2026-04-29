#!/usr/bin/env sh

basepath=$(cd `dirname $0`; pwd)

cd $basepath/../python

colony_num=$1

is_clear=$2

if [ "$is_clear" = "ct" ] || [ "$is_clear" = "nct" ]; then
    is_clear=$2
else
    echo "is_clear must be ct or nct"
    exit 1
fi

echo "is_clear: $is_clear"

shift 1

./playbook.py --playbook_path emergency/restart_main.yaml --colony_num $colony_num --extravars "is_clear=$is_clear" "$@"
