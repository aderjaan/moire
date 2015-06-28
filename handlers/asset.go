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

	valid := ValidateSignature(request)
	if valid == false {
		return
	}

	_, no_redirect := request.GetArgument("no_redirect").(string)

	url, err := getURL(asset)

	if asset.FileType == ImageFile && no_redirect != true {
		// If its an Image and no_redirect is not provided, error is ir-relevant.
		// It will always return the default image anyway.
	} else if err != nil {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			err.Error(),
		})

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
