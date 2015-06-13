package api

import (
	"errors"
	"verdb/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
)

func Versionize(c *gin.Context) {
	sess := c.MustGet("sess").(*mgo.Session)
	rm := c.MustGet("rm").(*models.RegManager)
	database := c.MustGet("database").(string)
	collection := c.MustGet("collection").(string)

	reg := rm.GetReg(database, collection)
	if reg == nil {
		jsonError(c, errors.New("Cant find registry"))
		return
	}

	var newDoc map[string]interface{}
	c.Bind(&newDoc)

	if err := reg.Versionize(newDoc, sess); err != nil {
		jsonError(c, err)
		return
	}

	jsonOk(c, "Successfully versionized")
	return
}
