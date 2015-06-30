package server

import (
	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
	"testing"
)

func TestUrls(t *testing.T) {
	server := MockServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/urls"

	server.Test(&req, func(msg *tests.MockResponse) {
		if msg.Status != 200 {
			t.Error("Expected Status return", msg.Status)
		}

		urlResponse := map[string]string{}
		utils.Convert(&msg.Data, &urlResponse)

		if _, ok := urlResponse["assets"]; !ok {
			t.Error("Invalid response from Server")
		}

		if _, ok := urlResponse["thumbnail"]; !ok {
			t.Error("Invalid response from Server")
		}
	})
}
