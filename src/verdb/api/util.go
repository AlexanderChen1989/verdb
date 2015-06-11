package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getDBSession(sess interface{}, err error) (*mgo.Session, error) {
	if err != nil {
		return nil, err
	}
	dbsess, ok := sess.(*mgo.Session)
	if !ok {
		return nil, errors.New("interface{} cant convert to *mgo.Session")
	}
	return dbsess, nil
}

func jsonError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, bson.M{"status": "error", "msg": err.Error()})
}

func jsonOk(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, bson.M{"status": "success", "msg": obj})
}
