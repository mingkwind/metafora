package locate

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"data-server/lib/rabbitmq"
)

func locate(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func LocateApi(w http.ResponseWriter, r *http.Request) {
	// GET /locate/<object_name>
	// 只允许GET请求
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 获取object_name
	objectName := strings.Split(r.URL.EscapedPath(), "/")[2]
	if objectName == "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if locate(objectName) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
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
			object, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				panic(e)
			}
			if locate(os.Getenv("STORAGE_ROOT") + "/objects/" + object) {
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
