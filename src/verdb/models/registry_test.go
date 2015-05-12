package models

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2"
)

func TestVersionize(t *testing.T) {
	// 需要测试的新版本数
	num := 10000

	// 基础记录
	baseJSON := `
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
	var base map[string]interface{}
	json.Unmarshal([]byte(baseJSON), &base)
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
	// 基础记录修改函数
	var count = 100
	var getCount = func() int {
		count++
		return count
	}
	var changeA = func(doc map[string]interface{}) map[string]interface{} {
		doc["a"] = getCount()
		return doc
	}
	var changeBC = func(doc map[string]interface{}) map[string]interface{} {
		doc["b"].(map[string]interface{})["c"] = getCount()
		return doc
	}
	var changeDE = func(doc map[string]interface{}) map[string]interface{} {
		for _, item := range doc["d"].([]interface{}) {
			item.(map[string]interface{})["e"] = getCount()
		}
		return doc
	}
	var changeDI = func(doc map[string]interface{}) map[string]interface{} {
		for _, item := range doc["d"].([]interface{}) {
			item.(map[string]interface{})["i"] = getCount()
		}
		return doc
	}
	var changeFGH = func(doc map[string]interface{}) map[string]interface{} {
		for _, item := range doc["f"].([]interface{}) {
			for _, nitem := range item.(map[string]interface{})["g"].([]interface{}) {
				nitem.(map[string]interface{})["h"] = getCount()
			}
		}
		return doc
	}
	var changeJK = func(doc map[string]interface{}) map[string]interface{} {
		doc["j"].(map[string]interface{})["k"] = getCount()
		return doc
	}

	// 生成新版本修改函数列表
	changeLst := []func(map[string]interface{}) map[string]interface{}{changeA, changeBC, changeDE, changeDI, changeFGH}

	// 更新最新记录函数列表
	updateLst := []func(map[string]interface{}) map[string]interface{}{changeJK}

	// 从函数列表中随机选择一些函数
	randSelect := func(fns []func(map[string]interface{}) map[string]interface{}) []func(map[string]interface{}) map[string]interface{} {
		size := len(fns)
		var nfns []func(map[string]interface{}) map[string]interface{}
		for i := 0; i < 1+rand.Intn(size); i++ {
			nfns = append(nfns, fns[rand.Intn(size)])
		}
		return nfns
	}

	// registery
	regJSON := `
	{
	    "databaseName": "testdb",
	    "compareKey": "pk",
	    "collectionName": "testver",
	    "indexKeys": [
	        ["a"],
	        ["b.c"],
	        ["d.e"],
	        ["d.i"],
	        ["f.g.h"]
	    ],
	    "verKeys": [
			"a",
			"b.c",
			"d.e",
			"d.i",
			"f.g.h"
	    ]
    }`

	var reg = &Registry{}
	json.Unmarshal([]byte(regJSON), reg)

	// 链接数据库
	sess, _ := mgo.Dial("localhost")

	// 清空数据库
	sess.DB(reg.DatabaseName).C(reg.CollectionName).DropCollection()

	rmVerInfo := func(doc map[string]interface{}) {
		for k := range doc {
			if len(k) > 0 && k[0] == '_' {
				delete(doc, k)
			}
		}
	}

	findLast := func(sess *mgo.Session, reg *Registry) map[string]interface{} {
		var last map[string]interface{}
		sess.DB(reg.DatabaseName).C(reg.CollectionName).Find(
			map[string]interface{}{
				"_is_latest":   true,
				reg.CompareKey: base[reg.CompareKey],
			},
		).One(&last)
		return last
	}

	countNum := func(sess *mgo.Session, reg *Registry) int {
		num, _ := sess.DB(reg.DatabaseName).C(reg.CollectionName).Count()
		return num
	}

	// prettyJSON := func(data interface{}) string {
	// 	s, _ := json.MarshalIndent(data, "", "\t")
	// 	return string(s)
	// }

	// 测试插入第一个记录
	reg.Versionize(base, sess)
	if countNum(sess, reg) != 1 {
		t.Errorf("插入第一条记录错误\n")
		return
	}

	// 生成num个新版
	for i := 0; i < num; i++ {
		// 生成几个新版本，但是修改的key不需要生成新版本，只需要更新数据库中最新的记录
		for j := 0; j < 1+rand.Intn(6); j++ {
			for _, fn := range randSelect(updateLst) {
				base = fn(base)
			}
			base["other_key"] = i + j
			count := countNum(sess, reg)
			oldLast := findLast(sess, reg)
			reg.Versionize(base, sess)
			last := findLast(sess, reg)
			if !(oldLast["_next"].(int64) <= last["_next"].(int64)) {
				t.Errorf("非版本化键修改，插入版本时未更新_next\n")
				return
			}
			if oldLast["_ver"].(int64) != last["_ver"].(int64) {
				t.Errorf("非版本化键修改，更新了版本\n")
				return
			}
			rmVerInfo(base)
			rmVerInfo(last)
			if !reflect.DeepEqual(base, last) {
				t.Errorf("非版本化键修改，插入版本时更新错误\n %#v \n %#v\n", (base), (last))
				return
			}
			if count != countNum(sess, reg) {
				t.Errorf("非版本化键修改，版本数据不对\n")
				return
			}
		}

		// 添加新版本
		for _, fn := range randSelect(changeLst) {
			base = fn(base)
		}
		for _, fn := range randSelect(updateLst) {
			base = fn(base)
		}

		count := countNum(sess, reg)
		oldLast := findLast(sess, reg)
		reg.Versionize(base, sess)
		last := findLast(sess, reg)
		if !(oldLast["_next"].(int64) <= last["_next"].(int64)) {
			t.Errorf("版本化键修改，插入版本时未更新_next\n")
		}
		if oldLast["_ver"].(int64) == last["_ver"].(int64) {
			t.Errorf("版本化键修改，未生成版本\n")
		}
		if last["_ver"].(int64) != last["_next"].(int64) {
			t.Errorf("版本化键修改，_ver 不等于 _next\n")
		}
		rmVerInfo(last)
		rmVerInfo(base)
		if !reflect.DeepEqual(base, last) {
			t.Errorf("版本化键修改，插入新版本时更新错误\n")
		}
		if count+1 != countNum(sess, reg) {
			t.Errorf("版本化键修改，版本数目不对\n")
		}
	}
}
