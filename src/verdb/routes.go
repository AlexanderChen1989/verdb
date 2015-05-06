package verdb

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func registerRoute(c *gin.Context) error {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*RegManager)
	var reg Registry
	err := json.NewDecoder(c.Request.Body).Decode(&reg)
	if err != nil {
		return err
	}
	err = rm.Register(reg, sess)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

type ErrDoc struct {
	Err string `json:"err"`
	Doc bson.M `json:"doc"`
}

func versionizeRoute(c *gin.Context) error {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*RegManager)

	db := c.Params.ByName("db")
	repo := c.Params.ByName("repo")

	reg := rm.GetReg(db, repo)
	if reg == nil {
		return errors.New("cant find registry for " + db + "/" + repo)
	}

	var rdocs []map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&rdocs)
	if err != nil {
		return err
	}
	var errs []ErrDoc
	for _, rdoc := range rdocs {
		err = reg.Versionize(rdoc, sess)
		if err != nil {
			errs = append(
				errs,
				ErrDoc{err.Error(), rdoc},
			)
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": errs})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}

	return nil
}

func upsertRoute(c *gin.Context) error {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*RegManager)

	db := c.Params.ByName("db")
	repo := c.Params.ByName("repo")

	reg := rm.GetReg(db, repo)
	if reg == nil {
		return errors.New("cant find registry for " + db + "/" + repo)
	}

	var body struct {
		Insert bson.M
		Update bson.M
	}
	err := json.NewDecoder(c.Request.Body).Decode(&body)
	if err != nil {
		return err
	}

	err = upsert(body.Insert, body.Update, db, repo, rm, sess)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil

}

func seachRoute(c *gin.Context) error {
	sess := c.MustGet("sess").(*mgo.Session)
	db := c.Params.ByName("db")
	repo := c.Params.ByName("repo")

	var query bson.M
	err := json.NewDecoder(c.Request.Body).Decode(&query)
	if err != nil {
		return err
	}
	var result []bson.M
	err = sess.DB(db).C(repo).Find(query).All(&result)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, bson.M{"status": "success", "payload": result})
	return nil
}

func createJobRoute(c *gin.Context) error {
	jm := c.MustGet("jm").(*JobsManager)
	sess := c.MustGet("sess").(*mgo.Session)

	job, err := jm.CreateJob(sess, c.Request.Body)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "msg": job})

	return nil
}

func schedJobRoute(c *gin.Context) error {
	idStr := c.Params.ByName("id")
	if !bson.IsObjectIdHex(idStr) {
		return errors.New("Invalid job id: " + idStr)
	}

	jm := c.MustGet("jm").(*JobsManager)
	sess := c.MustGet("sess").(*mgo.Session)
	id := bson.ObjectIdHex(idStr)
	info, result, err := jm.Sched(sess, bson.M{"_id": id})
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, bson.M{"status": "success", "info": info, "result": result})
	return nil
}
