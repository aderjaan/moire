package tests

import (
	"testing"

	"github.com/bulletind/moire/server"
	"gopkg.in/simversity/gottp.v3/tests"
	"gopkg.in/simversity/gottp.v3/utils"
)

func TestCreatePDFCollection(t *testing.T) {
	server := server.MockDBServer()
	defer server.Close()

	req := tests.MockRequest{}
	req.Url = "/assets"
	req.Method = "post"
	req.Data = map[string]interface{}{
		"mime_type":  "application/pdf",
		"name":       randSeq(10),
		"collection": randSeq(5),
	}

	server.Test(&req, func(msg *tests.MockResponse) {
		utils.Convert(&msg.Data, &assetRet)

		if msg.Status != 200 {
			reason := "Asset creation should return status 200."
			log.Error("TestCreatePDFCollection failed", "reason", reason, "msg", msg)
			t.Error(reason)
		}

		for _, key := range []string{"upload_url", "url", "_id"} {
			if val, ok := assetRet[key]; !ok || len(val) == 0 {
				reason := " should be a valid string in creation return."
				log.Error("TestCreatePDFCollection failed", "reason", reason, "key", key)
				t.Error(key + reason)
			}
		}
	})
}
