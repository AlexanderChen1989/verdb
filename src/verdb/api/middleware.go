package api

import "github.com/gin-gonic/gin"

func setupMiddleware(svr *Server) {
	svr.Use(func(c *gin.Context) {
		sess := svr.sess.Clone()
		c.Set("sess", sess)
		defer sess.Close()
		c.Next()
	})
	svr.Use(func(c *gin.Context) {
		c.Set("rm", svr.rm)
	})
}
