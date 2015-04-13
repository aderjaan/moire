package handlers

import (
	"net/http"

	"github.com/bulletind/moire/db"

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
	Name     string `json:"name" required:"true"`
	MimeType string `json:"mime_type" required:"true"`
}

func createAsset(conn *db.MConn, args *assetArgs) *db.Asset {
	asset := db.Asset{
		Id:        bson.NewObjectId(),
		CreatedOn: db.EpochNow(),
		Name:      args.Name,
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

	conn := getConn()
	asset := createAsset(conn, &args)

	request.Write(db.M{
		"_id": asset.Id,
		"url": "/assets/" + asset.Id,
	})

	return
}
