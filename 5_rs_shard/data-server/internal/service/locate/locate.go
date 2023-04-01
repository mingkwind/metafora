package locate

import (
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"data-server/internal/pkg/rabbitmq"
	"path/filepath"
)

var (
	locateMap = sync.Map{}
)

type locateMessage struct {
	Hash string `json:"hash"`
	Id   int    `json:"id"`
	Addr string `json:"addr"`
}

func init() {
	// 判断os.Getenv("STORAGE_ROOT")是否存在，不存在则创建
	if _, err := os.Stat(os.Getenv("STORAGE_ROOT")); os.IsNotExist(err) {
		os.MkdirAll(os.Getenv("STORAGE_ROOT"), 0777)
	}
	// 判断os.Getenv("STORAGE_ROOT"), "/objects/"是否存在，不存在则创建
	if _, err := os.Stat(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/")); os.IsNotExist(err) {
		os.MkdirAll(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/"), 0777)
	}
	// 判断os.Getenv("STORAGE_ROOT"), "/temp/"是否存在，不存在则创建
	if _, err := os.Stat(path.Join(os.Getenv("STORAGE_ROOT"), "/temp/")); os.IsNotExist(err) {
		os.MkdirAll(path.Join(os.Getenv("STORAGE_ROOT"), "/temp/"), 0777)
	}
	// 遍历所有的文件，将文件的hash值和文件的路径存储到locateMap中
	filepath.Walk(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fileName := strings.Split(info.Name(), ".")
			id, _ := strconv.Atoi(fileName[1])
			locateMap.Store(fileName[0], id)
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

func Add(hash string, id int) {
	locateMap.Store(hash, id)
}

func Del(hash string) {
	locateMap.Delete(hash)
}

func locate(name string) int {
	id, ok := locateMap.Load(name)
	if ok {
		return id.(int)
	} else {
		return -1
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
			hash, err := strconv.Unquote(string(msg.Body))
			if err != nil {
				log.Println("locate service: strconv.Unquote error: ", err)
				continue
			}
			id := locate(hash)
			if id != -1 {
				q.Send(msg.ReplyTo, locateMessage{Addr: os.Getenv("DATA_SERVER_ADDRESS"), Id: id})
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
