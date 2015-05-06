package verdb

import (
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func upsert(idoc, udoc bson.M, dbName, repoName string, rm *RegManager, sess *mgo.Session) error {
	// get registry from RegManager
	reg := rm.GetReg(dbName, repoName)
	if reg == nil {
		return errors.New(dbName + "/" + repoName + " is not registered.")

	}

	repo := sess.DB(dbName).C(repoName)
	n, err := repo.Find(
		bson.M{
			reg.CompareKey: idoc[reg.CompareKey],
		},
	).Count()

	if err != nil {
		return err
	}

	if n > 0 {
		if udoc == nil {
			return nil
		}
		return repo.Update(
			bson.M{reg.CompareKey: idoc[reg.CompareKey]},
			udoc,
		)
	}

	return repo.Insert(idoc)
}
