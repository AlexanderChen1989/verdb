package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gopkg.in/mgo.v2/bson"
)

func jsonError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, bson.M{"status": "error", "msg": err.Error()})
}

func jsonOk(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, bson.M{"status": "success", "msg": obj})
}
