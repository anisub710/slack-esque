
docker build -t ask710/usersdb .

docker push ask710/usersdb

docker rm -f usersdb

docker run -d \
-p 3306:3306 \
--name usersdb \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
ask710/usersdb

#redis 
# docker run --name devredis -d -p 6379:6379 redis

#For local test
# docker run -it \
# --rm \
# --network host \
# mysql sh -c "mysql -h127.0.0.1 -uroot -p$MYSQL_ROOT_PASSWORD"