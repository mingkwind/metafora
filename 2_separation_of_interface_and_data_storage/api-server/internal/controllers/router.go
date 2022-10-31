package controllers

import (
	"api-server/internal/controllers/api/locate"
	"api-server/internal/controllers/api/objects"
	"log"
	"net/http"
	"os"
)

func Router() {
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	log.Println("Listening on " + os.Getenv("LISTEN_ADDRESS"))
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
