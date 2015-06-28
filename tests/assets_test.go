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
