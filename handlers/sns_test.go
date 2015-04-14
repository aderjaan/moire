package handlers

import (
	"testing"

	"gopkg.in/simversity/gottp.v2/utils"
)

func TestSNSNoticeFail(t *testing.T) {
	jsonStr := `{
		"Type" : "Notification",
		"MessageId" : "f6cdc099-1020-554c-9f24-8d3f8067f821",
		"TopicArn" : "arn:aws:sns:eu-west-1:646538960088:gallery-upload",
		"Subject" : "Amazon S3 Notification"
	}`

	n := snsNotice{}
	utils.Decoder([]byte(jsonStr), &n)

	errs := utils.Validate(&n)
	if len(*errs) == 0 {
		t.Error("Message should be absent.")
		return
	}
}

func TestSNSNotice(t *testing.T) {
	jsonStr := `
	{
		"Type" : "Notification",
		"MessageId" : "f6cdc099-1020-554c-9f24-8d3f8067f821",
		"TopicArn" : "arn:aws:sns:eu-west-1:646538960088:gallery-upload",
		"Subject" : "Amazon S3 Notification",
		"Message" : "{\"Records\":[{\"eventVersion\":\"2.0\",\"eventSource\":\"aws:s3\",\"awsRegion\":\"eu-west-1\",\"eventTime\":\"2015-04-14T03:48:23.469Z\",\"eventName\":\"ObjectCreated:Put\",\"userIdentity\":{\"principalId\":\"AWS:AIDAJUDYYAFFWAKERTPSU\"},\"requestParameters\":{\"sourceIPAddress\":\"10.3.77.81\"},\"responseElements\":{\"x-amz-request-id\":\"5EFEAB74AFBE71DD\",\"x-amz-id-2\":\"xifX0V/e13Sq2C4BoIFcw+zGre8HMBaLB0bTpGB4uH6tbdtBbNrc6eVbopTfSa5D\"},\"s3\":{\"s3SchemaVersion\":\"1.0\",\"configurationId\":\"GalleryNotifications\",\"bucket\":{\"name\":\"sc-gallery\",\"ownerIdentity\":{\"principalId\":\"A318KGBVB9LIDC\"},\"arn\":\"arn:aws:s3:::sc-gallery\"},\"object\":{\"key\":\"image/Screen+Shot+2015-04-10+at+20.41.20.PNG\",\"size\":71501,\"eTag\":\"70e3e31e0d129d58cb925f6b834a477b\"}}}]}",
		"Timestamp" : "2015-04-14T03:48:23.584Z",
		"SignatureVersion" : "1",
		"Signature" : "liP1M+gnXDSo5A4mZJ/lO8Ah0rsC1ThfU0cmU5QmLezGB/VRq5G9V1QObO5phohsWLhMZiNTLVDWe9KCg9zKx+/X1S880Ytjd+Dyj4y1G29zATG3hzuRI1Ernp0dqHyIMwvbLrh6mqge65EPA/dzWVUjIehlGnLCeM9fSWrHqpPdyCT0egeC21eA98TxCvs5aWoND9pIcfUh0zSH6J7CT+QxEcjBKIb2dHhARdE75lrfDyM5QkVg6kEvQ/M9LEJExXC5KCXHVpKlsEQwL/qU4YKSTlIzkU2RpJrbMEtNjWFottBxr5WzkV98/CxkxjysEXtFW7xF7kyVsGOjmkKzHQ=="
	}
	`

	var errs *[]error

	ExpectedSize := 71501

	n := snsNotice{}
	utils.Decoder([]byte(jsonStr), &n)

	errs = utils.Validate(&n)

	if len(*errs) != 0 {
		t.Error("Could not parse JSON message")
		return
	}

	msg := snsMessage{}
	utils.Decoder([]byte(n.Message), &msg)

	errs = utils.Validate(&msg)

	if len(*errs) != 0 {
		t.Error("Could not parse Records")
		return
	}

	record := msg.Records[0]

	if record.S3.Object.Size != ExpectedSize {
		t.Error("Object Size should have been", ExpectedSize, "Found", record.S3.Object.Size)
	}
}
