package locate

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"data-server/internal/pkg/rabbitmq"
	"path/filepath"
)

// 将locate服改成访问内存，减少磁盘io次数
var (
	locateMap = sync.Map{}
)

func Add(hash string) {
	locateMap.Store(hash, struct{}{})
}

func Delete(hash string) {
	locateMap.Delete(hash)
}

func init() {
	// 读取os.Getenv("STORAGE_ROOT") + "/objects/下所有的object,并将其放入内存中
	filepath.Walk(os.Getenv("STORAGE_ROOT")+"/objects/", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			locateMap.Store(info.Name(), struct{}{})
		}
		return nil
	})
	locateMap.Range(func(key, value interface{}) bool {
		fmt.Println(key)
		return true
	})
}

func locate(name string) bool {
	//_, err := os.Stat(name)
	// return !os.IsNotExist(err)
	_, ok := locateMap.Load(name)
	return ok
}

func StartLocateService() {
	q := rabbitmq.RetryNew(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("dataServers")
	c := q.Consume()
	log.Println("locate service started")
	notifyCloseChan, notifyCancelChan := q.GetNotifyChan()
	// 这两个chan是用来处理rabbitmq的连接断开的情况
	// 一旦连接断开，就会向这两个chan发送消息
	// notifyCloseChan是用来处理正常断开的情况
	// notifyCancelChan是用来处理异常断开的情况
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
			if locate(object) {
				// q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
				// rabbitmq 的rpc模式，这里应该是发送回dataServer的地址
				fmt.Println(msg.ReplyTo)
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
