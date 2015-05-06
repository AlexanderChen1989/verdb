package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"verdb"

	"gopkg.in/mgo.v2/bson"
)

func panicerr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var num int
	flag.IntVar(&num, "n", 100, "Number of versionz.")
	flag.Parse()
	// create base
	s := `
	{
		"pk": 1,
		"a": 1,
		"b": {
			"c": 1
		},
		"d": [
			{"e": 1,  "i": 2},
			{"e": 2, "i": 3}
		],
		"f": [
			{"g": [{"h": 1}, {"h": 2}]}
		],
		"j": {
			"k": 1
		}
	}`
	var base bson.M
	json.Unmarshal([]byte(s), &base)
	/*
		change:
			versionized:
				a
				b.c
				d.e
				d.i
				f.g.h
		updated:
			j.k

	*/
	// base change funcs
	var count = 100
	var getCount = func() int {
		count++
		return count
	}
	var changeA = func(doc bson.M) bson.M {
		doc["a"] = getCount()
		return doc
	}
	var changeBC = func(doc bson.M) bson.M {
		doc["b"].(map[string]interface{})["c"] = getCount()
		return doc
	}
	var changeDE = func(doc bson.M) bson.M {
		for _, item := range doc["d"].([]interface{}) {
			item.(map[string]interface{})["e"] = getCount()
		}
		return doc
	}
	var changeDI = func(doc bson.M) bson.M {
		for _, item := range doc["d"].([]interface{}) {
			item.(map[string]interface{})["i"] = getCount()
		}
		return doc
	}
	var changeFGH = func(doc bson.M) bson.M {
		for _, item := range doc["f"].([]interface{}) {
			for _, nitem := range item.(map[string]interface{})["g"].([]interface{}) {
				nitem.(map[string]interface{})["h"] = getCount()
			}
		}
		return doc
	}
	var changeJK = func(doc bson.M) bson.M {
		doc["j"].(map[string]interface{})["k"] = getCount()
		return doc
	}

	changeLst := []func(bson.M) bson.M{changeA, changeBC, changeDE, changeDI, changeFGH}
	updateLst := []func(bson.M) bson.M{changeJK}
	randSelect := func(fns []func(bson.M) bson.M) []func(bson.M) bson.M {
		size := len(fns)
		var nfns []func(bson.M) bson.M
		for i := 0; i < 1+rand.Intn(size); i++ {
			nfns = append(nfns, fns[rand.Intn(size)])
		}
		return nfns
	}

	//
	const (
		db   = "testdb"
		repo = "testver"
	)
	var testVerGenFn = func(base int, skip int) func() int {
		return func() int {
			base += skip
			return base
		}
	}

	api := verdb.NewApiServer("localhost", db, 10, testVerGenFn(100, 2))
	defer api.Close()

	sess := api.Clone()
	defer sess.Close()

	// do some clean up
	sess.DB(db).C(repo).DropCollection()

	// registery
	reg := `
	{
	    "db_name": "testdb",
	    "compare_key": "pk",
	    "repo": "testver",
	    "index_keys": [
	        ["a"],
	        ["b.c"],
	        ["d.e"],
	        ["d.i"],
	        ["f.g.h"]
	    ],
	    "ver_keys": [
			"a",
			"b.c",
			"d.e",
			"d.i",
			"f.g.h"
	    ],
	    "update_keys": [
			"j.k"
	    ]
    }`

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reg))
	panicerr(err)
	api.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		panic(fmt.Sprintf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes())))

	}

	var nrep int
	var sendBase = func(doc bson.M) {
		w := httptest.NewRecorder()
		buf := bytes.NewBuffer(nil)
		err := json.NewEncoder(buf).Encode([]interface{}{doc})
		if err != nil {
			panic(fmt.Sprintf("[Fail] %s", err))
		}
		req, err := http.NewRequest("POST", "/api/versionize/testdb/testver", buf)
		panicerr(err)
		api.ServeHTTP(w, req)
		nrep++
		if w.Code != http.StatusOK {
			panic(fmt.Sprintf("Status Code: %v\n%s\n", w.Code, string(w.Body.Bytes())))
		}
	}

	// generate num versions
	ta := time.Now()
	for i := 0; i < num; i++ {
		// send 6 normal versions
		for j := 0; j < 1+rand.Intn(6); j++ {
			for _, fn := range randSelect(updateLst) {
				base = fn(base)
			}
			base["other_key"] = i + j
			sendBase(base)
		}

		// send a new version
		for _, fn := range randSelect(changeLst) {
			base = fn(base)
		}
		for _, fn := range randSelect(updateLst) {
			base = fn(base)
		}
		sendBase(base)
	}
	fmt.Println("Update", base["j"].(map[string]interface{})["k"])
	fmt.Println("Time since", time.Since(ta), float64(nrep)/time.Since(ta).Seconds(), "reqs/s")
}
