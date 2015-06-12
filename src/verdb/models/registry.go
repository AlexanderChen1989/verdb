package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// Hourly 一个小时包含的秒数
	Hourly = 60 * 60
	// Daily 一天包含的秒数
	Daily = 24 * Hourly
	// Weekly 一个星期包含的秒数
	Weekly = 7 * Daily
	// Monthly 一个月包含的秒数
	Monthly = 30 * Daily
)

/*
Registry 注册模型
{
	"databaseName" : "frradar", // 目标数据库
	"collectionName" : "serverInfo", // 目标集合
	"name" : "frradar/serverInfo", // 命名方式： databaseName/collectionName
	"compareKey" : "serverId", // 用来标识同一个实体
	'verInterval': 24 * 60 * 60, // 版本粒度，示例为一天记录一个版本
	"indexKeys" : [ // 需要添加index的键
		"xxx.xxx",
		"xxx.xxxxx",
		...
	],
	"verKeys" : [ // 需要版本化记录的建
		"xxx.xxx",
		"xx.xxx.xx",
		...
	]
}
*/
type Registry struct {
	sync.Mutex     `json:"-" bson:"-"`
	DatabaseName   string   `json:"databaseName" bson:"databaseName" binding:"required"`
	CollectionName string   `json:"collectionName" bson:"collectionName" binding:"required"`
	Name           string   `json:"-" bson:"name"`
	CompareKey     string   `json:"compareKey" bson:"compareKey" binding:"required"`
	VerInterval    int64    `json:"verInterval" bson:"verInterval" binding:"required"`
	IndexKeys      []string `json:"indexKeys" bson:"indexKeys"`
	VerKeys        []string `json:"verKeys" bson:"verKeys"`
}

// GenVer 基于VerInterval生成版本号: unix seconds / interval
func (reg *Registry) GenVer() int64 {
	if reg.VerInterval <= 0 { // 用于测试时生成新版本
		return time.Now().UnixNano()
	}
	return int64(time.Now().Second()) / reg.VerInterval
}

// GenName 基于DatabaseName, CollectionName生成name
func (reg *Registry) GenName() string {
	return fmt.Sprintf("%s/%s", reg.DatabaseName, reg.CollectionName)
}

/*
Versionize 版本化记录数据
1. 对于提交的记录，添加_ver,_next生成新版本new
2. new的格式是{CompareKey: <int|string>, _ver: <int>, _next: <int>} 且 new._ver 相同于 new._next
3. 在数据库中查询相同CompareKey，且_next最大的版本old
4. 如果 old._ver 相同于 new._ver
	* 用new的内容跟新old，然后返回
5. 如果 old 相同于 new
	* 更新old版本的_next为new的_ver，然后返回
6 如果 old 不同于 new
	* 更新old版本的_next为new的_ver
	* 插入new，然后返回
*/
func (reg *Registry) Versionize(newDoc map[string]interface{}, sess *mgo.Session) error {
	reg.Lock()
	defer reg.Unlock()

	// 新建记录添加版本信息
	ver := reg.GenVer()
	newDoc["_ver"] = ver
	newDoc["_next"] = ver
	newDoc["_is_latest"] = true

	// 记录存储表
	collection := sess.DB(reg.DatabaseName).C(reg.CollectionName)

	// 查询表中同一实例最新的记录
	var oldDoc map[string]interface{}
	err := collection.Find(bson.M{
		reg.CompareKey: newDoc[reg.CompareKey],
		"_is_latest":   true,
	}).One(&oldDoc)

	// 如果没有找到记录，表面新记录是第一个版本
	if err == mgo.ErrNotFound {
		return collection.Insert(newDoc)
	} else if err != nil {
		return err
	}

	// 如果提交的记录和数据库中最新记录在同一个Interval中，用提交记录的信息更新数据库中的最新记录
	if oldDoc["_ver"] == newDoc["_ver"] {
		return collection.UpdateId(oldDoc["_id"], newDoc)
	}

	// 如果提交的记录和数据库中的记录内容一致，更新数据库中记录_next，同时用提交数据的内容更新数据库记录
	if !changed(oldDoc, newDoc, reg.VerKeys) {
		setMap := bson.M{"_next": ver}
		for k, v := range newDoc {
			// 屏蔽键： "_ver", "_next", "_id", "_is_latest"
			if len(k) > 0 && k[0] == '_' {
				continue
			}
			setMap[k] = v
		}
		return collection.UpdateId(
			oldDoc["_id"],
			bson.M{"$set": setMap},
		)
	}

	if err = collection.UpdateId(
		oldDoc["_id"],
		bson.M{"$set": bson.M{
			"_next":      ver - 1,
			"_is_latest": false,
		}},
	); err != nil {
		return err
	}
	return collection.Insert(newDoc)
}

// 比对两条记录，keys对应的值有没有改变。
// 比对两条记录时，如果键对应的值是列表，要求列表内容有稳定性。
// 也就是俩个高列表里面如果内容一致，但是顺序不一致会认为是不一样的列表。
func changed(ldoc, rdoc map[string]interface{}, keys []string) bool {
	for _, key := range keys {
		parts := strings.Split(key, ".")
		va := collectVals(ldoc, parts)
		vb := collectVals(rdoc, parts)
		if !reflect.DeepEqual(va, vb) {
			return true
		}
	}
	return false
}

// 从记录中收集所有keys对应的值
func collectVals(m map[string]interface{}, parts []string) []interface{} {
	var nexts = []interface{}{m}
	var vals []interface{}
	for i, k := range parts {
		var nnexts []interface{}
		for _, next := range nexts {
			switch tnext := next.(type) {
			case map[string]interface{}:
				if tnext[k] == nil {
					continue
				}
				if i == len(parts)-1 {
					vals = append(vals, tnext[k])
				} else {
					nnexts = append(nnexts, tnext[k])
				}

			case []interface{}:
				for _, v := range tnext {
					if tm, ok := v.(map[string]interface{}); ok {
						if tm[k] == nil {
							continue
						}
						if i == len(parts)-1 {
							vals = append(vals, tm[k])
						} else {
							nnexts = append(nnexts, tm[k])
						}
					}
				}
			}
		}
		if len(nnexts) == 0 {
			break
		}
		nexts = nnexts
	}
	return vals
}
