package objects

import (
	"io"
	"log"
	"math/rand"
	"metafora/settings"
	"net/http"
	"os"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getObject(w, r)
	case "PUT":
		putObject(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getRandomString(length int) string {
	if length < 1 {
		return ""
	}
	char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := []byte(char)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func putObject(w http.ResponseWriter, r *http.Request) {
	// 生成十六进制随机文件名 + 后缀
	// 获取 strings.Split(r.URL.EscapedPath() 的后缀
	fileDir := getRandomString(8)
	// 创建文件夹
	err := os.MkdirAll(settings.StorageRoot+fileDir, 0777)
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
	file, err := os.Create(settings.StorageRoot + fileDir + "/" + fileName)
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
	// 返回文件路径
	w.Write([]byte("http://" + settings.ListenAddress + "/" + fileDir + "/" + fileName))
}

func getObject(w http.ResponseWriter, r *http.Request) {
	//  显示当前所在目录
	file, err := os.Open(settings.StorageRoot + r.URL.EscapedPath())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
