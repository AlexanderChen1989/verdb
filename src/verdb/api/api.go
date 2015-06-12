package api

/*
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:id
  - 删除：DELETE /api/registry/:id
*/
func setupApi(server *Server) {
	r := server
	r.POST("/api/registry", NewRegistry)
	r.POST("/api/registry/search", SearchRegistry)
	r.PUT("/api/registry/:id", UpdateRegistry)
	r.DELETE("/api/registry/:id", DeleteRegistry)
}
