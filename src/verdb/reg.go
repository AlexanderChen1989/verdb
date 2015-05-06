package verdb

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"
)

/*
{
    "db_name": "servers",
    "repo": "xxxx",
    "compare_key": "server_id",
    "index_keys": [
        ["xxxx.xxx.xxx", "xxx.xxx"]
    ],
    "ver_keys": [
        "xxxx.xxx.xxx",
        "xxx.xxx.xxx"
    ]
}
*/
type Registry struct {
	sync.Mutex
	DbName     string     `bson:"db_name" json:"db_name"`
	Repo       string     `bson:"repo" json:"repo"`
	Name       string     `bson:"name" json:"name"` // Name = DbName/Repo
	CompareKey string     `bson:"compare_key" json:"compare_key"`
	IndexKeys  [][]string `bson:"index_keys" json:"index_keys"`
	VerKeys    []string   `bson:"ver_keys" json:"ver_keys"`
	UpdateKeys []string   `bson:"update_keys" json:"update_keys"`

	verGen       func() int
	lastCleanTxn time.Time
}

func (self *Registry) GenVer() int {
	if self.verGen != nil {
		return self.verGen()
	}
	return int(time.Now().Unix() / (24 * 60 * 60))
}

// FIXME: add clean for txn data
func (self *Registry) txnRunner(sess *mgo.Session) *txn.Runner {
	// bson.M{"_id": bson.M{"$lt": bson.NewObjectIdWithTime(minus24h)}}
	repo := sess.DB(self.DbName).C("txn_collection")
	// remove txn older then 2 * 60 mins
	if self.lastCleanTxn.IsZero() {
		self.lastCleanTxn = time.Now()
	}
	if time.Now().Sub(self.lastCleanTxn) > 30*time.Minute {
		repo.RemoveAll(bson.M{
			"_id": bson.M{
				"$lt": bson.NewObjectIdWithTime(time.Now().Add(-60 * 2 * time.Minute)),
			},
		})
		self.lastCleanTxn = time.Now()
	}

	return txn.NewRunner(repo)
}

/*
* 对于提交的记录，添加_ver,_next生成新版本new
* new的格式是{pid: <int|string>, _ver: <int>, _next: <int>} 且 new._ver 相同于 new._next
* 在数据库中查询相同pid，且_next最大的版本old
* 如果 old._ver 相同于 new._ver
	* 用new的内容跟新old，然后返回
* 如果 old 相同于 new
	* 更新old版本的_next为new的_ver，然后返回
* 如果 old 不同于 new
	* 更新old版本的_next为new的_ver
	* 插入new，然后返回
*/
func (self *Registry) Versionize(newDoc map[string]interface{}, sess *mgo.Session) error {
	self.Lock()
	defer self.Unlock()

	// repo to versionize
	repo := sess.DB(self.DbName).C(self.Repo)

	// update update keys
	update := bson.M{}
	for _, key := range self.UpdateKeys {
		parts := strings.Split(key, ".")
		val := getVal(newDoc, parts)
		if val != nil {
			update[key] = val
		}
	}
	if len(update) > 0 {
		if _, err := repo.UpdateAll(
			bson.M{self.CompareKey: newDoc[self.CompareKey]},
			bson.M{"$set": update},
		); err != nil {
			return err
		}
	}

	// clean _ver, _next, _is_latest in newDoc
	delete(newDoc, "_ver")
	delete(newDoc, "_next")
	delete(newDoc, "_is_latest")

	// set up _ver, _next for new doc
	ver := self.GenVer()
	next := ver
	newDoc["_ver"] = ver
	newDoc["_next"] = next
	newDoc["_is_latest"] = true

	// query old doc from db
	var oldDoc map[string]interface{}
	err := repo.Find(bson.M{
		self.CompareKey: newDoc[self.CompareKey],
		"_is_latest":    true,
	}).One(&oldDoc)

	// oldDoc not found, just insert newDoc as first version
	if err == mgo.ErrNotFound {
		return repo.Insert(newDoc)
	} else if err != nil {
		return err
	}

	// old._ver == new._ver, update doc with new value
	if oldDoc["_ver"] == newDoc["_ver"] {
		return repo.UpdateId(oldDoc["_id"], newDoc)
	}

	// if not changed since last version, update last version's _next to now
	if !isChanged(oldDoc, newDoc, self.VerKeys) {
		setMap := bson.M{"_next": ver}
		for k, v := range newDoc {
			// keys: "_ver", "_next", "_id", "_is_latest"
			if len(k) > 0 && k[0] == '_' {
				continue
			}
			setMap[k] = v
		}
		return repo.UpdateId(
			oldDoc["_id"],
			bson.M{"$set": setMap},
		)
	}

	// if changed, update last version's _next to ver-1, insert new version
	/*
		runner := self.txnRunner(sess)
		ops := []txn.Op{{
			C:  self.Repo,
			Id: oldDoc["_id"],
			Update: bson.M{"$set": bson.M{
				"_next":      ver - 1,
				"_is_latest": false,
			}},
		}, {
			C:      self.Repo,
			Id:     bson.NewObjectId(),
			Insert: newDoc,
		}}
		id := bson.NewObjectId()
		err = runner.Run(ops, id, nil)
		if err != nil {
			runner.Resume(id)
		}
	*/
	// FIXME: transaction disabled!
	if err = repo.UpdateId(
		oldDoc["_id"],
		bson.M{"$set": bson.M{
			"_next":      ver - 1,
			"_is_latest": false,
		}},
	); err != nil {
		return err
	}
	return repo.Insert(newDoc)
}

type RegManager struct {
	db   string // db to save regs
	repo string // regs repo
	regs map[string]*Registry

	verGen func() int
	mux    sync.RWMutex
}

func NewRegManger(db, repo string, sess *mgo.Session, verGen func() int) *RegManager {
	regsRepo := sess.DB(db).C(repo)
	defer sess.Close()

	// ensure index for regs repo
	index := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	}
	err := regsRepo.EnsureIndex(index)
	if err != nil {
		log.Println(err)
		return nil
	}

	// read all regs
	var regs []Registry
	err = regsRepo.Find(nil).All(&regs)
	if err != nil {
		log.Println(err)
		return nil
	}

	// create name-> reg map
	rm := &RegManager{
		db:     db,
		repo:   repo,
		regs:   map[string]*Registry{},
		verGen: verGen,
	}
	for i, _ := range regs {
		rm.regs[regs[i].Name] = &regs[i]
	}

	return rm
}

func (self *RegManager) Size() int {
	return len(self.regs)
}

func (self *RegManager) Register(reg Registry, sess *mgo.Session) error {
	// lock rlock
	self.mux.Lock()
	defer self.mux.Unlock()

	if reg.DbName == "" || reg.Repo == "" || reg.CompareKey == "" {
		return errors.New("db_name, db_name, compare_key cant be empty!")
	}
	if len(reg.VerKeys) == 0 {
		return errors.New("ver_keys cant be empty!")
	}

	reg.Name = fmt.Sprintf("%s/%s", reg.DbName, reg.Repo)

	// save reg to regs repo
	_, err := sess.DB(self.db).C(self.repo).Upsert(bson.M{"name": reg.Name}, reg)
	if err != nil {
		return err
	}
	// save reg to regs map
	self.regs[reg.Name] = &reg

	// ensure indexs for registered repo
	regRepo := sess.DB(reg.DbName).C(reg.Repo)
	for _, index := range append(
		reg.IndexKeys,
		[]string{reg.CompareKey},
		[]string{"_ver"},
		[]string{"_next"},
		[]string{"_is_latest"},
	) {
		err = regRepo.EnsureIndexKey(index...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *RegManager) GetReg(db, repo string) *Registry {
	// lock rlock
	self.mux.RLock()
	defer self.mux.RUnlock()

	// return reg from regs or nil
	reg := self.regs[fmt.Sprintf("%s/%s", db, repo)]
	if reg != nil {
		reg.verGen = self.verGen
	}
	return reg
}
