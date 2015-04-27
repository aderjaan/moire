package handlers

import (
	"strings"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	AudioFile = "audio"
	ImageFile = "image"
	VideoFile = "video"
	PlainFile = "plain"
)

func getConn() *db.MConn {
	return db.GetConn(
		config.Settings.Moire.DBName,
		config.Settings.Moire.DBAddress,
		config.Settings.Moire.DBUsername,
		config.Settings.Moire.DBPassword,
	)
}

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

func updateAsset(conn *db.MConn, _id string, doc db.M) {
	err := conn.Update(db.ASSET, db.M{"_id": bson.ObjectIdHex(_id)}, doc)
	if err != nil {
		panic(err)
	}
}

func getAsset(conn *db.MConn, _id string) *db.Asset {
	var asset db.Asset
	err := conn.GetOne(db.ASSET, db.M{"_id": bson.ObjectIdHex(_id)}, &asset)
	if err != nil {
		panic(err)
	}

	return &asset
}

func assetReady(conn *db.MConn, path, bucket string, doc db.M) *db.Asset {
	var asset db.Asset

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	err := conn.FindAndUpdate(db.ASSET, db.M{
		"path":   path,
		"bucket": bucket,
		"status": db.PENDING,
	}, doc, &asset)

	if err != nil {
		panic(err)
	}

	return &asset
}

func createAsset(conn *db.MConn, args *assetArgs) *db.Asset {
	assetId := bson.NewObjectId()

	asset := db.Asset{
		Id:        assetId,
		CreatedOn: db.EpochNow(),
		Name:      args.Name,
		Bucket:    config.Settings.S3.Bucket,
		FileType:  args.fileType,
		MimeType:  args.MimeType,
		Status:    db.PENDING,
		Path:      getUploadURL(assetId.Hex()),
	}

	conn.Insert(db.ASSET, &asset)

	return &asset
}

func isMimeAllowed(mimeType string) bool {
	if mimeType != ImageFile && mimeType != VideoFile && mimeType != AudioFile {
		return false
	}

	return true
}
