package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"verdb/models"

	"gopkg.in/mgo.v2"

	"github.com/gin-gonic/gin"
)

func TestNewRegistry(t *testing.T) {
	// init server
	r := gin.Default()
	sess, _ := mgo.Dial("localhost")
	sess.DB(MetaDB).C(RegCollection).DropCollection()

	server := NewServer(r, sess)

	// construct base data
	regJSON := `
	{
		"databaseName": "frradar", 
		"collectionName": "serverInfo", 
		"compareKey": "serverId", 
		'verInterval': 86400, 
		"indexKeys": [ 
			"a.b",
			"a.c.d"
		],
		"verKeys": [ 
			"a.b",
			"a.c.d",
			"a.e.f"
		]
	}
	`
	var reg models.Registry
	json.Unmarshal([]byte(regJSON), &reg)

	const num = 10
	for i := 0; i < num; i++ {
		nreg := reg
		nreg.CollectionName = fmt.Sprintf("%s[%d]", reg.CollectionName, i)

		buf, _ := json.Marshal(&nreg)
		response := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/registry", bytes.NewBuffer(buf))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(response, req)

		if response.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", response.Code, string(response.Body.Bytes()))
			return
		}
	}
	n, _ := sess.DB(MetaDB).C(RegCollection).Count()
	if n != num {
		t.Errorf("lost registries, expected: %d, got: %d\n", num, n)
		return
	}
}
