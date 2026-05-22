package oes

import (
	"go.uber.org/zap/zapcore"

	jobmodel "gin-artweb/internal/model/job"
)

type OesCronConf struct {
	ScriptName    string `yaml:"script_name"`
	ScriptLabel   string `yaml:"script_label"`
	Specification string `yaml:"specification"`
}

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
		if err := enc.AddObject("mon", stk.Mon); err != nil {
			return err
		}
	}
	if stk.CounterFetch != nil {
		if err := enc.AddObject("counter_fetch", stk.CounterFetch); err != nil {
			return err
		}
	}
	if stk.CounterDistribute != nil {
		if err := enc.AddObject("counter_distribute", stk.CounterDistribute); err != nil {
			return err
		}
	}
	if stk.Bse != nil {
		if err := enc.AddObject("bse", stk.Bse); err != nil {
			return err
		}
	}
	if stk.Sse != nil {
		if err := enc.AddObject("sse", stk.Sse); err != nil {
			return err
		}
	}
	if stk.Szse != nil {
		if err := enc.AddObject("szse", stk.Szse); err != nil {
			return err
		}
	}
	if stk.Csdc != nil {
		if err := enc.AddObject("csdc", stk.Csdc); err != nil {
			return err
		}
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
		if err := enc.AddObject("mon", crd.Mon); err != nil {
			return err
		}
	}
	if crd.CounterFetch != nil {
		if err := enc.AddObject("counter_fetch", crd.CounterFetch); err != nil {
			return err
		}
	}
	if crd.CounterDistribute != nil {
		if err := enc.AddObject("counter_distribute", crd.CounterDistribute); err != nil {
			return err
		}
	}
	if crd.Sse != nil {
		if err := enc.AddObject("sse", crd.Sse); err != nil {
			return err
		}
	}
	if crd.Szse != nil {
		if err := enc.AddObject("szse", crd.Szse); err != nil {
			return err
		}
	}
	if crd.Csdc != nil {
		if err := enc.AddObject("csdc", crd.Csdc); err != nil {
			return err
		}
	}
	if crd.SseLate != nil {
		if err := enc.AddObject("sse_late", crd.SseLate); err != nil {
			return err
		}
	}
	if crd.SzseLate != nil {
		if err := enc.AddObject("szse_late", crd.SzseLate); err != nil {
			return err
		}
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
		if err := enc.AddObject("mon", opt.Mon); err != nil {
			return err
		}
	}
	if opt.CounterFetch != nil {
		if err := enc.AddObject("counter_fetch", opt.CounterFetch); err != nil {
			return err
		}
	}
	if opt.CounterDistribute != nil {
		if err := enc.AddObject("counter_distribute", opt.CounterDistribute); err != nil {
			return err
		}
	}
	if opt.Sse != nil {
		if err := enc.AddObject("sse", opt.Sse); err != nil {
			return err
		}
	}
	if opt.Szse != nil {
		if err := enc.AddObject("szse", opt.Szse); err != nil {
			return err
		}
	}
	return nil
}
