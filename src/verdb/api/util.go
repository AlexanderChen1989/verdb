package api

import (
	"errors"
	"net/http"
	"verdb/models"

	"github.com/gin-gonic/gin"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getDBSession(sess interface{}, err error) (*mgo.Session, error) {
	if err != nil {
		return nil, err
	}
	_sess, ok := sess.(*mgo.Session)
	if !ok {
		return nil, errors.New("interface{} cant convert to *mgo.Session")
	}
	return _sess, nil
}

func getRegManager(rm interface{}, err error) (*models.RegManager, error) {
	if err != nil {
		return nil, err
	}
	_rm, ok := rm.(*models.RegManager)
	if !ok {
		return nil, errors.New("interface{} cant convert to *RegManger")
	}
	return _rm, nil
}

func jsonError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, bson.M{"status": "error", "msg": err.Error()})
}

func jsonOk(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, bson.M{"status": "success", "msg": obj})
}
