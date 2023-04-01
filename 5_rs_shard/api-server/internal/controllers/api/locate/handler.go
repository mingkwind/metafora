package locate

import (
	"api-server/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("GET locate request: ", r.URL.EscapedPath())
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	filename := strings.Split(r.URL.EscapedPath(), "/")[2]
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
	metadata, err := models.GetMetadata(filename, version)
	// 如果metadata为size为空，且hash为空，说明元数据不存在
	if err != nil || metadata.Size == 0 && metadata.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	// 在消息队列发送消息，获取数据所在的服务器
	info := Locate(metadata.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(info) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(info)
	w.Write(b)
}
