package tests

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"

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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Asset creation should return status 200.")
			return
		}

		upload_url, err := url.Parse(assetRet["upload_url"])
		if err != nil {
			t.Error(err.Error())
			return
		}

		fmt.Println(assetRet["upload_url"])
		if !strings.HasSuffix(upload_url.Path, ".pdf") {
			t.Error("URL must end with pdf. Found: " + upload_url.Path)
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				fmt.Printf("%# v", pretty.Formatter(msg))
				t.Error(key + " should be a valid string in creation return.")
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
		if msg.Status != 404 {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Message should be in pending state.")
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Message should be in pending state.")
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "attachment.png") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "not meant to be monitored") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
		exception := "This message should be ignored."
		if msg.Status != 200 {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if msg.Message != "" {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
		exception := "This asset should be marked as ready."
		if msg.Status != 200 {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if msg.Message != "" {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
		fmt.Printf("%# v", pretty.Formatter(msg))
		fmt.Println(msg.Error)
	})
}
