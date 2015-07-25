package server

import (
	"github.com/bulletind/moire/config"
	"gopkg.in/simversity/gottp.v3"
	"gopkg.in/simversity/gottp.v3/tests"
	"os"
	"sync"
)

type cachedServers struct {
	sync.RWMutex
	dbInit bool
	server *tests.MockServer
}

func (self cachedServers) get() *tests.MockServer {
	self.Lock()
	if self.server == nil {
		self.server = tests.NewServer()
		registerHandlers()
	}
	self.Unlock()
	return self.server
}

var cached = cachedServers{}

func MockDBServer() (server *tests.MockServer) {
	server = cached.get()

	if !cached.dbInit {
		// This can still run into a repeated run if test are run parallely
		// and both of them decide to Initialize the DB at the same time.
		cached.dbInit = true
		gottp.MakeConfig(&config.Settings)

		moire_host := os.Getenv("MOIRE_HOST")
		moire_username := os.Getenv("MOIRE_USERNAME")
		moire_password := os.Getenv("MOIRE_PASSWORD")

		if moire_host != "" {
			config.Settings.Moire.DBAddress = moire_host
		}

		if moire_username != "" {
			config.Settings.Moire.DBUsername = moire_username
		}

		if moire_password != "" {
			config.Settings.Moire.DBPassword = moire_password
		}

		config.Settings.Moire.DBName = "gallery-testing"
		config.Settings.Moire.SignRequests = true
		config.Settings.Gottp.EmailDummy = true
		config.Settings.Moire.ImageTimeout = 5
		Setup()
	}

	return
}
