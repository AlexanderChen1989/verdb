package api

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// SearchInfo 收集特定表的数据，使用mongodb自带是搜寻语法
func SearchInfo(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	database, collection := c.Param("database"), c.Param("collection")
	var query bson.M
	if err := c.BindJSON(&query); err != nil {
		jsonError(c, err)
		return
	}
	result, err := searchInfo(database, collection, query, sess)
	if err != nil {
		jsonError(c, err)
		return
	}

	jsonOk(c, bson.M{"result": result})
}

func searchInfo(database, collection string, query bson.M, sess *mgo.Session) ([]bson.M, error) {
	var result []bson.M
	if err := sess.DB(database).C(collection).Find(query).All(&result); err != nil {
		return nil, err
	}
	return result, nil
}
