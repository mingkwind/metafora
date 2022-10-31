package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func GetObject(w http.ResponseWriter, r *http.Request) {
	log.Println("GET object", r.URL.EscapedPath())
	//从url中获取hash值
	hash := r.URL.Query().Get("hash")
	//从url中获取文件名
	file, err := os.Open(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", hash))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
