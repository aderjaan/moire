package handlers

import (
	"net/http"

	"gopkg.in/simversity/gottp.v2"
)

type Asset struct {
	gottp.BaseHandler
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

	url, err := getURL(asset)
	if err != nil {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			err.Error(),
		})

	} else {
		request.Redirect(url, 302)
	}

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

	signedUrl := getSignedUploadURL(asset.Bucket, asset.Path, asset.MimeType)

	request.Write(signedUrl)
	return
}
