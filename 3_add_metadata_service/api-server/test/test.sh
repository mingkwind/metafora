export RABBITMQ_SERVER=amqp://root:123456@127.0.0.1:5672
export LISTEN_ADDRESS=127.0.0.1:5000
export ES_URL=http://localhost:9200
export ES_USER=elastic
export ES_PASS=123456
go run ../cmd/app/main.go