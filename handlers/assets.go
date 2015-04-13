package handlers

import (
	"net/http"
	"strings"

	"github.com/bulletind/moire/db"

	"github.com/bulletind/moire/config"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/simversity/gottp.v2"
	"gopkg.in/simversity/gottp.v2/utils"
)

func ConcatenateErrors(errs *[]error) string {
	var errString string
	for i := 0; i < len(*errs); i++ {
		errString += (*errs)[i].Error()
		if (len(*errs) - i) > 1 {
			errString += ","
		}
	}
	return errString
}

type Assets struct {
	gottp.BaseHandler
}

type assetArgs struct {
	fileType string
	Name     string `json:"name" required:"true"`
	MimeType string `json:"mime_type" required:"true"`
}

func isMimeAllowed(mimeType string) bool {
	if mimeType != ImageFile && mimeType != VideoFile && mimeType != AudioFile {
		return false
	}

	return true
}

func createAsset(conn *db.MConn, args *assetArgs) *db.Asset {
	asset := db.Asset{
		Id:        bson.NewObjectId(),
		CreatedOn: db.EpochNow(),
		Name:      args.Name,
		Bucket:    config.Settings.S3.Bucket,
		FileType:  args.fileType,
		MimeType:  args.MimeType,
		Status:    db.PENDING,
	}

	conn.Insert(db.ASSET, &asset)

	return &asset
}

func (self *Assets) Post(request *gottp.Request) {
	args := assetArgs{}

	request.ConvertArguments(&args)

	err := utils.Validate(&args)
	if len(*err) > 0 {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			ConcatenateErrors(err),
		})

		return
	}

	var fileType string

	if strings.HasPrefix(args.MimeType, ImageFile) {
		fileType = ImageFile
	} else if strings.HasPrefix(args.MimeType, VideoFile) {
		fileType = VideoFile
	} else if strings.HasPrefix(args.MimeType, AudioFile) {
		fileType = AudioFile
	}

	if !isMimeAllowed(fileType) {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			"mime_type " + args.MimeType + " is not supported.",
		})

		return
	}

	args.fileType = fileType

	conn := getConn()
	asset := createAsset(conn, &args)

	_id := asset.Id.Hex()

	request.Write(db.M{
		"_id": _id,
		"url": "/assets/" + _id,
	})

	return
}
