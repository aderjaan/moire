package tests

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/server"

	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

var imageRet = map[string]string{}

func TestCreateImage(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "image/jpeg",
		"name":      randSeq(10),
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &imageRet)

		if msg.Status != 200 {
			reason := "Asset creation should return status 200."
			log.Error("Test Failed", "reason", reason, "msg", msg)
			t.Error(reason)
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := imageRet[key]; !ok || len(val) == 0 {
				reason := " should be a valid string in creation return."
				log.Error("Test Failed", "reason", reason, "key", key)
				t.Error(key + reason)
			}
		}
	})
}

func TestGetImageNeedSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = "/assets/" + imageRet["_id"]
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "This asset needs signature."
		if msg.Status != 412 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestGetImageTimeout(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/" + imageRet["_id"])
	req.Method = "get"

	time1 := time.Now()

	server.Test(&req, func(msg *tests.MockResponse) {
		elapsed := time.Since(time1)
		if elapsed < time.Duration(config.Settings.Moire.ImageTimeout)*time.Second {
			reason := "Should have taken more than 5 seconds to return"
			log.Error("Test Failed", "reason", reason, "msg", msg)
			t.Error(reason)
		}
	})
}

func TestGetImage(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/" + imageRet["_id"])
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "Message should be in pending state."

		if msg.Status != 301 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "thumbnail.png") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestImageThumbnailNeedsSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + imageRet["_id"] + "/thumbnail/?no_redirect=true"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "This asset needs signature."
		if msg.Status != 412 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestImageThumbnailNoRedirect(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/"+imageRet["_id"]+"/thumbnail") + "&no_redirect=true"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "Message should be in pending state."

		if msg.Status != 404 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestImageThumbnailGetDefault(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/" + imageRet["_id"] + "/thumbnail")
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "Message should be in pending state."

		if msg.Status != 301 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "thumbnail.png") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestImageSNSMessage(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	parsed_url, err := url.Parse(imageRet["upload_url"])
	if err != nil {
		t.Error(err.Error())
		return
	}

	upload_path := parsed_url.Path

	log.Debug("Submitting SNS Message", "path", upload_path)

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
		if msg.Status != 500 {
			exception := "This asset should be marked as ready."
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "ffmpeg") {
			exception := "ffmpeg should not have been found."
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestGetImageWithSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = makeSignatureURL("/assets/" + imageRet["_id"])
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		log.Debug("GetImage", "msg", msg, "error", msg.Error)
	})
}

func TestGetThumbnailWithSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = makeSignatureURL("/assets/" + imageRet["_id"] + "/thumbnail")
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		log.Debug("GetThumbnail", "msg", msg, "error", msg.Error)
	})
}
