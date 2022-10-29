package main

import (
	"log"
	"metafora/objects"
	"metafora/settings"
	"net/http"
)

func main() {
	http.HandleFunc("/", objects.Handler)
	log.Println("Listening on", settings.ListenAddress)
	log.Fatal(http.ListenAndServe(settings.ListenAddress, nil))
}
