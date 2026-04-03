package oes

import (
	"go.uber.org/zap/zapcore"

	jobmodel "gin-artweb/internal/model/job"
)

type OesCronTask struct {
	ScriptID      uint32 `yaml:"script_id"`
	Specification string `yaml:"specification"`
}

func (oes OesCronTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("script_id", oes.ScriptID)
	enc.AddString("specification", oes.Specification)
	return nil
}

type StkColonyTaskRecordIDs struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Bse               uint32
	Sse               uint32
	Szse              uint32
	Csdc              uint32
}

func (stk StkColonyTaskRecordIDs) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "bse", "sse", "szse", "csdc"}
}

func (stk StkColonyTaskRecordIDs) GetRecordIDs() []uint32 {
	return []uint32{stk.Mon, stk.CounterFetch, stk.CounterDistribute, stk.Bse, stk.Sse, stk.Szse, stk.Csdc}
}

func (stk StkColonyTaskRecordIDs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	recordIDs := stk.GetRecordIDs()
	for i, task := range stk.GetTaskList() {
		enc.AddUint32(task, recordIDs[i])
	}
	return nil
}

func (stk StkColonyTaskRecordIDs) GetValidRecordIDs() []uint32 {
	var validIDs []uint32
	recordIDs := stk.GetRecordIDs()
	for _, id := range recordIDs {
		if id != 0 {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs
}

type StkColonyTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *jobmodel.ScriptRecordModel
	CounterFetch      *jobmodel.ScriptRecordModel
	CounterDistribute *jobmodel.ScriptRecordModel
	Bse               *jobmodel.ScriptRecordModel
	Sse               *jobmodel.ScriptRecordModel
	Szse              *jobmodel.ScriptRecordModel
	Csdc              *jobmodel.ScriptRecordModel
}

func (stk StkColonyTaskExecutionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("colony_num", stk.ColonyNum)
	if stk.Mon != nil {
		enc.AddObject("mon", stk.Mon)
	}
	if stk.CounterFetch != nil {
		enc.AddObject("counter_fetch", stk.CounterFetch)
	}
	if stk.CounterDistribute != nil {
		enc.AddObject("counter_distribute", stk.CounterDistribute)
	}
	if stk.Bse != nil {
		enc.AddObject("bse", stk.Bse)
	}
	if stk.Sse != nil {
		enc.AddObject("sse", stk.Sse)
	}
	if stk.Szse != nil {
		enc.AddObject("szse", stk.Szse)
	}
	if stk.Csdc != nil {
		enc.AddObject("csdc", stk.Csdc)
	}
	return nil
}

type CrdColonyTaskRecordIDs struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Sse               uint32
	Szse              uint32
	Csdc              uint32
	SseLate           uint32
	SzseLate          uint32
}

func (crd CrdColonyTaskRecordIDs) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "sse", "szse", "csdc", "szse_late", "sse_late"}
}

func (crd CrdColonyTaskRecordIDs) GetRecordIDs() []uint32 {
	return []uint32{crd.Mon, crd.CounterFetch, crd.CounterDistribute, crd.Sse, crd.Szse, crd.Csdc, crd.SseLate, crd.SzseLate}
}

func (crd CrdColonyTaskRecordIDs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	recordIDs := crd.GetRecordIDs()
	for i, task := range crd.GetTaskList() {
		enc.AddUint32(task, recordIDs[i])
	}
	return nil
}

func (crd CrdColonyTaskRecordIDs) GetValidRecordIDs() []uint32 {
	var validIDs []uint32
	recordIDs := crd.GetRecordIDs()
	for _, id := range recordIDs {
		if id != 0 {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs
}

type CrdColonyTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *jobmodel.ScriptRecordModel
	CounterFetch      *jobmodel.ScriptRecordModel
	CounterDistribute *jobmodel.ScriptRecordModel
	Sse               *jobmodel.ScriptRecordModel
	Szse              *jobmodel.ScriptRecordModel
	Csdc              *jobmodel.ScriptRecordModel
	SseLate           *jobmodel.ScriptRecordModel
	SzseLate          *jobmodel.ScriptRecordModel
}

func (crd CrdColonyTaskExecutionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("colony_num", crd.ColonyNum)
	if crd.Mon != nil {
		enc.AddObject("mon", crd.Mon)
	}
	if crd.CounterFetch != nil {
		enc.AddObject("counter_fetch", crd.CounterFetch)
	}
	if crd.CounterDistribute != nil {
		enc.AddObject("counter_distribute", crd.CounterDistribute)
	}
	if crd.Sse != nil {
		enc.AddObject("sse", crd.Sse)
	}
	if crd.Szse != nil {
		enc.AddObject("szse", crd.Szse)
	}
	if crd.Csdc != nil {
		enc.AddObject("csdc", crd.Csdc)
	}
	if crd.SseLate != nil {
		enc.AddObject("sse_late", crd.SseLate)
	}
	if crd.SzseLate != nil {
		enc.AddObject("szse_late", crd.SzseLate)
	}
	return nil
}

type OptColonyTaskRecordIDs struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Sse               uint32
	Szse              uint32
}

func (opt OptColonyTaskRecordIDs) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "sse", "szse"}
}

func (opt OptColonyTaskRecordIDs) GetRecordIDs() []uint32 {
	return []uint32{opt.Mon, opt.CounterFetch, opt.CounterDistribute, opt.Sse, opt.Szse}
}

func (opt OptColonyTaskRecordIDs) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	recordIDs := opt.GetRecordIDs()
	for i, task := range opt.GetTaskList() {
		enc.AddUint32(task, recordIDs[i])
	}
	return nil
}

func (opt OptColonyTaskRecordIDs) GetValidRecordIDs() []uint32 {
	var validIDs []uint32
	recordIDs := opt.GetRecordIDs()
	for _, id := range recordIDs {
		if id != 0 {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs
}

type OptColonyTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *jobmodel.ScriptRecordModel
	CounterFetch      *jobmodel.ScriptRecordModel
	CounterDistribute *jobmodel.ScriptRecordModel
	Sse               *jobmodel.ScriptRecordModel
	Szse              *jobmodel.ScriptRecordModel
}

func (opt OptColonyTaskExecutionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("colony_num", opt.ColonyNum)
	if opt.Mon != nil {
		enc.AddObject("mon", opt.Mon)
	}
	if opt.CounterFetch != nil {
		enc.AddObject("counter_fetch", opt.CounterFetch)
	}
	if opt.CounterDistribute != nil {
		enc.AddObject("counter_distribute", opt.CounterDistribute)
	}
	if opt.Sse != nil {
		enc.AddObject("sse", opt.Sse)
	}
	if opt.Szse != nil {
		enc.AddObject("szse", opt.Szse)
	}
	return nil
}
