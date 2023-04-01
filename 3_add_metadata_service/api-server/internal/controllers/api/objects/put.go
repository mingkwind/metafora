package objects

import (
	"api-server/internal/controllers/api/locate"
	"api-server/internal/models"
	"api-server/internal/pkg/objectstream"
	"api-server/internal/pkg/utils"
	"api-server/internal/service/heartbeat"
	"fmt"
	"io"
	"log"
	"net/http"
)

func PutObjects(w http.ResponseWriter, r *http.Request) {
	log.Println("PUT object", r.URL.EscapedPath())
	hash := utils.GetHashFromHeader(r.Header)
	// 首先判断文件是否存在
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("hash is empty"))
		return
	}
	name := utils.GetFileNameFromRequest(r)
	metadata, err := models.GetLatestVersion(name)
	if err != nil {
		log.Println("GetLatestVersion ERROR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("GetLatestVersion ERROR"))
		return
	}
	if metadata.Hash == hash {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("file already exists"))
		return
	}
	c, e := storeObject(r.Body, hash)
	if e != nil {
		log.Println(e)
	}
	w.WriteHeader(c)
	if c != http.StatusOK {
		w.Write([]byte(e.Error()))
	} else {
		w.Write([]byte("OK"))
		size := utils.GetSizeFromHeader(r.Header)
		log.Println("PUT FILE: name:", name, "size:", size, "hash:", hash)
		models.PutMetadata(&models.Metadata{
			Name: name,
			Size: size,
			Hash: hash,
		})
	}
}

func putStream(hash string) (*objectstream.PutStream, error) {
	server := heartbeat.ChooseRandomDataServer()
	if server == "" {
		return nil, fmt.Errorf("cannot find any dataServer")
	}

	return objectstream.NewPutStream(server, hash), nil
}

func storeObject(r io.Reader, hash string) (int, error) {
	// 首先判断哈哈希值是否已经存在
	// 如果存在则直接返回
	// 如果不存在则将数据写入到数据服务器
	info, err := locate.Locate(hash)
	if err == nil && len(info) != 0 {
		// 相同hash的文件已经存在
		log.Println("same hash file already exists")
		return http.StatusOK, nil
	}
	stream, e := putStream(hash)
	if e != nil {
		return http.StatusServiceUnavailable, e
	}
	io.Copy(stream, r)
	e = stream.Close()
	if e != nil {
		return http.StatusInternalServerError, e
	}
	return http.StatusOK, nil
}
