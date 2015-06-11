package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"

	"verdb/api"
)

func main() {
	sess, _ := mgo.Dial("localhost")
	r := gin.Default()
	server := api.NewServer(r, sess)
	server.Run(":8080")
}
