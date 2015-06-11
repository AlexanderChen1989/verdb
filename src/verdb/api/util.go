package api

import (
	"errors"

	"gopkg.in/mgo.v2"
)

func getDBSession(sess interface{}, err error) (*mgo.Session, error) {
	if err != nil {
		return nil, err
	}
	dbsess, ok := sess.(*mgo.Session)
	if !ok {
		return nil, errors.New("interface{} cant convert to *mgo.Session")
	}
	return dbsess, nil
}
