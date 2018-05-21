#! /usr/bin/env bash
docker rm -f client
docker pull ask710/auth-client
docker run -d \
--name client \
-p 80:80 -p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
ask710/auth-client