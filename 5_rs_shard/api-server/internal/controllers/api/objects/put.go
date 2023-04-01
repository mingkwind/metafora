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
	size := utils.GetSizeFromHeader(r.Header)
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
	c, e := storeObject(r.Body, hash, size)
	if e != nil {
		log.Println(e)
	}
	w.WriteHeader(c)
	if c != http.StatusOK {
		w.Write([]byte(e.Error()))
	} else {
		w.Write([]byte("OK"))
		name := utils.GetFileNameFromRequest(r)
		log.Println("PUT FILE: name:", name, "size:", size, "hash:", hash)
		models.AddVersion(name, size, hash)
	}
}

func putStream(hash string, size int64) (*objectstream.RSPutStream, error) {
	servers := heartbeat.ChooseRandomDataServers(objectstream.ALL_SHARDS, nil)
	fmt.Println(servers)
	if len(servers) != objectstream.ALL_SHARDS {
		return nil, fmt.Errorf("cannot find enough data servers")
	}
	return objectstream.NewRSPutStream(servers, hash, size)
}

// 该函数使用tee同时将数据写入到数据服务器和本地内存中，本地内存用于计算hash值，用于后续的校验
// 如果哈希值为真，就让服务器的temp文件变成正式文件
// 如果哈希值为假，就删除服务器的temp文件
func storeObject(r io.Reader, hash string, size int64) (int, error) {
	// 首先判断哈哈希值是否已经存在
	// 如果存在则直接返回
	// 如果不存在则将数据写入到数据服务器
	if locate.Exist(hash) {
		// 相同hash的文件已经存在
		log.Println("same hash file already exists")
		return http.StatusOK, nil
	}
	stream, e := putStream(hash, size)
	if e != nil {
		return http.StatusServiceUnavailable, e
	}

	// io.Copy(stream, r)
	reader := io.TeeReader(r, stream)
	// io.Wirter原型是
	// type Writer interface {
	// 	Write(p []byte) (n int, err error)
	// }
	// TempPutStream实现了io.Writer接口，因此可以使用
	/// Teereader 会将r的内容拷贝到stream中，同时返回一个reader，这个reader的内容和r一样
	// 相当于将r的内容拷贝了一份，然后计算这个拷贝的内容的hash
	digest := utils.CalculateHash(reader)

	if digest != hash {
		stream.Commit(false)
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, expect %s, get %s", hash, digest)
	} else {
		stream.Commit(true)
		return http.StatusOK, nil
	}
}
