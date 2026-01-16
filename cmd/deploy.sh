#!/bin/bash

basepath=$(cd "$(dirname "$0")/.."; pwd)
cd "$basepath"

ssh ansible@192.168.122.130 "mkdir -p /home/ansible/artweb/{bin,config,resource,storage,sql}"
scp -r bin config resource storage sql ansible@192.168.122.130:/home/ansible/artweb/

podman run -d \
  --name mariadb \
  --memory=1G \
  --cpus=0.5 \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=Quant360 \
  -e MYSQL_DATABASE=mysql \
  -e TZ=Asia/Shanghai \
  docker.io/library/mariadb:latest \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci \
  --max_connections=100 \
  --slow_query_log=0

podman run --name opengauss --privileged=true -d -e GS_PASSWORD=Quant@360 -p 5432:5432 opengauss/opengauss-server:latest
