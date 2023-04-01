package locate

import (
	"api-server/internal/pkg/objectstream"
	"api-server/internal/pkg/rabbitmq"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type locateMessage struct {
	Hash string `json:"hash"`
	Id   int    `json:"id"`
	Addr string `json:"addr"`
}

func Locate(hash string) (locateInfo map[int]string) {
	locateInfo = map[int]string{}
	q, err := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	if err != nil {
		return
	}
	q.Publish("dataServers", hash)
	c := q.Consume()
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()
	// 此处因为有6个节点，需要循环接收消息队列中的消息
	for i := 0; i < 6; i++ {
		msg := <-c
		if len(msg.Body) == 0 {
			return
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		var info locateMessage
		json.Unmarshal(msg.Body, &info)
		if err != nil {
			return
		}
		locateInfo[info.Id] = info.Addr
	}
	return
}

func Exist(hash string) bool {
	return len(Locate(hash)) >= objectstream.DATA_SHARDS
}
