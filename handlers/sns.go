package handlers

import "gopkg.in/simversity/gottp.v2"

type SNS struct {
	gottp.BaseHandler
}

func (self *SNS) Post(request *gottp.Request) {
}

type snsMessage struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key  string `json:"key"`
				Size int    `json:"size"`
			} `json:"object"`
		}
	} `json:"Records"`
}

type snsNotice struct {
	Type           string `json:"Type"`
	MessageId      string `json:"MessageId"`
	TopicArn       string `json:"TopicArn"`
	Subject        string `json:"Subject"`
	Message        string `json:"Message"`
	Timestamp      string `json:"Timestamp"`
	Signature      string `json:"Signature"`
	SigningCertURL string `json:"SigningCertURL"`
	UnsubscribeURL string `json:"UnsubscribeURL"`
}
