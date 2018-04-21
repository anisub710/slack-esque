#!/bin/echo please source using `source setenv.sh`
export MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32)
export MYSQL_DATABASE=users
export MYSQL_ADDR=127.0.0.1:3306

export REDISADDR=localhost:6379
export SESSIONKEY="test key"

export TLSKEY=./tls/privkey.pem
export TLSCERT=./tls/fullchain.pem

export DSN=root:$MYSQL_ROOT_PASSWORD@tcp($MYSQL_ADDR)/$MYSQL_DATABASE
export ADDR=localhost:4000