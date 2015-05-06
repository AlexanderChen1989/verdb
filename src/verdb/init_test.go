package verdb

import (
	"log"
	"sync"

	"gopkg.in/mgo.v2"
)

var (
	dbsess *mgo.Session
	oncedb sync.Once
)

func initdb() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var err error
	dbsess, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
}
