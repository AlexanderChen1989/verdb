package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestSearchInfo(t *testing.T) {
	const (
		testdb         = "testdb"
		testcollection = "testcollection"
	)
	sess, err := mgo.Dial("localhost")
	if err != nil {
		t.Errorf("无法连接mongodb %s", err.Error())
		return
	}
	// 初始化数据库
	coll := sess.DB(testdb).C(testcollection)
	coll.DropCollection()

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		coll.Insert(bson.M{"a": i, "b": bson.M{"c": i / 10}})
	}

	// 初始化server
	server := NewServer(gin.Default(), sess)

	// 构建查询条件
	queries := []string{
		`{"b": {"$gt": 130}}`,
		`{"b": {"$gt": 200}}`,
		`{"b": {"$gt": 200}, "c": {"$lte": 60}}`,
	}
	for _, query := range queries {
		var queryM bson.M
		if err := json.Unmarshal([]byte(query), &queryM); err != nil {
			t.Errorf("%s\n", err.Error())
			return
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/api/search/%s/%s", testdb, testcollection), bytes.NewBufferString(query))
		if err != nil {
			t.Errorf("无法创建Request\n")
			return
		}
		req.Header.Add("Content-Type", "application/json")
		server.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			data, _ := ioutil.ReadAll(res.Body)

			t.Errorf("无法查询%s\n", string(data))
			return
		}
		var result struct {
			Result []bson.M
		}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			t.Errorf("返回结果错误\n")
			return
		}

		var dbResult []bson.M
		if err := coll.Find(queryM).All(&dbResult); err != nil {
			t.Errorf("无法从数据库中查询\n")
			return
		}

		if !reflect.DeepEqual(result.Result, dbResult) {
			t.Errorf("发回结果和查询结果不一致\n")
			return
		}
	}

}
