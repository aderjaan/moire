package server

import (
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"log"
)

func Setup() {
	log.Println("Connecting Database:" + config.Settings.Moire.DBName + " " +
		"at address:" + config.Settings.Moire.DBAddress)

	db.Conn = db.GetConn(config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress, config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword)

	log.Println("Database Connected :" + config.Settings.Moire.DBName + " " +
		"at address:" + config.Settings.Moire.DBAddress)

	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	registerHandlers()

	log.Println("Media Server is ready to serve.")

}
