#! /usr/bin/env bash
./build.sh
docker push ask710/messages
ssh root@api.ask710.me 'bash -s' < update.sh 