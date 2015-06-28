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
	return self.server
}

func (self cachedServers) set(server *tests.MockServer) {
	self.Lock()
	self.server = server
	self.Unlock()
}

var cached = cachedServers{}

func MockServer() (server *tests.MockServer) {
	server = cached.get()
	if server == nil {
		server = tests.NewServer()
		registerHandlers()
		cached.set(server)
	}

	return
}

func MockDBServer() (server *tests.MockServer) {
	server = MockServer()

	if !cached.dbInit {
		// This can still run into a repeated run if test are run parallely
		// and both of them decide to Initialize the DB at the same time.
		cached.dbInit = true
		gottp.MakeConfig(&config.Settings)
		config.Settings.Moire.DBName = "gallery-testing"
		config.Settings.Moire.SignRequests = true
		config.Settings.Gottp.EmailDummy = true
		Setup()
	}

	return
}
