package server

import (
	"github.com/bulletind/moire/config"
	"gopkg.in/simversity/gottp.v3"
	"gopkg.in/simversity/gottp.v3/tests"
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

		config.Settings.Moire.DbName = "gallery-testing"
		config.Settings.Moire.SignRequests = true
		config.Settings.Gottp.EmailDummy = true
		config.Settings.Moire.ImageTimeout = 5
		Setup()
	}

	return
}
