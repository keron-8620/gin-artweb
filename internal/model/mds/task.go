package mds

import (
	"go.uber.org/zap/zapcore"

	jobmodel "gin-artweb/internal/model/job"
)

type MdsCronTask struct {
	ScriptID      uint32 `yaml:"script_id"`
	Specification string `yaml:"specification"`
}

func (mds MdsCronTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("script_id", mds.ScriptID)
	enc.AddString("specification", mds.Specification)
	return nil
}

type MdsColonyTaskRecordIDs struct {
	ColonyNum string
	Mon       uint32
	Bse       uint32
	Sse       uint32
	Szse      uint32
}

func (mds MdsColonyTaskRecordIDs) GetTaskList() []string {
	return []string{"mon", "bse", "sse", "szse"}
}

func (mds MdsColonyTaskRecordIDs) GetRecordIDs() []uint32 {
	return []uint32{mds.Mon, mds.Bse, mds.Sse, mds.Szse}
}

func (mds MdsColonyTaskRecordIDs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	recordIDs := mds.GetRecordIDs()
	for i, task := range mds.GetTaskList() {
		enc.AddUint32(task, recordIDs[i])
	}
	return nil
}

func (mds MdsColonyTaskRecordIDs) GetValidRecordIDs() []uint32 {
	var validIDs []uint32
	recordIDs := mds.GetRecordIDs()
	for _, id := range recordIDs {
		if id != 0 {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs
}

type MdsColonyTaskExecutionInfo struct {
	ColonyNum string
	Mon       *jobmodel.ScriptRecordModel
	Bse       *jobmodel.ScriptRecordModel
	Sse       *jobmodel.ScriptRecordModel
	Szse      *jobmodel.ScriptRecordModel
}

func (mds MdsColonyTaskExecutionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("colony_num", mds.ColonyNum)
	if mds.Mon != nil {
		enc.AddObject("mon", mds.Mon)
	}
	if mds.Bse != nil {
		enc.AddObject("bse", mds.Bse)
	}
	if mds.Sse != nil {
		enc.AddObject("sse", mds.Sse)
	}
	if mds.Szse != nil {
		enc.AddObject("szse", mds.Szse)
	}
	return nil
}
