package main

import (
	"data-server/internal/controllers"
	"data-server/internal/service/heartbeat"
	"data-server/internal/service/locate"
)

func main() {
	go heartbeat.StartHeartbeatService()
	go locate.StartLocateService()
	controllers.Router()
}
