package objects

import (
	"data-server/internal/pkg/utils"
	"data-server/internal/service/locate"
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
	// 计算该文件的hash值
	digest := utils.CalculateHash(file)
	if digest != hash {
		log.Println("hash mismatch")
		w.WriteHeader(http.StatusNotFound)
		// 删除文件
		os.Remove(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", hash))
		locate.Del(hash)
		return
	}
	file, err = os.Open(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", hash))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 总共要读取两次文件
	io.Copy(w, file)
}
