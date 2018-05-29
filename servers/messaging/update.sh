export MESSAGESADDR=messages:80
export MYSQL_ROOT_PASSWORD=lkdsnalkfnadkjbdflajbslajbd
export MYSQL_DATABASE=users
export MYSQL_ADDR=usersdb
export MQADDR=messagequeue
export MQNAME=messagequeue

docker rm -f messages

docker pull ask710/messages

docker run -d \
--network authnet \
--name messages \
-e ADDR=$MESSAGESADDR \
-e MYSQL_ADDR=$MYSQL_ADDR \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MQADDR=$MQADDR \
-e MQNAME=$MQNAME \
ask710/messages