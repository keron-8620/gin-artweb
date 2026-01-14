#!/usr/bin/env sh

basepath=$(cd "$(dirname "$0")/.."; pwd)
cd "$basepath"

scp bin/gin-artweb ansible@192.168.122.130:/home/ansible/artweb/bin/gin-artweb
