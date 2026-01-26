package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"

	bizJobs "gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type OptTaskRecordCache struct {
	ColonyNum         string
	Mon               uint32
	CounterFetch      uint32
	CounterDistribute uint32
	Sse               uint32
	Szse              uint32
}

func (mc OptTaskRecordCache) GetTaskList() []string {
	return []string{"mon", "counter_fetch", "counter_distribute", "sse", "szse"}
}

func (mc OptTaskRecordCache) GetRecordIDs() []uint32 {
	return []uint32{mc.Mon, mc.CounterFetch, mc.CounterDistribute, mc.Sse, mc.Szse}
}

type OptTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *bizJobs.ScriptRecordModel
	CounterFetch      *bizJobs.ScriptRecordModel
	CounterDistribute *bizJobs.ScriptRecordModel
	Sse               *bizJobs.ScriptRecordModel
	Szse              *bizJobs.ScriptRecordModel
}

type OptTaskExecutionInfoUsecase struct {
	log      *zap.Logger
	ucRecord *RecordUsecase
}

func NewOptTaskExecutionInfoUsecase(
	log *zap.Logger,
	ucRecord *RecordUsecase,
) *OptTaskExecutionInfoUsecase {
	return &OptTaskExecutionInfoUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

func (uc *OptTaskExecutionInfoUsecase) BuildTaskExecutionInfos(
	ctx context.Context,
	ms []OesColonyModel,
) (*[]OptTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	trs := make([]OptTaskRecordCache, len(ms))
	for i, m := range ms {
		if m.SystemType != "opt" {
			continue
		}
		tr, err := uc.LoadOptTaskRecordCacheFromFiles(ctx, m.ColonyNum)
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
	cache, rErr := uc.ucRecord.FindRecordsByIDs(ctx, recoids)
	if rErr != nil {
		return nil, rErr
	}
	tasks := make([]OptTaskExecutionInfo, len(trs))
	for i, tr := range trs {
		info, err := uc.BuildTaskExecutionInfo(ctx, tr, cache)
		if err != nil {
			return nil, err
		}
		tasks[i] = info
	}
	return &tasks, nil
}

func (uc *OptTaskExecutionInfoUsecase) BuildTaskExecutionInfo(
	ctx context.Context,
	tr OptTaskRecordCache,
	cache map[uint32]bizJobs.ScriptRecordModel,
) (OptTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return OptTaskExecutionInfo{}, errors.FromError(err)
	}
	return OptTaskExecutionInfo{
		ColonyNum:         tr.ColonyNum,
		Mon:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Mon, tr.ColonyNum, "mon"),
		CounterFetch:      uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterFetch, tr.ColonyNum, "counter_fetch"),
		CounterDistribute: uc.ucRecord.FindRecordsByMap(ctx, cache, tr.CounterDistribute, tr.ColonyNum, "counter_distribute"),
		Sse:               uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Sse, tr.ColonyNum, "sse"),
		Szse:              uc.ucRecord.FindRecordsByMap(ctx, cache, tr.Szse, tr.ColonyNum, "szse"),
	}, nil
}

func (uc *OptTaskExecutionInfoUsecase) ExtractValidRecordIDsFromCaches(
	ctx context.Context,
	trs []OptTaskRecordCache,
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
	}
	return recordIDs, nil
}

func (uc *OptTaskExecutionInfoUsecase) LoadOptTaskRecordCacheFromFiles(
	ctx context.Context,
	colonyNum string,
) (*OptTaskRecordCache, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	mc := OptTaskRecordCache{
		ColonyNum:         colonyNum,
		Mon:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".mon")),
		CounterFetch:      uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_fetch")),
		CounterDistribute: uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_distribute")),
		Sse:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".sse")),
		Szse:              uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".szse")),
	}
	return &mc, nil
}
