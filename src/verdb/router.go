package verdb

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Server struct {
	*gin.Engine
	sess *mgo.Session
}

// verGen for test purpose
func NewApiServer(mgoUrl string, metaDb string, max int, verGen func() int) *Server {
	sess, err := mgo.Dial(mgoUrl)
	if err != nil {
		panic(err)
	}
	rm := NewRegManger(metaDb, "regs", sess.Clone(), verGen)
	jm := NewJobsManager(metaDb, "jobs", max)

	r := gin.New()

	// middlewares
	r.Use(recoverMiddleware)
	r.Use(gin.Logger())
	r.Use(func(c *gin.Context) {
		tsess := sess.Clone()
		defer tsess.Close()
		c.Set("sess", tsess)
		c.Set("rm", rm)
		c.Set("jm", jm)
		c.Next()
	})

	// routes
	// register route
	r.POST("/api/register", errWrapper(registerRoute))
	// versionize document route
	r.POST("/api/versionize/:db/:repo", errWrapper(versionizeRoute))
	// direct insert document to db route
	r.POST("/api/upsert/:db/:repo", errWrapper(upsertRoute))
	// search route
	r.POST("/api/search/:db/:repo", errWrapper(seachRoute))
	// create job
	r.POST("/api/job", errWrapper(createJobRoute))
	// schedule job
	r.GET("/api/job/:id/sched", errWrapper(schedJobRoute))

	return &Server{r, sess}
}

func (self *Server) Close() {
	self.sess.Close()
}

func (self *Server) Clone() *mgo.Session {
	return self.sess.Clone()
}

func recoverMiddleware(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, true)
			log.Printf("Panic Error: %s\n", err)
			log.Printf("Stack Trace: %s", buf[:n])
			c.JSON(
				http.StatusInternalServerError,
				bson.M{
					"status": "error",
					"msg":    fmt.Sprintf("%s", err),
				},
			)
			return
		}
	}()
	c.Next()
}

func errWrapper(f func(c *gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
			return
		}
	}
}
