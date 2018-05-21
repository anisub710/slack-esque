#! /usr/bin/env bash
./build.sh
docker push ask710/summary-client
ssh root@ask710.me 'bash -s' < update.sh