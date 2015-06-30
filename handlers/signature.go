package handlers

import (
	"net/http"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/signature"
	"gopkg.in/simversity/gottp.v3"
)

func ValidateSignature(request *gottp.Request) bool {
	if config.Settings.Moire.SignRequests != true {
		return true
	}

	public_key, ok := request.GetArgument("public_key").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusPreconditionFailed,
			"public_key is a required parameter",
		})
		return false
	}

	private_key := signature.GetSecretKey(public_key)
	if private_key == "" {
		request.Raise(gottp.HttpError{
			http.StatusForbidden,
			"Invalid public_key supplied",
		})
		return false
	}

	timestamp, ok := request.GetArgument("timestamp").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusPreconditionFailed,
			"timestamp is a required parameter",
		})
		return false
	}

	sign, ok := request.GetArgument("signature").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusPreconditionFailed,
			"signature is a required parameter",
		})
		return false
	}

	sign_error := signature.IsRequestValid(
		public_key,
		private_key,
		timestamp,
		sign,
		request.Request.URL.Path,
	)

	if sign_error != nil {
		request.Raise(gottp.HttpError{http.StatusNotFound, sign_error.Error()})
		return false
	}

	return true
}
