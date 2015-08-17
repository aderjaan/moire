package handlers

import (
	log "github.com/Sirupsen/logrus"

	"strings"
	"time"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"gopkg.in/mgo.v2/bson"
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

func pollUntilReady(conn *db.MConn, _id string) {
	timeout := time.After(time.Duration(config.Settings.Moire.ImageTimeout) * time.Second)
	tick := time.Tick(1000 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			log.Debugln("Polling timed out")
			return
			// Got a tick, we should check on doSomething()
		case <-tick:
			asset := getAsset(conn, _id)
			ok := asset.Status == db.READY
			if ok {
				return
			}
		}
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

	if args.Collection == "" {
		args.Collection = DefaultCollection
	}

	asset := db.Asset{
		Id:         assetId,
		CreatedOn:  db.EpochNow(),
		Name:       args.Name,
		Bucket:     config.Settings.S3.Bucket,
		FileType:   args.fileType,
		MimeType:   args.MimeType,
		Status:     db.PENDING,
		Collection: args.Collection,
		Path:       getUploadURL(assetId.Hex(), args.Collection, args.Name),
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
