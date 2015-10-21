package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"github.com/evalphobia/logrus_sentry"
)

func Setup() {
	setupLogging()

	setupDatabase()

	registerHandlers()

	log.Println("Media Server is ready to serve.")

}

func setupLogging() {
	if config.Settings.Moire.Debug {
		log.Info("Logging in DEBUG mode")
		log.SetLevel(log.DebugLevel)
	}

	// log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.SetFormatter(&log.TextFormatter{})

	if config.Settings.Moire.SentryDSN != "" {
		levels := []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
		}

		hook, err := logrus_sentry.NewSentryHook(config.Settings.Moire.SentryDSN, levels)
		if err != nil {
			log.Error("Unable to connect to sentry")
		} else {
			log.Info("Adding Sentry Hook")
			log.AddHook(hook)
		}
	}
}

func setupDatabase() {
	log.Println("Connecting Database:" + config.Settings.Moire.DBName + " " +
		"at address:" + config.Settings.Moire.DBAddress)

	db.Conn = db.GetConn(config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress, config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword)

	log.Println("Database Connected :" + config.Settings.Moire.DBName + " " +
		"at address:" + config.Settings.Moire.DBAddress)
}
