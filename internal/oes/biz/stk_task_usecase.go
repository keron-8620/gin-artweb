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

type StkTaskExecutionInfo struct {
	ColonyNum         string
	Mon               *bizJobs.ScriptRecordModel
	CounterFetch      *bizJobs.ScriptRecordModel
	CounterDistribute *bizJobs.ScriptRecordModel
	Bse               *bizJobs.ScriptRecordModel
	Sse               *bizJobs.ScriptRecordModel
	Szse              *bizJobs.ScriptRecordModel
	Csdc              *bizJobs.ScriptRecordModel
}

type StkTaskExecutionInfoUsecase struct {
	log      *zap.Logger
	ucRecord *RecordUsecase
}

func NewStkTaskExecutionInfoUsecase(
	log *zap.Logger,
	ucRecord *RecordUsecase,
) *StkTaskExecutionInfoUsecase {
	return &StkTaskExecutionInfoUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

func (uc *StkTaskExecutionInfoUsecase) BuildTaskExecutionInfos(
	ctx context.Context,
	ms []OesColonyModel,
) (*[]StkTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	trs := make([]StkTaskRecordCache, len(ms))
	for i, m := range ms {
		if m.SystemType != "STK" {
			continue
		}
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
	cache map[uint32]bizJobs.ScriptRecordModel,
) (StkTaskExecutionInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return StkTaskExecutionInfo{}, errors.FromError(err)
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
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagDir := filepath.Join(config.StorageDir, "oes", "flags", colonyNum)
	mc := StkTaskRecordCache{
		ColonyNum:         colonyNum,
		Mon:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".mon")),
		CounterFetch:      uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_fetch")),
		CounterDistribute: uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".counter_distribute")),
		Bse:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".bse")),
		Sse:               uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".sse")),
		Szse:              uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".szse")),
		Csde:              uc.ucRecord.ReadUint32FromFile(filepath.Join(flagDir, ".csde")),
	}
	return &mc, nil
}
