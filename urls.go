package main

import (
	"github.com/bulletind/moire/handlers"
	"gopkg.in/simversity/gottp.v2"
)

func registerHandlers() {
	gottp.NewUrl("assets", "^/assets/?$", new(handlers.Assets))
	gottp.NewUrl("asset", "^/assets/(?P<_id>\\w+)/?$", new(handlers.Asset))
	gottp.NewUrl("thumbnail", "^/assets/(?P<_id>\\w+)/thumbnail/?$", new(handlers.Thumbnail))
}
