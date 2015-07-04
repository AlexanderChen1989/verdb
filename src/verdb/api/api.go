package api

/*
Registry
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:id
  - 删除：DELETE /api/registry/:id

Versionize
	POST /api/ver
*/
func setupAPI(server *Server) {
	r := server

	// Registry CRUD
	r.POST("/api/registry", NewRegistry)
	r.POST("/api/registry/search", SearchRegistry)
	r.PUT("/api/registry/:id", UpdateRegistry)
	r.DELETE("/api/registry/:id", DeleteRegistry)

	// 版本化存储
	r.POST("/api/versionize/:database/:collection", Versionize)

	// 结果查询
	r.POST("/api/search/:database/:collection", SearchInfo)

}
