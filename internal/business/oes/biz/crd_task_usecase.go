package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizJobs "gin-artweb/internal/infra/jobs/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type CrdTaskRecordCache struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Sse               uint32
	Szse              uint32
	Csde              uint32
	SseLate           uint32
	SzseLate          uint32
}

func (mc CrdTaskRecordCache) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "sse", "szse", "csde", "szse_late", "sse_late"}
}

func (mc CrdTaskRecordCache) GetRecordIDs() []uint32 {
	return []uint32{mc.Mon, mc.CounterFetch, mc.CounterDistribute, mc.Sse, mc.Szse, mc.Csde, mc.SseLate, mc.SzseLate}
}

func (mc CrdTaskRecordCache) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for i, task := range mc.GetTaskList() {
		enc.AddUint32(task, mc.GetRecordIDs()[i])
	}
	return nil
}

type CrdTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *bizJobs.ScriptRecordModel
	CounterFetch      *bizJobs.ScriptRecordModel
	CounterDistribute *bizJobs.ScriptRecordModel
	Sse               *bizJobs.ScriptRecordModel
	Szse              *bizJobs.ScriptRecordModel
	Csdc              *bizJobs.ScriptRecordModel
	SseLate           *bizJobs.ScriptRecordModel
	SzseLate          *bizJobs.ScriptRecordModel
}

type CrdTaskExecutionInfoUsecase struct {
	log      *zap.Logger
	ucRecord *RecordUsecase
}

func NewCrdTaskExecutionInfoUsecase(
	log *zap.Logger,
	ucRecord *RecordUsecase,
) *CrdTaskExecutionInfoUsecase {
	return &CrdTaskExecutionInfoUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

func (uc *CrdTaskExecutionInfoUsecase) BuildTaskExecutionInfos(
	ctx context.Context,
	ms []OesColonyModel,
) (*[]CrdTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	// 过滤出两融类型的oes集群
	var crdModels []OesColonyModel
	for _, m := range ms {
		if m.SystemType == "CRD" {
			crdModels = append(crdModels, m)
		}
	}

	// 获取集群的执行记录,统计执行记录id
	trs := make([]CrdTaskRecordCache, len(crdModels))
	for i, m := range crdModels {
		tr, err := uc.LoadCrdTaskRecordCacheFromFiles(ctx, m.ColonyNum)
		if err != nil {
			return nil, errors.FromError(err)
		}
		if tr != nil {
			trs[i] = *tr
		}
	}
	recoids, rErr := uc.ExtractValidRecordIDsFromCaches(ctx, trs)
	if rErr != nil {
		return nil, rErr
	}

	// 执行数据库查询，获取集群对应的执行记录
	cache, rErr := uc.ucRecord.FindRecordsByIDs(ctx, recoids)
	if rErr != nil {
		return nil, rErr
	}
	tasks := make([]CrdTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		info, err := uc.BuildTaskExecutionInfo(ctx, tr, cache)
		if err != nil {
			return nil, err
		}
		tasks[i] = info
	}
	return &tasks, nil
}

func (uc *CrdTaskExecutionInfoUsecase) BuildTaskExecutionInfo(
	ctx context.Context,
	tr CrdTaskRecordCache,
	cache map[uint32]bizJobs.ScriptRecordModel,
) (CrdTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return CrdTaskExecutionInfo{}, errors.FromError(err)
	}
	return CrdTaskExecutionInfo{
		ColonyNum:         tr.ColonyNum,
		Mon:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Mon, tr.ColonyNum, "mon"),
		CounterFetch:      uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterFetch, tr.ColonyNum, "counter_fetch"),
		CounterDistribute: uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterDistribute, tr.ColonyNum, "counter_distribute"),
		Sse:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Sse, tr.ColonyNum, "sse"),
		Szse:              uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Szse, tr.ColonyNum, "szse"),
		Csdc:              uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Csde, tr.ColonyNum, "csde"),
		SseLate:           uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Sse, tr.ColonyNum, "sse_late"),
		SzseLate:          uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Szse, tr.ColonyNum, "szse_late"),
	}, nil
}

func (uc *CrdTaskExecutionInfoUsecase) ExtractValidRecordIDsFromCaches(
	ctx context.Context,
	trs []CrdTaskRecordCache,
) ([]uint32, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	var recordIDs []uint32
	for _, tr := range trs {
		if tr.Mon != 0 {
			recordIDs = append(recordIDs, tr.Mon)
		}
		if tr.CounterFetch != 0 {
			recordIDs = append(recordIDs, tr.CounterFetch)
		}
		if tr.CounterDistribute != 0 {
			recordIDs = append(recordIDs, tr.CounterDistribute)
		}
		if tr.Sse != 0 {
			recordIDs = append(recordIDs, tr.Sse)
		}
		if tr.Szse != 0 {
			recordIDs = append(recordIDs, tr.Szse)
		}
		if tr.Csde != 0 {
			recordIDs = append(recordIDs, tr.Csde)
		}
		if tr.SseLate != 0 {
			recordIDs = append(recordIDs, tr.SseLate)
		}
		if tr.SzseLate != 0 {
			recordIDs = append(recordIDs, tr.SzseLate)
		}
	}
	return recordIDs, nil
}

func (uc *CrdTaskExecutionInfoUsecase) LoadCrdTaskRecordCacheFromFiles(
	ctx context.Context,
	colonyNum string,
) (*CrdTaskRecordCache, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	mc := CrdTaskRecordCache{
		ColonyNum:         colonyNum,
		Mon:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".mon")),
		CounterFetch:      uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_fetch")),
		CounterDistribute: uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_distribute")),
		Sse:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".sse")),
		Szse:              uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".szse")),
		Csde:              uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".csde")),
		SseLate:           uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".sse_late")),
		SzseLate:          uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".szse_late")),
	}
	uc.log.Debug(
		"查询oes两融任务状态对应的执行记录id成功",
		zap.Object("crd_task_record", mc),
	)
	return &mc, nil
}
