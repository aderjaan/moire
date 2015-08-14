package main

import (
	log "github.com/Sirupsen/logrus"

	"runtime"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/server"
	"gopkg.in/simversity/gottp.v3"
)

func sysInit() {
	<-(gottp.SysInitChan) //Buffered Channel to receive the server upstart boolean
	server.Setup()
}

func main() {
	cores := runtime.NumCPU()
	log.Infoln("Setting no. of Cores as", cores)
	runtime.GOMAXPROCS(cores)

	go sysInit()

	gottp.MakeServer(&config.Settings)
}
