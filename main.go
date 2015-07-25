package main

import (
	"runtime"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/logger"
	"github.com/bulletind/moire/server"
	"gopkg.in/simversity/gottp.v3"
)

var log = logger.Logger

func sysInit() {
	<-(gottp.SysInitChan) //Buffered Channel to receive the server upstart boolean
	server.Setup()
}

func main() {
	cores := runtime.NumCPU()
	log.Debug("Setting", "cores", cores)
	runtime.GOMAXPROCS(cores)

	go sysInit()

	gottp.MakeServer(&config.Settings)
}
