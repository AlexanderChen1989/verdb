package api

/*
  - 新建：POST /api/registry
  - 查询：POST /api/registry/search
  - 修改：PUT /api/registry/:registryId
  - 删除：DELETE /api/registry/:registryId
*/
func setupApi(server *Server) {
	r := server
	r.POST("/api/registry", NewRegistry)
	r.POST("/api/registry/search", SearchRegistry)
	r.PUT("/api/registry/:registryId", UpdateRegistry)
	r.DELETE("/api/registry/:registryId", DeleteRegistry)
}
