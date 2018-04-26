#! /usr/bin/env bash
export MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32)
export MYSQL_DATABASE=users
export MYSQL_ADDR=usersdb:3306

export REDISADDR=sessionServer:6379
export SESSIONKEY=$(openssl rand -hex 32)

export DSN="root:$MYSQL_ROOT_PASSWORD@tcp($MYSQL_ADDR)/$MYSQL_DATABASE?parseTime=true"

docker rm -f usersdb
docker rm -f sessionServer
docker network rm authnet
docker network create authnet


docker pull ask710/usersdb

docker run -d \
--network authnet \
--name usersdb \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
ask710/usersdb


docker run -d \
--network authnet \
--name sessionServer \
redis

docker rm -f gateway

docker pull ask710/gateway

docker run -d \
--network authnet \
--name gateway \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSKEY=/etc/letsencrypt/live/api.ask710.me/privkey.pem \
-e TLSCERT=/etc/letsencrypt/live/api.ask710.me/fullchain.pem \
-e DSN=$DSN \
-e SESSIONKEY=$SESSIONKEY \
-e REDISADDR=$REDISADDR \
ask710/gateway




