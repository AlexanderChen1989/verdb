package api

import (
	"fmt"
	"verdb/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

/*
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:registryId
  - 删除：DELETE /api/registry/:registryId
*/
func NewRegistry(c *gin.Context) {
	sess, err := getDBSession(c.Get("sess"))
	if err != nil {
		jsonError(c, err)
		return
	}
	var reg models.Registry
	c.Bind(&reg)
	reg.Name = fmt.Sprintf("%s/%s", reg.DatabaseName, reg.CollectionName)
	if _, err = sess.DB(MetaDB).C(RegCollection).Upsert(bson.M{"name": reg.Name}, reg); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, reg)
}

func SearchRegistry(c *gin.Context) {

}

func UpdateRegistry(c *gin.Context) {

}

func DeleteRegistry(c *gin.Context) {

}
