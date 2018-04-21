
docker build -t ask710/usersdb .

docker run -d \
-p 3306:3306 \
--name usersdb \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
ask710/usersdb

#For local test
# docker run -it \
# --rm \
# --network host \
# mysql sh -c "mysql -h127.0.0.1 -uroot -p$MYSQL_ROOT_PASSWORD"