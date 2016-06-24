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

func (self *Assets) Post(request *gottp.Request) {
	args, err := convertRequest(request)
	if len(*err) > 0 {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			ConcatenateErrors(err),
		})
		return
	}

	var asset *db.Asset
	conn := getConn()

	// existing one - pending
	if len(args.Id) > 0 {
		asset = getAsset(conn, args.Id)
		if asset != nil && asset.Status != db.PENDING {
			request.Raise(gottp.HttpError{
				http.StatusForbidden,
				"Asset has no pending state",
			})
			return
		}
	}

	// new asset
	if asset == nil {
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
	fileType := PlainFile
	if strings.HasPrefix(mimeType, ImageFile) {
		fileType = ImageFile
	} else if strings.HasPrefix(mimeType, VideoFile) {
		fileType = VideoFile
	} else if strings.HasPrefix(mimeType, AudioFile) {
		fileType = AudioFile
	}
	return fileType
}
