package locate

import (
	"api-server/internal/pkg/rabbitmq"
	"os"
	"strconv"
	"time"
)

func Locate(name string) (string, error) {
	q, err := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	if err != nil {
		return "", err
	}
	q.Publish("dataServers", name)
	c := q.Consume()
	// 1秒后关闭连接
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()
	msg := <-c
	s, _ := strconv.Unquote(string(msg.Body))
	return s, nil
}

func Exist(name string) bool {
	rst, _ := Locate(name)
	return rst != ""
}
