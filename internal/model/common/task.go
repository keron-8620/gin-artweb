package common

type TaskInfo struct {
	// 任务名称
	TaskName string `json:"task_name" example:"mon"`

	// 执行记录ID(0表示非正常执行的任务)
	RecordID uint32 `json:"record_id"`

	// 执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)
	Status int `json:"status" example:"2"`

	// 创建时间
	StartTime string `json:"start_time" example:"2023-01-01 12:00:00"`

	// 更新时间
	EndTime string `json:"end_time" example:"2023-01-01 12:00:00"`

	// 触发类型(cron/api,未执行为空)
	TriggerType string `json:"trigger_type" example:"cron"`
}
