package common


type TaskInfo struct { 
	// 任务名称
	TaskName string `json:"task_name" example:"mon"`

	// 执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)
	TaskTastus string `json:"task_status" example:"--"`

	// 执行记录ID(0表示非正常执行的任务)
	RecordID uint32 `json:"record_id"`
}
