package job

import (
	"time"

	jobmodel "gin-artweb/internal/model/job"
)

func GetRecordIDByMap(
	cache map[uint32]jobmodel.ScriptRecordModel,
	recordID uint32,
) *jobmodel.ScriptRecordModel {
	task, exists := cache[recordID]
	if !exists {
		return nil
	}
	return &task
}

func BuildTaskInfoFromScriptRecord(
	taskName string,
	m *jobmodel.ScriptRecordModel,
) jobmodel.BizTaskInfo {
	result := jobmodel.BizTaskInfo{
		TaskName: taskName,
	}
	if m != nil {
		result.RecordID = m.ID
		result.Status = m.Status
		result.StartTime = m.CreatedAt.Format(time.DateTime)
		result.EndTime = m.UpdatedAt.Format(time.DateTime)
		result.TriggerType = m.TriggerType
	}
	return result
}
