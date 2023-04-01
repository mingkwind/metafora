package objects

import (
	"api-server/internal/models"
	"log"
	"net/http"
	"strings"
)

func DeleteObjects(w http.ResponseWriter, r *http.Request) {
	log.Println("DELETE object", r.URL.EscapedPath())
	object := strings.Split(r.URL.EscapedPath(), "/")[2]
	err := delete(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("DELETE Success"))
	}
}

func delete(object string) error {
	return models.PutMetadata(&models.Metadata{
		Name: object,
		Size: 0,
		Hash: "",
	})
}
