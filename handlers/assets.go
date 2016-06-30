package handlers

import (
	"net/http"
	"strings"

	"github.com/bulletind/moire/db"

	"gopkg.in/simversity/gottp.v3"
	"gopkg.in/simversity/gottp.v3/utils"
)

type Assets struct {
	gottp.BaseHandler
}

type assetArgs struct {
	fileType   string
	Id         string `json:"_id" required:"false"`
	Name       string `json:"name" required:"true"`
	MimeType   string `json:"mime_type" required:"true"`
	Collection string `json:"collection"`
}

func (args assetArgs) HasId() bool {
	return len(args.Id) > 0
}

func (self *Assets) Post(request *gottp.Request) {
	args, err := convertRequest(request)
	if len(*err) > 0 {
		throwError(request, http.StatusBadRequest, ConcatenateErrors(err))
		return
	}

	var asset *db.Asset
	var createNew = !args.HasId()
	conn := getConn()

	// existing one - does it exist and is it pending?
	if !createNew {
		var err error
		asset, err = getAsset(conn, args.Id)
		if err != nil {
			createNew = true
		} else if asset.Status != db.PENDING {
			throwError(request, http.StatusForbidden, "Asset has no pending state")
			return
		}
	}

	// new asset
	if createNew {
		args.fileType = getFileType(args.MimeType)
		asset = createAsset(conn, &args)
	}

	request.Write(db.M{
		"_id":        asset.Id.Hex(),
		"url":        "/assets/" + asset.Id.Hex(),
		"upload_url": getSignedUploadURL(asset.Bucket, asset.Path, asset.MimeType),
	})

	return
}

func convertRequest(request *gottp.Request) (assetArgs, *[]error) {
	args := assetArgs{}
	request.ConvertArguments(&args)

	err := utils.Validate(&args)
	return args, err
}

func getFileType(mimeType string) string {
	fileTypes := []string{ImageFile, VideoFile, AudioFile}
	for _, fileType := range fileTypes {
		if strings.HasPrefix(mimeType, fileType) {
			return fileType
		}
	}
	return PlainFile
}

func throwError(request *gottp.Request, status int, message string) {
	request.Raise(gottp.HttpError{
		status,
		message,
	})
}
