package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	jobsmodel "gin-artweb/internal/model/jobs"
	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
)

type StkTaskRecordCache struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Bse               uint32
	Sse               uint32
	Szse              uint32
	Csde              uint32
}

func (mc StkTaskRecordCache) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "bse", "sse", "szse", "csde"}
}

func (mc StkTaskRecordCache) GetRecordIDs() []uint32 {
	return []uint32{mc.Mon, mc.CounterFetch, mc.CounterDistribute, mc.Bse, mc.Sse, mc.Szse, mc.Csde}
}

func (mc StkTaskRecordCache) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for i, task := range mc.GetTaskList() {
		enc.AddUint32(task, mc.GetRecordIDs()[i])
	}
	return nil
}

type StkTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *jobsmodel.ScriptRecordModel
	CounterFetch      *jobsmodel.ScriptRecordModel
	CounterDistribute *jobsmodel.ScriptRecordModel
	Bse               *jobsmodel.ScriptRecordModel
	Sse               *jobsmodel.ScriptRecordModel
	Szse              *jobsmodel.ScriptRecordModel
	Csdc              *jobsmodel.ScriptRecordModel
}

type StkTaskExecutionInfoUsecase struct {
	log      *zap.Logger
	ucRecord *JobsService
}

func NewStkTaskExecutionInfoUsecase(
	log *zap.Logger,
	ucRecord *JobsService,
) *StkTaskExecutionInfoUsecase {
	return &StkTaskExecutionInfoUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

func (uc *StkTaskExecutionInfoUsecase) BuildTaskExecutionInfos(
	ctx context.Context,
	ms []oesmodel.OesColonyModel,
) (*[]StkTaskExecutionInfo, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	// 过滤出STK类型的模型
	var stkModels []oesmodel.OesColonyModel
	for _, m := range ms {
		if m.SystemType == "STK" {
			stkModels = append(stkModels, m)
		}
	}

	// 获取集群的执行记录,统计执行记录id
	trs := make([]StkTaskRecordCache, len(stkModels))
	for i, m := range stkModels {
		tr, err := uc.LoadStkTaskRecordCacheFromFiles(ctx, m.ColonyNum)
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
	tasks := make([]StkTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		info, err := uc.BuildTaskExecutionInfo(ctx, tr, cache)
		if err != nil {
			return nil, err
		}
		tasks[i] = info
	}
	return &tasks, nil
}

func (uc *StkTaskExecutionInfoUsecase) BuildTaskExecutionInfo(
	ctx context.Context,
	tr StkTaskRecordCache,
	cache map[uint32]jobsmodel.ScriptRecordModel,
) (StkTaskExecutionInfo, *errors.Error) {
	if ctx.Err() != nil {
		return StkTaskExecutionInfo{}, errors.FromError(ctx.Err())
	}
	return StkTaskExecutionInfo{
		ColonyNum:         tr.ColonyNum,
		Mon:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Mon, tr.ColonyNum, "mon"),
		CounterFetch:      uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterFetch, tr.ColonyNum, "counter_fetch"),
		CounterDistribute: uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterDistribute, tr.ColonyNum, "counter_distribute"),
		Bse:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Bse, tr.ColonyNum, "bse"),
		Sse:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Sse, tr.ColonyNum, "sse"),
		Szse:              uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Szse, tr.ColonyNum, "szse"),
		Csdc:              uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Csde, tr.ColonyNum, "csde"),
	}, nil
}

func (uc *StkTaskExecutionInfoUsecase) ExtractValidRecordIDsFromCaches(
	ctx context.Context,
	trs []StkTaskRecordCache,
) ([]uint32, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
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
		if tr.Bse != 0 {
			recordIDs = append(recordIDs, tr.Bse)
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
	}
	return recordIDs, nil
}

func (uc *StkTaskExecutionInfoUsecase) LoadStkTaskRecordCacheFromFiles(
	ctx context.Context,
	colonyNum string,
) (*StkTaskRecordCache, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	var (
		getTaskIDErr      error
		mon               uint32
		counterFetch      uint32
		counterDistribute uint32
		bse               uint32
		sse               uint32
		szse              uint32
		csde              uint32
	)
	mon, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".mon"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	counterFetch, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".counter_fetch"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	counterDistribute, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".counter_distribute"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	bse, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".bse"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	sse, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".sse"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	szse, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".szse"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	csde, getTaskIDErr = common.ReadUint32FromFile(filepath.Join(flagDir, ".csde"))
	if getTaskIDErr != nil {
		return nil, errors.FromError(getTaskIDErr)
	}
	mc := &StkTaskRecordCache{
		ColonyNum:         colonyNum,
		Mon:               mon,
		CounterFetch:      counterFetch,
		CounterDistribute: counterDistribute,
		Bse:               bse,
		Sse:               sse,
		Szse:              szse,
		Csde:              csde,
	}
	uc.log.Debug(
		"查询oes现货任务状态对应的执行记录id成功",
		zap.Object("stk_task_record", mc),
	)
	return mc, nil
}
