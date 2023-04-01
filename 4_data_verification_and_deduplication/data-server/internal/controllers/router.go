package controllers

import (
	"data-server/internal/controllers/api/objects"
	"data-server/internal/controllers/api/temp"
	"log"
	"net/http"
	"os"
)

func Router() {
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	log.Println("Listening on " + os.Getenv("LISTEN_ADDRESS"))
	err := http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
