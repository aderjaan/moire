package handlers

import "gopkg.in/simversity/gottp.v2"

type Thumbnail struct {
	gottp.BaseHandler
}

func (self *Thumbnail) Get(request *gottp.Request) {
	//Redirect to the proper thumbnail.
	//Accepts thumbnail sizing.
	//Crete a new image if one does not exist of right size.
}
