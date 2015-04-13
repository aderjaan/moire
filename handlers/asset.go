package handlers

import "gopkg.in/simversity/gottp.v2"

type Asset struct {
	gottp.BaseHandler
}

func (self *Asset) Get(request *gottp.Request) {
	//Redirect to the proper Image URI
}
