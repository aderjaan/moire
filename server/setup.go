package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/sentry"
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
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
	log.Debug("Connecting Database:" + config.Settings.Moire.DbUrl)

	db.Conn = db.GetConn(config.Settings.Moire.DbUrl, config.Settings.Moire.DbName)

	log.Println("Database Connected :" + config.Settings.Moire.DbName)
}
