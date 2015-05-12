package models

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// RegManager 注册信息管理者
type RegManager struct {
	sync.RWMutex

	database   string // 存储注册数据的库
	collection string // 存储注册数据的表

	registries map[string]*Registry // 内存注册缓存
}

// NewRegManger 返回新生成的RegManager
func NewRegManger(database, collection string, sess *mgo.Session) *RegManager {
	coll := sess.DB(database).C(collection)

	// 确保注册数据表添加index
	index := mgo.Index{
		Key:        []string{"name"},
		Unique:     true, // 祖册名称需要是唯一的
		DropDups:   false,
		Background: false,
		Sparse:     true,
	}
	err := coll.EnsureIndex(index)
	if err != nil {
		log.Println(err)
		return nil
	}

	// 数据库中读取注册信息
	var regs []Registry
	err = coll.Find(nil).All(&regs)
	if err != nil {
		log.Println(err)
		return nil
	}

	// 创建 name -> Registry 映射
	rm := &RegManager{
		database:   database,
		collection: collection,
		registries: map[string]*Registry{},
	}

	for i := range regs {
		rm.registries[regs[i].Name] = &regs[i]
	}

	return rm
}

// Size 返回注册数目
func (rm *RegManager) Size() int {
	rm.RLock()
	defer rm.RUnlock()

	return len(rm.registries)
}

// Register 注册一条注册信息
func (rm *RegManager) Register(reg *Registry, sess *mgo.Session) error {
	rm.Lock()
	defer rm.Unlock()

	if reg.DatabaseName == "" || reg.CollectionName == "" || reg.CompareKey == "" {
		return errors.New("db_name, db_name, compare_key cant be empty")
	}
	if len(reg.VerKeys) == 0 {
		return errors.New("ver_keys cant be empty")
	}

	reg.Name = fmt.Sprintf("%s/%s", reg.DatabaseName, reg.CollectionName)

	// 保存注册信息到数据库
	_, err := sess.DB(rm.database).C(rm.collection).Upsert(bson.M{"name": reg.Name}, reg)
	if err != nil {
		return err
	}
	// 缓存注册信息到缓存
	rm.registries[reg.Name] = reg

	// 添加index
	regRepo := sess.DB(reg.DatabaseName).C(reg.CollectionName)
	for _, index := range append(reg.IndexKeys,
		reg.CompareKey,
		"_ver",
		"_next",
		"_is_latest",
	) {
		if err = regRepo.EnsureIndexKey(index); err != nil {
			return err
		}
	}

	return nil
}

// GetReg 查询注册信息
func (rm *RegManager) GetReg(database, collection string) *Registry {
	rm.RLock()
	defer rm.RUnlock()

	// return reg from regs or nil
	return rm.registries[fmt.Sprintf("%s/%s", database, collection)]
}