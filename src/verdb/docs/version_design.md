# 数据版本化设计

---

## 版本化格式
	# versions: 0 -- 3 -- 5 -- 6
	{_ver: 0, _next: 2}
	{_ver: 3, _next: 4}
	{_ver: 5, _next: 6}

## 版本存储过程
* 对于提交的记录，添加\_ver,\_next生成新版本new
* new的格式是{pid: <int|string>, _ver: <int>, _next: <int>} 且 new.\_ver 相同于 new.\_next
* 在数据库中查询相同pid，且_next最大的版本old
* 如果 old._ver 相同于 new._ver
	* 用new的内容跟新old，然后返回
* 如果 old 相同于 new
	* 更新old版本的\_next为new的\_ver，然后返回
* 如果 old 不同于 new
	* 更新old版本的\_next为new的\_ver
	* 插入new，然后返回



