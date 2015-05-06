package verdb

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// job type
	CJob  = "CountJob"
	DJob  = "DistinctJob"
	PJob  = "PipelineJob"
	MRJob = "MapReduceJob"

	// job status
	Ready   = "Ready"
	Running = "Running"
)

type Job struct {
	Id         bson.ObjectId `bson:"_id" json:"_id"`
	Name       string        `bson:"name" json:"name"`
	Type       string        `bson:"type" json:"type"`
	TargetDB   string        `bson:"target_db" json:"target_db"`
	TargetRepo string        `bson:"target_repo" json:"target_repo"`
	Status     string        `bson:"status" json:"status"`

	Query       bson.M        `bson:"query" json:"query"`
	DistinctKey string        `bson:"distinct_key" json:"distinct_key"`
	MapReduce   mgo.MapReduce `bson:"map_reduce" json:"map_reduce"`
	Pipeline    []bson.M      `bson:"pipeline" json:"pipeline"`
}

// TODO: add validation!
func (self *Job) Valid() error {
	return nil
}

func (self *Job) checkRunnable() error {
	if self.Status != Ready {
		return errors.New("Job is busy!")
	}
	return nil
}

type CountJob struct{ *Job }

func (self CountJob) Exec(sess *mgo.Session, result interface{}) (info interface{}, err error) {
	if err = self.checkRunnable(); err != nil {
		return
	}
	log.Printf("count job....\n%++v\n", self.Job)

	*(result.(*int)), err = sess.DB(self.TargetDB).C(self.TargetRepo).Find(self.Query).Count()
	return
}

type DistinctJob struct{ *Job }

func (self DistinctJob) Valid() (err error) {
	err = self.Job.Valid()
	if err != nil {
		return
	}
	// TODO: add specific validations.
	return nil
}

func (self DistinctJob) Exec(sess *mgo.Session, result interface{}) (info interface{}, err error) {
	if err = self.checkRunnable(); err != nil {
		return
	}

	err = sess.DB(self.TargetDB).C(self.TargetRepo).Find(self.Query).Distinct(self.DistinctKey, result)
	return
}

type PipelineJob struct{ *Job }

func (self PipelineJob) Valid() (err error) {
	err = self.Job.Valid()
	if err != nil {
		return
	}
	// TODO: add specific validations.
	return nil
}

func (self PipelineJob) Exec(sess *mgo.Session, result interface{}) (info interface{}, err error) {
	if err = self.checkRunnable(); err != nil {
		return
	}

	err = sess.DB(self.TargetDB).C(self.TargetRepo).Pipe(self.Pipeline).All(result)
	return
}

type MapReduceJob struct{ *Job }

func (self MapReduceJob) Valid() (err error) {
	err = self.Job.Valid()
	if err != nil {
		return
	}
	// TODO: add specific validations.
	return nil
}

func (self MapReduceJob) Exec(sess *mgo.Session, result interface{}) (info interface{}, err error) {
	if err = self.checkRunnable(); err != nil {
		return
	}
	info, err = sess.DB(self.TargetDB).C(self.TargetRepo).Find(self.Query).MapReduce(&self.Job.MapReduce, result)
	log.Printf("mr job....\n%+v\n%v\n", info, err)
	return
}
