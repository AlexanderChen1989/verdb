package verdb

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func testVerGenFn(base int, skip int) func() int {
	return func() int {
		base += skip
		return base
	}
}

func TestApi(t *testing.T) {
	const (
		db   = "testdb"
		repo = "testver"
	)

	api := NewApiServer("localhost", db, 10, testVerGenFn(100, 20))
	defer api.Close()

	sess := api.Clone()
	defer sess.Close()

	// do some clean up
	sess.DB(db).C(repo).DropCollection()

	// test register
	reg := `
	{
	    "db_name": "testdb",
	    "compare_key": "tid",
	    "repo": "testver",
	    "index_keys": [
	        ["a"],
	        ["b"],
	        ["e"],
	        ["f"]
	    ],
	    "ver_keys": [
	        "a", "b"
	    ],
	    "update_keys": [
			"e", "f"
	    ]
    }`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reg))
	api.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
	}

	// test versionize
	vers := []string{
		`[{"tid": 1, "a": 1, "b": 2, "e": 1, "f": 2}]`,
		`[{"tid": 1, "a": 1, "b": 1, "e": 1, "f": 2}]`,
		`[{"tid": 1, "a": 1, "b": 1, "e": 1, "f": 2}]`,
		`[{"tid": 1, "a": 1, "b": 1, "e": 1, "f": 2}]`,
		`[{"tid": 1, "a": 1, "b": 1, "e": 1, "f": 2}]`,
		`[{"tid": 1, "a": 2, "b": 1, "e": 1, "f": 2}]`,
		`[
			{"tid": 1, "a": 2, "b": 1, "e": 1, "f": 2},
			{"tid": 1, "a": 2, "b": 1, "e": 1, "f": 2},
			{"tid": 1, "a": 2, "b": 1, "e": 1, "f": 2},
			{"tid": 1, "a": 2, "b": 1, "e": 1, "f": 2}
		]`,
		`[{"tid": 1, "a": 2, "b": 1, "e": 4, "f": 3}]`,
	}
	for _, ver := range vers {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/versionize/testdb/testver", bytes.NewBufferString(ver))
		api.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
		}
	}

	var docs []bson.M
	sess.DB(db).C(repo).Find(nil).All(&docs)
	if len(docs) != 3 {
		t.Error("lost version!", len(docs), docs)
		return
	}
	for _, doc := range docs {
		m := map[string]float64{"e": 4.0, "f": 3.0}
		for k, v := range m {
			parts := strings.Split(k, ".")
			if !(getVal(doc, parts).(float64) == v) {
				t.Errorf("%v\t%v\n", getVal(doc, parts), v)
			}
		}
	}

	// test search
	// test cases
	var tcs = []struct {
		searchJson  []byte
		expectedLen int
		isValid     func(bson.M) bool
	}{
		{
			[]byte(`{"a": {"$gte": 1}, "b": 1}`),
			2,
			func(item bson.M) bool {
				return item["a"].(float64) >= 1.0 && item["b"].(float64) == 1.0
			},
		},
		{
			[]byte(`{"a": {"$gte": 2}, "b": 1}`),
			1,
			func(item bson.M) bool {
				return item["a"].(float64) >= 2.0 && item["b"].(float64) == 1.0
			},
		},
	}

	for _, tc := range tcs {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/search/testdb/testver", bytes.NewBuffer(tc.searchJson))
		api.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
			return
		}

		var result struct {
			Status  string   `json:"status"`
			Payload []bson.M `json:"payload"`
		}

		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result.Payload) != tc.expectedLen {
			t.Errorf("Failed: %v\n", result)
			return
		}
		for _, item := range result.Payload {
			if !tc.isValid(item) {
				t.Errorf("Failed: %v\n", result)
				return
			}
		}
	}
}

func sTestJobsApi(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	api := NewApiServer("localhost", testDb, 10, nil)
	defer api.Close()

	sess := api.Clone()
	defer sess.Close()

	// do some clean up
	sess.DB(testDb).DropDatabase()

	// test create job
	tcs := []struct {
		Json    string
		Compare func(interface{}, interface{}) bool
	}{
		{
			`{
			"name": "sjb-count",
			"type": "CountJob",
			"target_db": "testdb",
			"target_repo": "testrepo",
			"query": {"a": {"$gte": 100}}
			}`,
			func(info, result interface{}) bool {
				return info == nil && result.(float64) == 900.0
			},
		},
		{
			`{
				"name": "sjb-distinct",
				"type": "DistinctJob",
				"target_db": "testdb",
				"target_repo": "testrepo",
				"distinct_key": "a",
				"query": {"a": {"$gte": 100}}
			}`,
			func(info, result interface{}) bool {
				return info == nil && len(result.([]interface{})) == 900
			},
		},
		{
			`{
				"name": "mrj-inline",
				"type": "MapReduceJob",
				"target_db": "testdb",
				"target_repo": "testrepo",
				"query": {"a": {"$gte": 100}},
				"map_reduce" : {
			        "map" : "function() { emit(this.b, this.a) }",
			        "reduce" : "function(key, values) { return Array.sum(values) }"
			    }
			}`,
			func(info, result interface{}) bool {
				_, ok := result.([]interface{})
				return info != nil && ok
			},
		},
		{
			`{
				"name": "mrj-out",
				"type": "MapReduceJob",
				"target_db": "testdb",
				"target_repo": "testrepo",
				"query": {"a": {"$gte": 100}},
				"map_reduce" : {
			        "map" : "function() { emit(this.b, this.a) }",
			        "reduce" : "function(key, values) { return Array.sum(values) }",
			        "out": "results"
			    }
			}`,
			func(info, result interface{}) bool {
				return info != nil && result == nil
			},
		},
		{
			`{
				"name": "pj",
				"type": "PipelineJob",
				"target_db": "testdb",
				"target_repo": "testrepo",
				"pipeline" : [{"$match": { "a": {"$gte": 700}}}, {"$group": {"_id": "$b", "total": {"$sum": "$a"}}}]
			}`,
			func(info, result interface{}) bool {
				return info == nil && len(result.([]interface{})) == 15
			},
		},
	}

	var jobs = map[string]struct {
		Type    string
		Compare func(interface{}, interface{}) bool
	}{}
	for _, tc := range tcs {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/job", bytes.NewBufferString(tc.Json))
		api.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
			return
		}
		var ret struct {
			Job Job `json:"msg"`
		}
		err := json.NewDecoder(w.Body).Decode(&ret)
		if err != nil {
			t.Errorf("Failed: %s", err)
			return
		}
		if !ret.Job.Id.Valid() {
			t.Errorf("Failed: wrong return Job.Id.\n%+v\n", ret.Job)
			return
		}
		jobs[ret.Job.Id.Hex()] = struct {
			Type    string
			Compare func(interface{}, interface{}) bool
		}{
			ret.Job.Type,
			tc.Compare,
		}
	}

	if num, err := sess.DB(testDb).C("jobs").Count(); err != nil || num != len(tcs) {
		t.Errorf("Failed: wrong jobs length.")
	}

	// clean
	testRepo := sess.DB("testdb").C("testrepo")
	testRepo.DropCollection()

	// insert test data
	for i := 0; i < 1000; i++ {
		testRepo.Insert(bson.M{"a": i, "b": i / 20})
	}
	for id, comp := range jobs {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/api/job/"+id+"/sched", nil)
		if err != nil {
			t.Errorf("Failed to create request: %s\n", err)
			return
		}
		api.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Failed: %v\n%s\n", id, string(w.Body.Bytes()))
			return
		}
		payload := w.Body.String()
		var result struct {
			Info   interface{}
			Result interface{}
		}

		err = json.NewDecoder(bytes.NewBufferString(payload)).Decode(&result)
		if err != nil {
			t.Errorf("Failed: %s", err)
			return
		}
		if !comp.Compare(result.Info, result.Result) {
			t.Errorf("Failed: %+v\n%+v\n", result.Info, result.Result)
			return
		}
	}
}

func sTestUpsertApi(t *testing.T) {
	api := NewApiServer("localhost", testDb, 10, nil)
	defer api.Close()

	sess := api.Clone()
	defer sess.Close()

	// do some clean up
	sess.DB("testdb").DropDatabase()

	// test register
	reg := `
	{
	    "db_name": "testdb",
	    "compare_key": "tid",
	    "repo": "testver",
	    "index_keys": [
	        ["a"],
	        ["b"]
	    ],
	    "ver_keys": [
	        "a", "b"
	    ]
    }`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reg))
	api.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
	}

	// test versionize

	vers := []string{
		`{"insert": {"tid": 1, "a": 1, "b": 2}}`,
		`{"insert": {"tid": 1, "a": 1, "b": 1}}`,
		`{"insert": {"tid": 1, "a": 1, "b": 1}}`,
		`{"insert": {"tid": 1, "a": 1, "b": 1}}`,
		`{"insert": {"tid": 1, "a": 1, "b": 1}}`,
		`{"insert": {"tid": 1, "a": 2, "b": 1}}`,
		`{"insert": {"tid": 1, "a": 2, "b": 1}}`,
		`{
			"insert": {"tid": 1, "a": 200, "b": 300},
			"update": {"$set": {"a": 200, "b": 300}}
		}`,
	}
	for _, ver := range vers {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/upsert/testdb/testver", bytes.NewBufferString(ver))
		api.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes()))
		}
	}

	var docs []bson.M
	sess.DB("testdb").C("testver").Find(nil).All(&docs)
	if len(docs) != 1 {
		t.Error("Failed: wrong number of docs.", len(docs), docs)
		return
	}
	if docs[0]["tid"].(float64) != 1.0 || docs[0]["a"].(float64) != 200.0 || docs[0]["b"].(float64) != 300.0 {
		t.Errorf("Failed to update: %v\n", docs)
	}
}
