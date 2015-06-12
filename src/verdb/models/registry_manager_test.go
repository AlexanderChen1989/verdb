package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/mgo.v2"
)

func TestRegManger(t *testing.T) {
	// 注册信息存取的表
	const (
		database   = "metaInfo"
		collection = "regitries"
	)
	// 初始化数据库
	sess, _ := mgo.Dial("localhost")
	defer sess.Close()

	// 测试RegManger 能否保证name唯一性
	sess.DB(database).C(collection).DropCollection()
	rm := NewRegManger(database, collection, sess)
	regJSON := `
	{
	    "databaseName": "testdb",
	    "compareKey": "pk",
	    "collectionName": "testver",
	    "indexKeys": [
	        "a",
	        "b.c",
	        "d.e",
	        "d.i",
	        "f.g.h"
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

	for i := 0; i < 10; i++ {
		rm.CreateRegistry(reg, sess)
	}

	count, _ := sess.DB(database).C(collection).Count()
	if count != 1 {
		t.Errorf("注册信息名称必须唯一\n")
		return
	}

	// 测试RegManger 能否正确的存储注册信息
	sess.DB(database).C(collection).DropCollection()
	rm = NewRegManger(database, collection, sess)

	const num = 100
	for i := 0; i < num; i++ {
		reg.CollectionName = fmt.Sprintf("CollectionName%v", i)
		rm.CreateRegistry(reg, sess)
	}

	var regs []Registry
	sess.DB(database).C(collection).Find(nil).All(&regs)
	if len(regs) != rm.Size() || len(regs) != num {
		t.Errorf("数据库注册信息数目和缓存数目不一致 %v != %v != %v\n", len(regs), rm.Size(), num)
		return
	}

	// 测试为注册的表添加index
	contentEq := func(sa, sb []string) bool {
		if len(sa) != len(sb) {
			return false
		}
		for _, aitem := range sa {
			has := false
			for _, bitem := range sb {
				if aitem == bitem {
					has = true
				}
			}
			if !has {
				return false
			}
		}
		return true
	}
	for i := range regs {
		indexs, _ := sess.DB(regs[i].DatabaseName).C(regs[i].CollectionName).Indexes()
		var regIndexs []string
		for _, index := range indexs {
			regIndexs = append(regIndexs, index.Key...)
		}

		keys := append(reg.IndexKeys, "_id", "_ver", "_next", "_is_latest", reg.CompareKey)
		if !contentEq(regIndexs, keys) {
			t.Errorf("注册的表添加index错误\n")
			return
		}
	}
}
