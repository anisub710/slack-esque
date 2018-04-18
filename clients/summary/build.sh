#! /usr/bin/env bash
echo "Building Docker Container Image..."
docker build -t ask710/summary-client .
echo  "Cleaning Up..."
docker image prune -f 