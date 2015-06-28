package tests

import (
	"log"
	"strings"
	"testing"

	"gopkg.in/bulletind/moire/server"
	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

func TestCreateAssetFail(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "image/png",
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		if msg.Status != 400 {
			t.Error("Asset creation should fail without mime_type and name")
		}
	})
}

func TestCreateAssetFailAgain(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type": "image/png",
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		if msg.Status != 400 {
			t.Error("Asset creation should fail without name")
		}
	})
}

var assetId = ""

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

	log.Println("Creating asset:", req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		assetRet := map[string]string{}
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				t.Error(key + " should be a valid string in creation return.")
			} else if key == "_id" {
				assetId = val
			}
		}
	})
}

func TestAssetShouldRaiseException(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/hello_world/"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		if msg.Status != 500 {
			t.Error("Message should raise Exception")
		}
	})
}

func TestGetAsset(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/" + assetId + "/?no_redirect=true"
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

	log.Println("Creating asset:", req.Data)

	server.Test(&req, func(msg *tests.MockResponse) {
		assetRet := map[string]string{}
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			t.Error("Asset creation should return status 200.")
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				t.Error(key + " should be a valid string in creation return.")
			} else if key == "_id" {
				assetId = val
			}
		}
	})
}
