package tests

import (
	"testing"

	"fmt"
	"github.com/bulletind/moire/server"
	"github.com/kr/pretty"
	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

var assetRet = map[string]string{}

func TestCreatePDFCollection(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type":  "application/pdf",
		"name":       "testAsset",
		"collection": "hello",
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		fmt.Println(msg.Data)

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
