package heartbeat

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"api-server/lib/rabbitmq"
)

var (
	dataServers = &sync.Map{}
)

func ListenHeartbeat() {
	q := rabbitmq.RetryNew(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("apiServers")
	c := q.Consume()
	go removeExpiredDataServer()
	notifyCloseChan, notifyCancelChan := q.GetNotifyChan()
	log.Println("listening heartbeat service started")
	for {
		select {
		case msg := <-c:
			dataServer, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				panic(e)
			}
			dataServers.Store(dataServer, time.Now())
		case err := <-notifyCloseChan:
			if err != nil {
				log.Println("listening heartbeat service: retrying to connect to rabbitmq")
				q.Close()
				ListenHeartbeat()
				return
			} else {
				log.Println("listening heartbeat service: rabbitmq channel closed gracefully")
			}
		case <-notifyCancelChan:
			q.Close()
			log.Println("listening heartbeat service: retrying to connect to rabbitmq")
			ListenHeartbeat()
			return
		}
	}
}

func removeExpiredDataServer() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		dataServers.Range(func(key, value interface{}) bool {
			if value.(time.Time).Add(10 * time.Second).Before(time.Now()) {
				dataServers.Delete(key)
			}
			return true
		})
	}
}

func GetDataServers() []string {
	ds := []string{}
	dataServers.Range(func(key, value interface{}) bool {
		ds = append(ds, key.(string))
		return true
	})
	return ds
}
