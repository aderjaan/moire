package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bulletind/moire/db"
	"gopkg.in/simversity/gottp.v3"
	"gopkg.in/simversity/gottp.v3/utils"
)

const SNS_SUBSCRIPTIONCONFIRMATION_TYPE = "SubscriptionConfirmation"

type SNS struct {
	gottp.BaseHandler
}

func parseMessage(n snsNotice) (msg snsMessage, errs *[]error) {
	errs = utils.Validate(&n)
	if len(*errs) > 0 {
		return
	}

	utils.Decoder([]byte(n.Message), &msg)

	errs = utils.Validate(&msg)
	return
}

func breakOnSubscriptionConfirmation(n snsNotice) {
	if n.Type == SNS_SUBSCRIPTIONCONFIRMATION_TYPE {
		msg := fmt.Sprintf("Subscription Confirmation Request, your SubscribeURL: %s", n.SubscribeURL)
		panic(msg)
	}
}

func (self *SNS) Post(request *gottp.Request) {
	var errs *[]error

	n := snsNotice{}
	request.ConvertArguments(&n)

	breakOnSubscriptionConfirmation(n)

	msg, errs := parseMessage(n)
	if len(*errs) > 0 {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			ConcatenateErrors(errs),
		})

		return
	}

	record := msg.Records[0]

	key, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			"Cannot escape Path",
		})

		return
	}

	if !strings.Contains(key, UploadPrefix) {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			"Skipping path as its not meant to be monitored.",
		})

		return
	}

	doc := db.M{"size": record.S3.Object.Size}

	conn := getConn()
	asset := assetReady(conn, key, record.S3.Bucket.Name, db.M{"$set": doc})

	assetId := asset.Id.Hex()

	thumbnailPath := getThumbnailUploadURL(assetId, asset.Collection, asset.Name)

	if asset.FileType == VideoFile {

		thumbPath := videoThumbnail(asset, thumbTime, thumbX*2, thumbY*2)
		optimizeThumbnail(thumbPath)

		uploadFile(asset.Bucket, thumbnailPath, thumbPath)
		updateAsset(
			conn, assetId,
			db.M{"$set": db.M{"thumbnail_path": thumbnailPath, "status": db.READY}},
		)

		cleanupThumbnail(thumbPath)

	} else if asset.FileType == ImageFile {

		thumbPath := imageThumbnail(asset, thumbX, thumbY)
		optimizeThumbnail(thumbPath)

		uploadFile(asset.Bucket, thumbnailPath, thumbPath)
		updateAsset(
			conn, assetId,
			db.M{"$set": db.M{"thumbnail_path": thumbnailPath, "status": db.READY}},
		)

		cleanupThumbnail(thumbPath)

	} else {
		updateAsset(conn, assetId, db.M{"$set": db.M{"status": db.READY}})
	}

	request.Write("asset " + assetId + " marked as ready")
}

type snsMessage struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name" required:"true"`
			} `json:"bucket" required:"true"`
			Object struct {
				Key  string `json:"key" required:"true"`
				Size int    `json:"size" required:"true"`
			} `json:"object" required:"true"`
		}
	} `json:"Records" required:"true"`
}

type snsNotice struct {
	Type         string `json:"Type"`
	MessageId    string `json:"MessageId" required:"true"`
	TopicArn     string `json:"TopicArn"`
	Subject      string `json:"Subject"`
	Message      string `json:"Message" required:"true"`
	Timestamp    string `json:"Timestamp"`
	Signature    string `json:"Signature" required:"true"`
	SubscribeURL string `json:"SubscribeURL"`
}
