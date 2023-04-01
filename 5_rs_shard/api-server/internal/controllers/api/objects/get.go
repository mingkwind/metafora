package objects

import (
	"api-server/internal/controllers/api/locate"
	"api-server/internal/models"
	"api-server/internal/pkg/objectstream"
	"api-server/internal/service/heartbeat"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetObjects(w http.ResponseWriter, r *http.Request) {
	log.Println("GET object", r.URL.EscapedPath())
	object := strings.Split(r.URL.EscapedPath(), "/")[2]
	versionId := r.URL.Query().Get("version")
	version := 0
	if versionId != "" {
		var err error
		version, err = strconv.Atoi(versionId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	// 从es中获取元数据
	metadata, err := models.GetMetadata(object, version)
	// 如果metadata为size为空，且hash为空，说明元数据不存在
	if err != nil || metadata.Size == 0 && metadata.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	stream, e := getStream(metadata.Hash, metadata.Size)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	// 设置响应头的文件名
	// w.Header().Set("Content-Disposition", "attachment; filename=\""+object+"\"")
	w.Header().Set("Digest", "SHA-256="+metadata.Hash)
	_, err = io.Copy(w, stream)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream.Close()

}

func getStream(hash string, size int64) (*objectstream.RSGetStream, error) {
	locateInfo := locate.Locate(hash)
	if len(locateInfo) < objectstream.DATA_SHARDS {
		return nil, fmt.Errorf("hash %s locate fail, result %v", hash, locateInfo)
	}
	dataServers := make([]string, 0)
	if len(locateInfo) != objectstream.ALL_SHARDS {
		// 说明有部分分片丢失，选取剩下的节点对分片进行恢复
		dataServers = heartbeat.ChooseRandomDataServers(objectstream.ALL_SHARDS-len(locateInfo), locateInfo)
	}
	return objectstream.NewRSGetStream(locateInfo, dataServers, hash, size)
}
