package objects

import (
	"api-server/internal/controllers/api/locate"
	"api-server/internal/models"
	"api-server/internal/pkg/objectstream"
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
	stream, e := getStream(metadata.Hash)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	// 设置响应头的文件名
	// w.Header().Set("Content-Disposition", "attachment; filename=\""+object+"\"")
	io.Copy(w, stream)
}

func getStream(hash string) (io.Reader, error) {
	// hash 就作为这个文件的文件名
	server, err := locate.Locate(hash)
	if err != nil {
		return nil, err
	}
	if server == "" {
		return nil, fmt.Errorf("hash %s locate fail", hash)
	}
	return objectstream.NewGetStream(server, hash)
}
