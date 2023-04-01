package objects

import (
	"data-server/internal/pkg/utils"
	"data-server/internal/service/locate"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetObject(w http.ResponseWriter, r *http.Request) {
	log.Println("GET object", r.URL.EscapedPath())
	name := r.URL.Query().Get("hash")
	//从url中获取文件名
	// 在STROAGE_ROOT/objects/目录下找出以name为前缀的文件
	files, err := filepath.Glob(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", name+".*"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	namestrs := strings.Split(name, ".")
	if len(files) != 1 {
		log.Println("object not found")
		w.WriteHeader(http.StatusNotFound)
		locate.Del(namestrs[0])
		return
	}
	filename := files[0]
	filename_strs := strings.Split(filename, ".")
	hash := filename_strs[2]
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
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
	file, err = os.Open(filename)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 总共要读取两次文件
	io.Copy(w, file)
}
