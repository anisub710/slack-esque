export SUMMARYADDR=summary:4000

docker rm -f summary

docker pull ask710/summary

docker run -d \
--network authnet \
--name summary \
-e ADDR=$SUMMARYADDR \
ask710/summary