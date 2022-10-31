package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func GetObject(w http.ResponseWriter, r *http.Request) {
	log.Println("GET object", r.URL.EscapedPath())
	file, err := os.Open(path.Join(os.Getenv("STORAGE_ROOT"), "/objects/", strings.Split(r.URL.EscapedPath(), "/")[2]))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
