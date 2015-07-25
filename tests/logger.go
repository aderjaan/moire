package tests

import (
	logging "gopkg.in/inconshreveable/log15.v2"
)

var log = logging.New()

func init() {
	log.SetHandler(logging.StdoutHandler)
}
