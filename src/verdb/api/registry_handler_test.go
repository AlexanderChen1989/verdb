package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"verdb/models"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/gin-gonic/gin"
)

func TestNewRegistry(t *testing.T) {
	// init server
	r := gin.Default()
	sess, err := mgo.Dial("localhost")
	if err != nil {
		t.Errorf("Error to connect to mongo %s\n", err)
	}
	sess.DB(MetaDB).C(RegCollection).DropCollection()

	server := NewServer(r, sess)

	// construct base data
	regJSON := `
	{
		"databaseName": "frradar", 
		"collectionName": "serverInfo", 
		"compareKey": "serverId", 
		"verInterval": 86400, 
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
	err = json.Unmarshal([]byte(regJSON), &reg)
	if err != nil {
		t.Errorf("Error %s", err)
		return
	}

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

		if i != 0 && i%2 == 0 {
			response := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/registry", bytes.NewBuffer(buf))
			req.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(response, req)
			if response.Code != http.StatusInternalServerError {
				t.Errorf("Allow create same registry %v\n", response.Code)
				return
			}
		}

	}

	type RegistryWithId struct {
		models.Registry `bson:",inline"`
		Id              bson.ObjectId `bson:"_id"`
	}
	var regs []RegistryWithId
	err = sess.DB(MetaDB).C(RegCollection).Find(nil).All(&regs)
	if err != nil || len(regs) != num {
		t.Errorf("lost registries, expected: %d, got: %d\n", num, len(regs))
		return
	}

	type ReturnRegistry struct {
		Msg models.Registry
	}

	for i := range regs {
		response := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/registry/"+regs[i].Id.Hex(), nil)
		server.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", response.Code, string(response.Body.Bytes()))
			return
		}
		var rreg ReturnRegistry
		json.NewDecoder(response.Body).Decode(&rreg)
		if !reflect.DeepEqual(rreg.Msg, regs[i].Registry) {
			t.Errorf("Delete not return right record\n%+v\n%+v\n", rreg.Msg, regs[i].Registry)
			return
		}
	}
}
