package verdb

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
)

const (
	testDb   = "testdb"
	regsRepo = "regs"
)

var _sess *mgo.Session

func getSession(t *testing.T) *mgo.Session {
	if _sess != nil {
		return _sess
	}
	var err error
	_sess, err = mgo.Dial("localhost")
	if err != nil {
		t.Fatal(err)
	}
	return _sess
}

var rmTest = struct{ db, repo string }{"testdb", "testregs"}

func TestNewRegManager(t *testing.T) {
	Convey("Test New RegManager", t, func() {
		sess, err := mgo.Dial("localhost")
		So(err, ShouldBeNil)
		defer sess.Close()

		// do some clean
		sess.DB(rmTest.db).C(rmTest.repo).DropCollection()
		var rm = NewRegManger(rmTest.db, rmTest.repo, sess.Clone(), nil)
		So(rm, ShouldNotBeNil)
		indexes, err := sess.DB(rmTest.db).C(rmTest.repo).Indexes()
		So(err, ShouldBeNil)
		fmt.Printf("%+v\n", indexes)
		So(len(indexes), ShouldEqual, 2)
		So(indexes[0].Key, ShouldResemble, []string{"_id"})
		So(indexes[1].Key, ShouldResemble, []string{"name"})
	})
}

//
func contentEq(lsta, lstb [][]string) bool {
	if len(lsta) != len(lstb) {
		return false
	}
	lstbIndexs := make([]bool, len(lstb))
	for _, ia := range lsta {
		has := false
		for i, ib := range lstb {
			if lstbIndexs[i] {
				continue
			}
			if reflect.DeepEqual(ia, ib) {
				lstbIndexs[i] = true
				has = true
				break
			}
		}
		if !has {
			return false
		}
	}
	return true
}

func TestRegister(t *testing.T) {
	Convey("Test Register", t, func() {
		var (
			comKey    = "server_id"
			indexKeys = [][]string{
				{"a"},
				{"a", "b"},
				{"a", "b", "c"},
			}
			vers = []string{
				"a.b.c.d",
			}
			updates = []string{
				"a.b.c.e",
			}

			regs []Registry
		)

		for i := 0; i < 10; i++ {
			repo := fmt.Sprintf("repo%d", i)
			name := rmTest.db + "/" + repo
			regs = append(
				regs,
				Registry{
					DbName:     rmTest.db,
					Repo:       repo,
					Name:       name,
					CompareKey: comKey,
					IndexKeys:  indexKeys,
					VerKeys:    vers,
					UpdateKeys: updates,
				},
			)
		}

		sess, err := mgo.Dial("localhost")
		So(err, ShouldBeNil)
		defer sess.Close()

		// do some clean
		sess.DB(rmTest.db).C(rmTest.repo).DropCollection()
		for _, reg := range regs {
			sess.DB(reg.DbName).C(reg.Repo).DropCollection()
		}

		rm := NewRegManger(rmTest.db, rmTest.repo, sess.Clone(), nil)
		So(rm, ShouldNotBeNil)

		// test register reg
		for _, reg := range regs {
			rm.Register(reg, sess)
		}

		// test update RegsManager's regs.
		So(len(rm.regs), ShouldEqual, len(regs))

		nrm := NewRegManger(rmTest.db, rmTest.repo, sess.Clone(), nil)
		So(nrm, ShouldNotBeNil)
		So(len(nrm.regs), ShouldEqual, len(regs))

		// test find reg from RegManager
		for _, reg := range regs {
			So(*nrm.GetReg(reg.DbName, reg.Repo), ShouldResemble, reg)
		}

		// test reg's repo's index
		for _, reg := range regs {
			indexs, _ := sess.DB(reg.DbName).C(reg.Repo).Indexes()
			var regIndexs [][]string
			for _, index := range indexs {
				regIndexs = append(regIndexs, index.Key)
			}

			keys := append(
				append(
					[][]string{
						{"_id"},
						{"_ver"},
						{"_next"},
						{"_is_latest"},
					},
					reg.IndexKeys...,
				),
				[]string{comKey},
			)
			So(contentEq(regIndexs, keys), ShouldBeTrue)
		}
	})
}
