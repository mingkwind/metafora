package objects

import (
	"data-server/internal/pkg/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

/*
func putObject(w http.ResponseWriter, r *http.Request) {
	// 生成十六进制随机文件名 + 后缀
	// 获取 strings.Split(r.URL.EscapedPath() 的后缀
	fileDir := utils.GetRandomString(8)
	// 创建文件夹
	err := os.MkdirAll(os.Getenv("STORAGE_ROOT")+fileDir, 0777)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 从url中获取文件名
	fileName := strings.Split(r.URL.EscapedPath(), "/")[1]
	// 如果文件名为空，则从请求头中获取文件名
	if fileName == "" {
		fileName = r.Header.Get("filename")
	}
	file, err := os.Create(os.Getenv("STORAGE_ROOT") + fileDir + "/" + fileName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	host := os.Getenv("LISTEN_ADDRESS")
	// 如果没有ip只有端口，则使用默认ip
	if strings.Split(host, ":")[0] == "" {
		host = "127.0.0.1" + ":" + strings.Split(host, ":")[1]
	}
	// 返回文件路径
	w.Write([]byte("http://" + host + "/" + fileDir + "/" + fileName))
}x
*/

func init() {
	// 判断os.Getenv("STORAGE_ROOT")是否存在，不存在则创建
	if _, err := os.Stat(os.Getenv("STORAGE_ROOT")); os.IsNotExist(err) {
		os.MkdirAll(os.Getenv("STORAGE_ROOT"), 0777)
	}
	// 判断os.Getenv("STORAGE_ROOT"), "/objects/"是否存在，不存在则创建
	if _, err := os.Stat(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/")); os.IsNotExist(err) {
		os.MkdirAll(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/"), 0777)
	}
}

func putObject(w http.ResponseWriter, r *http.Request) {
	log.Println("PUT Object: ", r.URL.EscapedPath())
	// 从headr中获取hash值
	hash := utils.GetHashFromHeader(r.Header)
	log.Println("hash:", hash)
	filePath := path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", hash)
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	io.Copy(file, r.Body)
	// 返回文件路径
	w.Write([]byte("Save OK"))
}
