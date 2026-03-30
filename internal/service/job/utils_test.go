package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
)

func TestGetRecordIDByMap(t *testing.T) {
	tests := []struct {
		name     string
		cache    map[uint32]jobmodel.ScriptRecordModel
		recordID uint32
		wantNil  bool
	}{
		{
			name: "record exists in cache",
			cache: map[uint32]jobmodel.ScriptRecordModel{
				1: {
					StandardModel: database.StandardModel{
						BaseModel: database.BaseModel{
							ID: 1,
						},
					},
					Status: 2,
				},
			},
			recordID: 1,
			wantNil:  false,
		},
		{
			name:     "record does not exist in cache",
			cache:    map[uint32]jobmodel.ScriptRecordModel{},
			recordID: 999,
			wantNil:  true,
		},
		{
			name:     "nil cache",
			cache:    nil,
			recordID: 1,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRecordIDByMap(tt.cache, tt.recordID)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.recordID, result.ID)
			}
		})
	}
}

func TestBuildTaskInfoFromScriptRecord(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		taskName   string
		m          *jobmodel.ScriptRecordModel
		wantNil    bool
		expectID   uint32
		expectStatus int
	}{
		{
			name:      "nil model returns empty BizTaskInfo",
			taskName:  "test_task",
			m:         nil,
			wantNil:   false,
			expectID:  0,
			expectStatus: 0,
		},
		{
			name:     "valid model returns correct BizTaskInfo",
			taskName: "test_task",
			m: &jobmodel.ScriptRecordModel{
				StandardModel: database.StandardModel{
					CreatedAt: now,
					UpdatedAt: now,
				},
				Status:      2,
				TriggerType: "cron",
			},
			wantNil:       false,
			expectID:      0,
			expectStatus: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTaskInfoFromScriptRecord(tt.taskName, tt.m)
			assert.Equal(t, tt.taskName, result.TaskName)
			if tt.m != nil {
				assert.Equal(t, tt.m.ID, result.RecordID)
				assert.Equal(t, tt.m.Status, result.Status)
				assert.Equal(t, tt.m.CreatedAt.Format(time.DateTime), result.StartTime)
				assert.Equal(t, tt.m.UpdatedAt.Format(time.DateTime), result.EndTime)
				assert.Equal(t, tt.m.TriggerType, result.TriggerType)
			} else {
				assert.Equal(t, uint32(0), result.RecordID)
				assert.Equal(t, 0, result.Status)
				assert.Equal(t, "", result.StartTime)
				assert.Equal(t, "", result.EndTime)
				assert.Equal(t, "", result.TriggerType)
			}
		})
	}
}
