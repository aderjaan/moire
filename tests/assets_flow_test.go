package tests

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/server"
	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

var assetRet = map[string]string{}

func TestCreateAsset(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "image/png",
		"name":      randSeq(10),
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				t.Error(key + " should be a valid string in creation return.")
			}
		}
	})
}

func TestGetAsset(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + assetRet["_id"] + "/?no_redirect=true"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		if msg.Status != 404 {
			t.Error("Message should be in pending state.")
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			t.Error("Message should be in pending state.")
		}
	})
}

func TestGetAssetThumbnail(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + assetRet["_id"] + "/thumbnail/?no_redirect=true"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "Message should be in pending state."

		if msg.Status != 404 {
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			t.Error(exception)
		}
	})
}

func TestCreatePDF(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "application/pdf",
		"name":      randSeq(10) + ".pdf",
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				t.Error(key + " should be a valid string in creation return.")
			}
		}
	})
}

func TestSNSMessageIgnoredPath(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	snsString := fmt.Sprintf(`{
		"Type" : "Notification",
		"MessageId" : "12",
		"TopicArn" : "arn",
		"Subject" : "Amazon S3 Notification",
		"Message" : "{\"Records\":[{\"s3\":{\"bucket\":{\"name\":\"sc-gallery\"},\"object\":{\"key\":\"%v\",\"size\":71501}}}]}",
		"Timestamp" : "2015-04-14T03:48:23.584Z",
		"SignatureVersion" : "1",
		"Signature" : "liP1M"
	}`, "/hello/world/")

	req := tests.MockRequest{}
	req.Url = "/notify/sns/"
	req.Method = "post"
	req.Data = map[string]interface{}{}
	utils.Decoder([]byte(snsString), &req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "This path should have been ignroed."

		if msg.Status != 400 {
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "not meant to be monitored") {
			t.Error(exception)
		}
	})
}

func TestSNSMessageNotFound(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	//parsed_url, err := url.Parse(assetRet["upload_url"])
	//if err != nil {
	//	t.Error(err.Error())
	//	return
	//}

	//upload_path := parsed_url.Path

	snsString := fmt.Sprintf(`{
		"Type" : "Notification",
		"MessageId" : "12",
		"TopicArn" : "arn",
		"Subject" : "Amazon S3 Notification",
		"Message" : "{\"Records\":[{\"s3\":{\"bucket\":{\"name\":\"sc-gallery\"},\"object\":{\"key\":\"%v\",\"size\":71501}}}]}",
		"Timestamp" : "2015-04-14T03:48:23.584Z",
		"SignatureVersion" : "1",
		"Signature" : "liP1M"
	}`, "original_file/134444")
	//strings.TrimPrefix(upload_path, "/"))

	req := tests.MockRequest{}
	req.Url = "/notify/sns/"
	req.Method = "post"
	req.Data = map[string]interface{}{}
	utils.Decoder([]byte(snsString), &req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "This asset should not be found."
		if msg.Status != 500 {
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "not found") {
			t.Error(exception)
		}
	})
}

func TestSNSMessage(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	parsed_url, err := url.Parse(assetRet["upload_url"])
	if err != nil {
		t.Error(err.Error())
		return
	}

	upload_path := parsed_url.Path

	fmt.Println("Submitting messages for", upload_path)

	snsString := fmt.Sprintf(`{
		"Type" : "Notification",
		"MessageId" : "12",
		"TopicArn" : "arn",
		"Subject" : "Amazon S3 Notification",
		"Message" : "{\"Records\":[{\"s3\":{\"bucket\":{\"name\":\"%v\"},\"object\":{\"key\":\"%v\",\"size\":71501}}}]}",
		"Timestamp" : "2015-04-14T03:48:23.584Z",
		"SignatureVersion" : "1",
		"Signature" : "liP1M"
	}`, config.Settings.S3.Bucket, strings.TrimPrefix(upload_path, "/"))

	req := tests.MockRequest{}
	req.Url = "/notify/sns/"
	req.Method = "post"
	req.Data = map[string]interface{}{}
	utils.Decoder([]byte(snsString), &req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		fmt.Println(msg)

		exception := "This asset should be marked as ready."
		if msg.Status != 200 {
			t.Error(exception)
		}

		if msg.Message != "" {
			t.Error(exception)
		}
	})
}

func TestGetAssetAfterSNS(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + assetRet["_id"] + "/?no_redirect=true"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		fmt.Println(msg)
	})
}
