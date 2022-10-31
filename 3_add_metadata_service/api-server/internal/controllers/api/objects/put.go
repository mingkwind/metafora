package objects

import (
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
	c, e := storeObject(r.Body, hash)
	if e != nil {
		log.Println(e)
	}
	w.WriteHeader(c)
	if c != http.StatusOK {
		w.Write([]byte(e.Error()))
	} else {
		w.Write([]byte("OK"))
		name := utils.GetFileNameFromRequest(r)
		size := utils.GetSizeFromHeader(r.Header)
		hash := utils.GetHashFromHeader(r.Header)
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
