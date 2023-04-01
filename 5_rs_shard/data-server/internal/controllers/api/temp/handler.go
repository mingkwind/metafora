package temp

import (
	"data-server/internal/pkg/utils"
	"data-server/internal/service/locate"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
)

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
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		PostTemp(w, r)
	case "PATCH":
		PatchTemp(w, r)
	case "PUT":
		PutTemp(w, r)
	case "DELETE":
		DeleteTemp(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// 小写字母无法序列化
type tempInfo struct {
	Uuid string `json:"uuid"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Id   int    `json:"id"`
}

func (t *tempInfo) WriteToFile() error {
	// 将该结构体序列化为json写入到文件中
	file, err := os.Create(path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", t.Uuid))
	if err != nil {
		return err
	}
	defer file.Close()

	// 将结构体t序列化为json
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(t)
	if err != nil {
		log.Println(err)
	}
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil

}

func PostTemp(w http.ResponseWriter, r *http.Request) {
	// http://ip:port/temp/xxxxx
	// 从请求头中获取文件名和文件大小
	//从url中读取hash值,hash是路径的最后一部分
	// 从请求头中Content-Length读取size
	// 这样同时生成了两个文件，一个是uuid，一个是uuid.dat
	name := strings.Split(strings.Split(r.URL.EscapedPath(), "/")[2], ".")
	id, _ := strconv.Atoi(name[1])
	size, err := strconv.ParseInt(r.Header.Get("Size"), 0, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	t := tempInfo{
		Uuid: uuid.NewV4().String(),
		Hash: name[0],
		Size: size,
		Id:   id,
	}
	t.WriteToFile()
	file, err := os.Create(path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", t.Uuid+".dat"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	w.Write([]byte(t.Uuid))
}

func readFromFile(uuid string) (*tempInfo, error) {
	file, err := os.Open(path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", uuid))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 读取文件内容
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b := make([]byte, 1024)
	n, err := file.Read(b)
	if err != nil {
		return nil, err
	}
	t := &tempInfo{}
	err = json.Unmarshal(b[:n], t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func PatchTemp(w http.ResponseWriter, r *http.Request) {
	// 读取url中的uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, err := readFromFile(uuid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", uuid)
	dataFile := infoFile + ".dat"
	// 写入文件到dataFile
	file, err := os.OpenFile(dataFile, os.O_APPEND|os.O_WRONLY, 0644)
	// os.O_APPEND|os.O_WRONLY表示以追加的方式打开文件
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// 读取请求体
	_, err = io.Copy(file, r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 写完后查看文件大小
	fileInfo, err := os.Stat(dataFile)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileInfo.Size() != tempinfo.Size {
		// 将文件都删掉
		os.Remove(infoFile)
		os.Remove(dataFile)
		log.Println("actual size:", fileInfo.Size(), "expected size:", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func PutTemp(w http.ResponseWriter, r *http.Request) {
	// 读取url中的uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, err := readFromFile(uuid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", uuid)
	dataFile := infoFile + ".dat"
	// 判断文件大小
	fileInfo, err := os.Stat(dataFile)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if fileInfo.Size() != tempinfo.Size {
		// 将文件都删掉
		os.Remove(infoFile)
		os.Remove(dataFile)
		log.Println("actual size:", fileInfo.Size(), "expected size:", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 将文件转正
	os.Remove(infoFile)
	commitTempObject(dataFile, tempinfo)
}

func commitTempObject(dataFile string, tempInfo *tempInfo) {
	f, _ := os.Open(dataFile)
	defer f.Close()
	// 计算哈希值
	hashOfX := utils.CalculateHash(f)
	os.Rename(dataFile, path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", fmt.Sprintf("%s.%d.%s", tempInfo.Hash, tempInfo.Id, hashOfX)))
	log.Println("commit temp object:", tempInfo.Uuid)
	locate.Add(tempInfo.Hash, tempInfo.Id)
}

func DeleteTemp(w http.ResponseWriter, r *http.Request) {
	// 读取url中的uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "/temp/", uuid)
	dataFile := infoFile + ".dat"
	os.Remove(infoFile)
	os.Remove(dataFile)
	log.Println("delete temp object", uuid)
}
