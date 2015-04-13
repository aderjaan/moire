package handlers

import (
	"net/http"

	"github.com/bulletind/moire/db"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/simversity/gottp.v2"
)

type Asset struct {
	gottp.BaseHandler
}

func getAsset(conn *db.MConn, _id string) *db.Asset {
	var asset db.Asset
	err := conn.GetOne(db.ASSET, db.M{"_id": bson.ObjectIdHex(_id)}, &asset)
	if err != nil {
		panic(err)
	}

	return &asset
}

func (self *Asset) Get(request *gottp.Request) {
	_id, ok := request.GetArgument("_id").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			"Not found",
		})

		return
	}

	conn := getConn()
	asset := getAsset(conn, _id)
	request.Write(asset)
	return
}

func (self *Asset) Post(request *gottp.Request) {
	_id, ok := request.GetArgument("_id").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			"Not found",
		})

		return
	}

	conn := getConn()
	asset := getAsset(conn, _id)
	request.Write(asset)
	return
}
