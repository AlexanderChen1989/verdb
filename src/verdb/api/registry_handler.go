package api

import "github.com/gin-gonic/gin"

/*
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:registryId
  - 删除：DELETE /api/registry/:registryId
*/
func NewRegistry(c *gin.Context) {
	_, err := getDBSession(c.Get("sess"))
	if err != nil {
		panic(err)
	}
}

func SearchRegistry(c *gin.Context) {

}

func UpdateRegistry(c *gin.Context) {

}

func DeleteRegistry(c *gin.Context) {

}
