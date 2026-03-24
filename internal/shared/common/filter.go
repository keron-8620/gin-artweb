package common

// Page2LimitOffset 将页码page、页大小size转换为limit和offset
// 参数说明：
//
//	page: 页码（从1开始，若传入<=0则自动转为1）
//	size: 页大小（若传入<=0则自动转为默认值10，可根据业务调整）
//
// 返回值：
//
//	limit: 每页条数
//	offset: 偏移量（(page-1)*size）
func Page2LimitOffset(page, size int) (limit, offset int) {
	// 处理非法页码（确保page>=1）
	if page <= 0 {
		page = 1
	}
	// 处理非法页大小（设置默认值）
	defaultSize := 10
	if size <= 0 {
		size = defaultSize
	}
	// 限制最大页大小（防止一次性查询过多数据，可选）
	maxSize := 100
	if size > maxSize {
		size = maxSize
	}

	limit = size
	offset = (page - 1) * size
	return
}
