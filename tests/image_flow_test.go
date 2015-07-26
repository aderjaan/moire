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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := imageRet[key]; !ok || len(val) == 0 {
				fmt.Printf("%# v", pretty.Formatter(msg))
				t.Error(key + " should be a valid string in creation return.")
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("Should have taken more than 5 seconds to return")
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "thumbnail.png") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "required parameter") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "content is still being uploaded") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "thumbnail.png") {
			fmt.Printf("%# v", pretty.Formatter(msg))
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
		if msg.Status != 500 {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error(exception)
		}

		if !strings.Contains(msg.Message, "mogrify") {
			fmt.Printf("%# v", pretty.Formatter(msg))
			t.Error("mogrify: unrecognized option")
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
		fmt.Printf("%# v", pretty.Formatter(msg))
		fmt.Println(msg.Error)
	})
}

func TestGetThumbnailWithSignature(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}

	req.Url = makeSignatureURL("/assets/" + imageRet["_id"] + "/thumbnail")
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		fmt.Printf("%# v", pretty.Formatter(msg))
		fmt.Println(msg.Error)
	})
}
