package main

import (
	"api-server/internal/controllers"
	"api-server/internal/service/heartbeat"
)

func main() {
	go heartbeat.ListenHeartbeat()
	controllers.Router()
}
