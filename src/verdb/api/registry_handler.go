package api

import (
	"verdb/models"

	"github.com/gin-gonic/gin"
)

/*
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:id
  - 删除：DELETE /api/registry/:id
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
	if err = rm.CreateRegistry(&reg, sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, reg)
}

func UpdateRegistry(c *gin.Context) {
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
	if err = rm.UpdateRegistry(c.Params.ByName("id"), &reg, sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, reg)
}

func SearchRegistry(c *gin.Context) {

}

func DeleteRegistry(c *gin.Context) {
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
	var reg *models.Registry

	if reg, err = rm.DeleteRegistry(c.Params.ByName("id"), sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, *reg)
}
