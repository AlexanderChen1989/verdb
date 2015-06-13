package api

import (
	"verdb/models"

	"gopkg.in/mgo.v2"

	"github.com/gin-gonic/gin"
)

// 新建：POST /api/registry
func NewRegistry(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*models.RegManager)

	var reg models.Registry
	c.Bind(&reg)

	var nreg *models.Registry
	var err error

	if nreg, err = rm.CreateRegistry(&reg, sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, *nreg)
}

// 修改：PUT /api/registry/:id
func UpdateRegistry(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*models.RegManager)

	var reg models.Registry
	c.Bind(&reg)

	var nreg *models.Registry
	var err error

	if nreg, err = rm.UpdateRegistry(c.Params.ByName("id"), &reg, sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, *nreg)
}

/*
	查询：POST /api/registry/search
	{
		query: {"field1": <value>, "field2": <value>, ...},
		sort: ["field1", "field2"],
		selection: {"field1": 1},
		limit: 10 or empty
	}
*/
func SearchRegistry(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*models.RegManager)
	var obj models.SearchStruct
	c.Bind(&obj)

	var regs []models.Registry
	var err error

	if regs, err = rm.SearchRegistries(&obj, sess); err != nil {
		jsonError(c, err)
		return
	}

	jsonOk(c, regs)
}

// 删除：DELETE /api/registry/:id
func DeleteRegistry(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*models.RegManager)
	var reg *models.Registry
	var err error

	if reg, err = rm.DeleteRegistry(c.Params.ByName("id"), sess); err != nil {
		jsonError(c, err)
		return
	}
	jsonOk(c, *reg)
}
