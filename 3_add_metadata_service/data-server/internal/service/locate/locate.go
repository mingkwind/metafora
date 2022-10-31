package locate

import (
	"log"
	"os"
	"path"
	"strconv"

	"data-server/internal/pkg/rabbitmq"
)

func locate(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func StartLocateService() {
	q := rabbitmq.RetryNew(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("dataServers")
	c := q.Consume()
	log.Println("locate service started")
	notifyCloseChan, notifyCancelChan := q.GetNotifyChan()
	defer close(notifyCloseChan)
	defer close(notifyCancelChan)
	for {
		select {
		case msg := <-c:
			log.Println("locate service: received locate request: ", string(msg.Body))
			hash, err := strconv.Unquote(string(msg.Body))
			if err != nil {
				log.Println("locate service: strconv.Unquote error: ", err)
				continue
			}
			if locate(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", hash)) {
				// q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
				q.Send(msg.ReplyTo, os.Getenv("DATA_SERVER_ADDRESS"))
			}
		case err := <-notifyCloseChan:
			if err != nil {
				log.Println("locate service: retrying to connect to rabbitmq")
				q.Close()
				StartLocateService()
				return
			} else {
				log.Println("locate service: rabbitmq channel closed gracefully")
			}
		case <-notifyCancelChan:
			log.Println("locate service: retrying to connect to rabbitmq")
			q.Close()
			StartLocateService()
			return
		}
	}
}
