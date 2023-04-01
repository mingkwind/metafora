package heartbeat

import (
	"log"
	"os"
	"time"

	"data-server/internal/pkg/rabbitmq"
)

func StartHeartbeatService() {
	q := rabbitmq.RetryNew(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	log.Println("heartbeat service started")
	// 发送心跳消息
	notifyCloseChan, notifyCancelChan := q.GetNotifyChan()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			q.Publish("apiServers", os.Getenv("DATA_SERVER_ADDRESS"))
		case err := <-notifyCloseChan:
			if err != nil {
				log.Println("heartbeat service: retrying to connect to rabbitmq")
				q.Close()
				StartHeartbeatService()
				return
			} else {
				log.Println("heartbeat service: rabbitmq channel closed gracefully")
			}
		case <-notifyCancelChan:
			log.Println("heartbeat service: retrying to connect to rabbitmq")
			q.Close()
			StartHeartbeatService()
			return
		}
	}
}
