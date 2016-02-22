package tests

import (
	"testing"

	"github.com/bulletind/moire/server"
	"github.com/bulletind/moire/config"
	"gopkg.in/simversity/gottp.v3/tests"
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
	config.Settings.Moire.SignRequests = false
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets/hello_world/"
	req.Method = "get"

	server.Test(&req, func(msg *tests.MockResponse) {
		config.Settings.Moire.SignRequests = true
		if msg.Status != 500 {
			t.Errorf("Message should return 500 (Exception), but returned %v", msg.Status)
		}
	})
}
