docker rm -f summary

docker pull ask710/summary

docker run -d \
--name summary \
--network authnet \
ask710/summary