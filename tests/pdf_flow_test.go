package tests

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/server"
	"github.com/bulletind/moire/signature"

	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

var assetRet = map[string]string{}

func makeSignatureURL(path string) string {
	public_key := "HelloWorldTest"
	private_key := signature.GetSecretKey(public_key)
	timestamp := time.Now().Format(time.RFC3339)

	sign := signature.MakeSignature(public_key, private_key, timestamp, path)

	values := url.Values{
		"signature":  {sign},
		"timestamp":  {timestamp},
		"public_key": {public_key},
	}

	sorted := values.Encode()
	escaped := strings.Replace(sorted, "+", "%20", -1)
	return path + "?" + escaped
}

func TestCreatePDFURLName(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	f1 := randSeq(5)
	f2 := randSeq(10)
	f3 := randSeq(3)

	name := fmt.Sprintf("%v--%v//-%v.pdf", f1, f2, f3)

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "application/pdf",
		"name":      name,
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			exception := "Asset creation should return status 200."
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
			return
		}

		upload_url, err := url.Parse(assetRet["upload_url"])
		if err != nil {
			log.Error("Test Failed", "reason", err, "msg", msg)
			t.Error(err.Error())
			return
		}

		if !strings.HasSuffix(upload_url.Path, ".pdf") {
			exception := "URL must end with pdf. Found: " + upload_url.Path
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
			return
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
		"name":      randSeq(10),
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			exception := "Asset creation should return status 200."
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				reason := " should be a valid string in creation return."
				log.Error("Test Failed", "reason", reason, "key", key)
				t.Error(key + reason)
			}
		}
	})
}

func TestGetPDFNeedSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = "/assets/" + assetRet["_id"]
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

func TestGetPDF(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/" + assetRet["_id"])
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

func TestPDFThumbnailNeedsSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + assetRet["_id"] + "/thumbnail/?no_redirect=true"
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

func TestPDFThumbnailNoRedirect(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/"+assetRet["_id"]+"/thumbnail") + "&no_redirect=true"
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

func TestPDFThumbnailGetDefault(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = makeSignatureURL("/assets/" + assetRet["_id"] + "/thumbnail")
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "Message should be in pending state."

		if msg.Status != 301 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "attachment.png") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
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
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "not meant to be monitored") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestSNSMessageNotFound(t *testing.T) {
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
	}`, "original_file/134444")

	req := tests.MockRequest{}
	req.Url = "/notify/sns/"
	req.Method = "post"
	req.Data = map[string]interface{}{}
	utils.Decoder([]byte(snsString), &req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		exception := "This asset should not be found."
		if msg.Status != 500 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "not found") {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestSNSMessage(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	parsed_url, err := url.Parse(assetRet["upload_url"])
	if err != nil {
		log.Error("Test Failed", "reason", err)
		t.Error(err.Error())
		return
	}

	upload_path := parsed_url.Path

	log.Debug("Submitting SNS", "upload_path", upload_path)

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
		exception := "This asset should be marked as ready."
		if msg.Status != 200 {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}

		if msg.Message != "" {
			log.Error("Test Failed", "reason", exception, "msg", msg)
			t.Error(exception)
		}
	})
}

func TestGetPDFWithSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = makeSignatureURL("/assets/" + assetRet["_id"])
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		log.Debug("GetPDF", "msg", msg, "error", msg.Error)
	})
}
