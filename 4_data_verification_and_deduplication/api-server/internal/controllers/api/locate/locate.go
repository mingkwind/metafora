package locate

import (
	"api-server/internal/pkg/rabbitmq"
	"os"
	"strconv"
	"time"
)

func Locate(hash string) (string, error) {
	q, err := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	if err != nil {
		return "", err
	}
	q.Publish("dataServers", hash)
	c := q.Consume()
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()
	msg := <-c
	s, _ := strconv.Unquote(string(msg.Body))
	return s, nil
}
