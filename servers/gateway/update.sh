#! /usr/bin/env bash
docker rm -f gateway
docker pull ask710/gateway
docker run -d \
--name gateway \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSKEY=/etc/letsencrypt/live/api.ask710.me/privkey.pem \
-e TLSCERT=/etc/letsencrypt/live/api.ask710.me/fullchain.pem \
ask710/gateway