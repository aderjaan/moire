package handlers

import (
	"net/http"
	"time"

	"github.com/bulletind/moire/db"
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

	valid := ValidateSignature(request)
	if valid == false {
		return
	}

	conn := getConn()
	asset, _ := getAsset(conn, _id)

	request.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	_, no_redirect := request.GetArgument("no_redirect").(string)

	if asset.Status != db.READY && asset.FileType == ImageFile {
		pollUntilReady(conn, _id)
		asset, _ = getAsset(conn, _id)
	}

	// This needs explanation and refactoring
	url, expiryDate, err := getURL(asset)

	if err != nil {
		if asset.FileType == ImageFile && no_redirect != true {
			// If its an Image and no_redirect is not provided, error is ir-relevant.
			// It will always return the placeholder image anyway.
			// we don't want the expires caching header to be set, as we expect the asset to be uploaded soon.
			//
			// Why we don't show a 404 for images just like with other assets or show a thumbnail for these other assets as well,
			// remains a mistery.
			request.Redirect(url, TemporaryRedirectCode)
			return
		}

		// show a 404 error for other assets
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			err.Error(),
		})

		return
	}

	// aws signing key expiration date and will be used to set relevant caching periods accordingly
	request.Writer.Header().Set("Expires", expiryDate.Format(time.RFC1123))
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
	asset, _ := getAsset(conn, _id)

	signedUrl := getSignedUploadURL(asset.Bucket, asset.Path, asset.MimeType)

	request.Write(signedUrl)
	return
}
