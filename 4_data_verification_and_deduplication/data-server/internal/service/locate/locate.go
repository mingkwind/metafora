package locate

import (
	"log"
	"os"
	"path"
	"strconv"
	"sync"

	"data-server/internal/pkg/rabbitmq"
	"path/filepath"
)

var (
	locateMap = sync.Map{}
)

func init() {
	// 遍历所有的文件，将文件的hash值和文件的路径存储到locateMap中
	filepath.Walk(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			locateMap.Store(info.Name(), struct{}{})
		}
		return nil
	})
	/*
		locateMap.Range(func(key, value interface{}) bool {
			log.Println("locate service: loaded file: ", key)
			return true
		})
	*/
	// 打印加载的文件数量
	Len := 0
	locateMap.Range(func(key, value interface{}) bool {
		Len++
		return true
	})
	log.Println("locate service: loaded ", Len, " files")
}

func Add(hash string) {
	locateMap.Store(hash, struct{}{})
}

func Del(hash string) {
	locateMap.Delete(hash)
}

func locate(hash string) bool {
	_, ok := locateMap.Load(hash)
	return ok
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
			if locate(hash) {
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
