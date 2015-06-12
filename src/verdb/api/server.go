package api

import (
	"verdb/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
)

type Server struct {
	*gin.Engine
	sess *mgo.Session
	rm   *models.RegManager
}

func NewServer(r *gin.Engine, sess *mgo.Session) *Server {
	rm := models.NewRegManger(MetaDB, RegCollection, sess)
	server := &Server{r, sess, rm}
	setupMiddleware(server)
	setupApi(server)
	return server
}
