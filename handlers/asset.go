package handlers

import (
	"net/http"

	"gopkg.in/simversity/gottp.v3"
)

const TemporaryRedirectCode = 302

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

	_, no_redirect := request.GetArgument("no_redirect").(string)

	url, err := getURL(asset, no_redirect)
	if err != nil {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			err.Error(),
		})

		return
	}

	valid := ValidateSignature(request)
	if valid == false {
		return
	}

	request.Redirect(url, TemporaryRedirectCode)
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
