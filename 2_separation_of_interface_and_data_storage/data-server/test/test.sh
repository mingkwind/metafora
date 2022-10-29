export RABBITMQ_SERVER=amqp://root:123456@127.0.0.1:5672
export LISTEN_ADDRESS=127.0.0.1:8000
export DATA_SERVER_ADDRESS=dataserver1:8000
export STORAGE_ROOT=../files/

go run ../dataServer.go