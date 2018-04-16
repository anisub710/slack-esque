docker build -t ask710/usersdb .

docker run -d \
-p 127.0.0.1:3306:3306 \
--name usersdb \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
ask710/usersdb