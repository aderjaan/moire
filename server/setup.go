package server

import (
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"github.com/bulletind/moire/logger"
	golog "log"
)

var log = logger.Logger

func Setup() {
	golog.SetFlags(golog.Ldate | golog.Ltime | golog.Llongfile)

	log.Debug("Connecting Database", "dbname", config.Settings.Moire.DBName, "address", config.Settings.Moire.DBAddress)

	db.Conn = db.GetConn(config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress, config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword)

	log.Debug("Database Connected", "dbname", config.Settings.Moire.DBName, "address", config.Settings.Moire.DBAddress)

	registerHandlers()

	log.Debug("Media Server is ready to serve.")

}
