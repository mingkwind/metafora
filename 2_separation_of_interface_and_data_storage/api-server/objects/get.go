package objects

import (
	"api-server/lib/objectstream"
	"api-server/locate"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	log.Println("GET object", r.URL.EscapedPath())
	object := strings.Split(r.URL.EscapedPath(), "/")[2]
	stream, e := getStream(object)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Object not found"))
		return
	}
	io.Copy(w, stream)
}

func getStream(object string) (io.Reader, error) {
	server, err := locate.Locate(object)
	if err != nil {
		return nil, err
	}
	if server == "" {
		return nil, fmt.Errorf("object %s locate fail", object)
	}
	return objectstream.NewGetStream(server, object)
}
