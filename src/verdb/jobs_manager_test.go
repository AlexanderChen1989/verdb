package verdb

import (
	"bytes"
	"sync"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

var (
	onceJobsInit sync.Once
)

func jobsinit() {
	// clean
	testRepo := dbsess.DB("testdb").C("testrepo")
	jobRepo := dbsess.DB("testdb").C("jobrepo")
	testRepo.DropCollection()
	jobRepo.DropCollection()

	// insert test data
	for i := 0; i < 1000; i++ {
		testRepo.Insert(bson.M{"a": i, "b": i / 20})
	}
}

func initJobsTest() {
	oncedb.Do(initdb)
	onceJobsInit.Do(jobsinit)
}

func TestSceduleCountJob(t *testing.T) {
	// init db
	initJobsTest()
	sess := dbsess.Clone()
	defer sess.Close()

	jm := NewJobsManager("testdb", "jobrepo", 10)

	// test schedule simple job
	jobJson := bytes.NewReader([]byte(`{
		"name": "sjb-count",
		"type": "CountJob",
		"target_db": "testdb",
		"target_repo": "testrepo",
		"query": {"a": {"$gte": 100}}
	}`))
	job, err := jm.CreateJob(sess, jobJson)
	if err != nil {
		t.Errorf("Failed to create job: %s\n%v\n", err, job)
		return
	}
	// count
	_, result, err := jm.Sched(sess, bson.M{"_id": job.Id})
	if err != nil {
		t.Errorf("Failed to schedule job: %v\n%s\n", job, err)
		return
	}
	if result != 900 {
		t.Errorf("Failed to count: %v\n", result)
		return
	}
}

func TestScheduleDistinctJob(t *testing.T) {
	// initdb
	initJobsTest()
	sess := dbsess.Clone()
	defer sess.Close()

	// schedule distinct job
	jm := NewJobsManager("testdb", "jobrepo", 10)
	jobJson := bytes.NewReader([]byte(`{
		"name": "sjb-distinct",
		"type": "DistinctJob",
		"target_db": "testdb",
		"target_repo": "testrepo",
		"distinct_key": "a",
		"query": {"a": {"$gte": 100}}
	}`))
	job, err := jm.CreateJob(sess, jobJson)
	if err != nil {
		t.Errorf("Failed to create job: %s\n%v\n", err, job)
		return
	}
	// distinct
	_, result, err := jm.Sched(sess, bson.M{"_id": job.Id})
	if err != nil {
		t.Errorf("Failed to schedule job: %v\n%s\n", job, err)
		return
	}
	if len(result.([]interface{})) != 900 {
		t.Errorf("Failed to distinct: %v\n", result)
		return
	}
}

func TestScheduleMapReduceJobInline(t *testing.T) {
	// initdb
	initJobsTest()
	sess := dbsess.Clone()
	defer sess.Close()

	// schedule distinct job
	jm := NewJobsManager("testdb", "jobrepo", 10)
	// mapreduce job inline
	jobJson := bytes.NewReader([]byte(`{
		"name": "mrj-inline",
		"type": "MapReduceJob",
		"target_db": "testdb",
		"target_repo": "testrepo",
		"query": {"a": {"$gte": 100}},
		"map_reduce" : {
	        "map" : "function() { emit(this.b, this.a) }",
	        "reduce" : "function(key, values) { return Array.sum(values) }"
	    }
	}`))
	job, err := jm.CreateJob(sess, jobJson)
	if err != nil {
		t.Errorf("Failed to create job: %s\n%v\n", err, job)
		return
	}
	info, result, err := jm.Sched(sess, bson.M{"_id": job.Id})
	if err != nil {
		t.Errorf("Failed to schedule job: %v\n%s\n", job, err)
		return
	}
	t.Logf("Result: %d Info: %++v\n", len(result.([]interface{})), info)
}

func TestScheduleMapReduceJobOut(t *testing.T) {
	// initdb
	initJobsTest()
	sess := dbsess.Clone()
	defer sess.Close()

	// schedule distinct job
	jm := NewJobsManager("testdb", "jobrepo", 10)
	// mapreduce job out to collection results
	jobJson := bytes.NewReader([]byte(`{
		"name": "mrj-out",
		"type": "MapReduceJob",
		"target_db": "testdb",
		"target_repo": "testrepo",
		"query": {"a": {"$gte": 100}},
		"map_reduce" : {
	        "map" : "function() { emit(this.b, this.a) }",
	        "reduce" : "function(key, values) { return Array.sum(values) }",
	        "out": "results"
	    }
	}`))
	job, err := jm.CreateJob(sess, jobJson)
	if err != nil {
		t.Errorf("Failed to create job: %s\n%v\n", err, job)
		return
	}
	info, _, err := jm.Sched(sess, bson.M{"_id": job.Id})
	if err != nil {
		t.Errorf("Failed to schedule job: %v\n%v\n%s\n", info, job, err)
		return
	}
}

func TestSchedulePipelineJob(t *testing.T) {
	// initdb
	initJobsTest()
	sess := dbsess.Clone()
	defer sess.Close()

	// schedule distinct job
	jm := NewJobsManager("testdb", "jobrepo", 10)
	// mapreduce job out to collection results
	jobJson := bytes.NewReader([]byte(`{
		"name": "pj",
		"type": "PipelineJob",
		"target_db": "testdb",
		"target_repo": "testrepo",
		"pipeline" : [{"$match": { "a": {"$gte": 700}}}, {"$group": {"_id": "$b", "total": {"$sum": "$a"}}}]
	}`))
	job, err := jm.CreateJob(sess, jobJson)
	if err != nil {
		t.Errorf("Failed to create job: %s\n%v\n", err, job)
		return
	}
	_, result, err := jm.Sched(sess, bson.M{"_id": job.Id})
	if err != nil {
		t.Errorf("Failed to schedule job: %v\n%s\n", job, err)
		return
	}
	if len(result.([]interface{})) != 15 {
		t.Errorf("Wrong result length: %v\n", result)
	}
}
