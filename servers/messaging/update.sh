export MESSAGESADDR=message:4000

docker rm -f messages

docker pull ask710/messages

docker run -d \
--network authnet \
--name messages \
-e ADDR=$MESSAGESADDR \
ask710/messages