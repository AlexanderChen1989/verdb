package verdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type JobsManager struct {
	lock     sync.RWMutex
	scheding int
	max      int

	jobsDb   string
	jobsRepo string
}

func NewJobsManager(jdb, jrepo string, max int) *JobsManager {
	return &JobsManager{jobsDb: jdb, jobsRepo: jrepo, max: max}
}

// create new job
func (self *JobsManager) CreateJob(sess *mgo.Session, r io.Reader) (job Job, err error) {
	err = json.NewDecoder(r).Decode(&job)
	if err != nil {
		return
	}

	// add default value to job
	job.Id = bson.NewObjectId()
	job.Status = Ready

	// valid job info
	jp := &job
	switch jp.Type {
	case CJob:
		if err = (CountJob{jp}).Valid(); err != nil {
			return
		}

	case DJob:
		if err = (DistinctJob{jp}).Valid(); err != nil {
			return
		}

	case PJob:
		if err = (PipelineJob{jp}).Valid(); err != nil {
			return
		}

	case MRJob:
		if err = (MapReduceJob{jp}).Valid(); err != nil {
			return
		}

	default:
		err = errors.New("Unkonw job type: " + jp.Type + ".")
		return
	}

	// save job
	err = sess.DB(self.jobsDb).C(self.jobsRepo).Insert(job)
	return
}

func (self *JobsManager) Up() {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.scheding += 1
}

func (self *JobsManager) Down() {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.scheding -= 1
}

func (self *JobsManager) check() error {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if self.scheding > self.max {
		return errors.New("To many running jobs!")
	}

	return nil
}

// Job Schedule
func (self *JobsManager) Sched(sess *mgo.Session, query bson.M) (info interface{}, result interface{}, err error) {
	// check JobsManager status
	if err = self.check(); err != nil {
		log.Println(err)
		return
	}

	// retrieve job from db
	var job Job
	err = sess.DB(self.jobsDb).C(self.jobsRepo).Find(query).One(&job)
	if err != nil {
		log.Println(err)
		return
	}

	// check runnable or not
	jp := &job
	err = jp.checkRunnable()
	if err != nil {
		log.Println(err)
		return
	}

	// update job status
	err = sess.DB(self.jobsDb).C(self.jobsRepo).Update(bson.M{"_id": jp.Id}, bson.M{"$set": bson.M{"status": Running}})
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		nerr := sess.DB(self.jobsDb).C(self.jobsRepo).Update(bson.M{"_id": jp.Id}, bson.M{"$set": bson.M{"status": Ready}})
		if nerr != nil {
			err = fmt.Errorf("Error: %s\nError%s\n", err, nerr)
		}
	}()

	// exec job
	self.Up()
	defer self.Down()

	switch jp.Type {
	case CJob:
		var res int
		info, err = (CountJob{jp}).Exec(sess, &res)
		result = res
		return

	case DJob:
		var res []interface{}
		info, err = (DistinctJob{jp}).Exec(sess, &res)
		result = res
		return

	case PJob:
		var res []interface{}
		info, err = (PipelineJob{jp}).Exec(sess, &res)
		result = res
		return

	case MRJob:
		if jp.MapReduce.Out != nil {
			info, err = (MapReduceJob{jp}).Exec(sess, nil)
			return
		} else {
			var res []interface{}
			info, err = (MapReduceJob{jp}).Exec(sess, &res)
			result = res
			return
		}

	default:
		err = errors.New("Unkonw job type: " + jp.Type + ".")
		return
	}
}
