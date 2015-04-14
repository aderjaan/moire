package handlers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bulletind/moire/db"
	"gopkg.in/simversity/gottp.v2"
	"gopkg.in/simversity/gottp.v2/utils"
)

func getAssetByPath(conn *db.MConn, path string) *db.Asset {
	var asset db.Asset
	err := conn.GetOne(db.ASSET, db.M{"path": path}, &asset)
	if err != nil {
		panic(err)
	}

	return &asset
}

type SNS struct {
	gottp.BaseHandler
}

func (self *SNS) Post(request *gottp.Request) {
	var errs *[]error

	n := snsNotice{}
	request.ConvertArguments(&n)

	errs = utils.Validate(&n)
	if len(*errs) > 0 {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			ConcatenateErrors(errs),
		})

		return
	}

	msg := snsMessage{}
	utils.Decoder([]byte(n.Message), &msg)

	errs = utils.Validate(&msg)
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

	log.Println(key, record.S3.Object.Size)
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
	Type      string `json:"Type"`
	MessageId string `json:"MessageId" required:"true"`
	TopicArn  string `json:"TopicArn"`
	Subject   string `json:"Subject"`
	Message   string `json:"Message" required:"true"`
	Timestamp string `json:"Timestamp"`
	Signature string `json:"Signature" required:"true"`
}
