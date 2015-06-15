package main

import (
	"log"

	"runtime"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"gopkg.in/simversity/gottp.v3"
)

func sysInit() {
	<-(gottp.SysInitChan) //Buffered Channel to receive the server upstart boolean

	db.Conn = db.GetConn(config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress, config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword)

	log.Println("Database Connected :" + config.Settings.Moire.DBName + " " +
		"at address:" + config.Settings.Moire.DBAddress)

	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.Println("Media Server is ready to serve.")
}

func main() {
	cores := runtime.NumCPU()
	log.Println("Setting no. of Cores as", cores)
	runtime.GOMAXPROCS(cores)

	go sysInit()

	registerHandlers()

	gottp.MakeServer(&config.Settings)
}
