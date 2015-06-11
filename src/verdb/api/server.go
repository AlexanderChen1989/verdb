package api

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
)

type Server struct {
	*gin.Engine
	sess *mgo.Session
}

func NewServer(r *gin.Engine, sess *mgo.Session) *Server {
	server := &Server{r, sess}
	setupApi(server)
	setupMiddleware(server)
	return server
}
