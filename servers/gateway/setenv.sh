#!/bin/echo
export MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32)
export MYSQL_DATABASE=users
export MYSQL_ADDR=:3306

export REDISADDR=:6379
export SESSIONKEY="test key"

export DSN="root:$MYSQL_ROOT_PASSWORD@tcp($MYSQL_ADDR)/$MYSQL_DATABASE"

#dev
# export TLSKEY=./tls/privkey.pem
# export TLSCERT=./tls/fullchain.pem
#export ADDR=localhost:4000