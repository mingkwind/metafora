package main

import (
	"data-server/heartbeat"
	"data-server/locate"
	"data-server/objects"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/objects/", objects.Handler)
	// http.HandleFunc("/locate/", locate.LocateApi)
	go heartbeat.StartHeartbeatService()
	// go heartbeat.TestComsumer()
	go locate.StartLocateService()
	log.Println("Listening on " + os.Getenv("LISTEN_ADDRESS"))
	err := http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
