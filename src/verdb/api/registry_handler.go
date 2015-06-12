package api

import (
	"verdb/models"

	"github.com/gin-gonic/gin"
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
	rm, err := getRegManager(c.Get("rm"))
	if err != nil {
		jsonError(c, err)
		return
	}
	var reg models.Registry
	c.Bind(&reg)
	if err = rm.Register(&reg, sess); err != nil {
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
